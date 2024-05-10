package driver

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/helpers/slices"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/stagingtargetpathbindstate"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/targetpathbind"
	volumebindstate "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volume/bindstate"
	proxycsi "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volume/csi"
	volumecsi "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volume/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

type NodeServer struct {
	csi.UnimplementedNodeServer
	IdentityServer

	nodeName string

	stagingTargetPathBindStateManager stagingtargetpathbindstate.StateManager
	targetPathBindManager             *targetpathbind.TargetPathBindManager
	volumeBindStateManager            volumebindstate.StateManager
	driverRegistry                    *proxycsi.DriverRegistry
}

func NewNodeServer(
	version string,
	nodeName string,
	stagingTargetPathBindStateManager stagingtargetpathbindstate.StateManager,
	targetPathManager *targetpathbind.TargetPathBindManager,
	volumeBindStateManager volumebindstate.StateManager,
	driverRegistry *proxycsi.DriverRegistry,
) *NodeServer {
	return &NodeServer{
		IdentityServer: IdentityServer{
			version: version,
		},

		nodeName: nodeName,

		stagingTargetPathBindStateManager: stagingTargetPathBindStateManager,
		targetPathBindManager:             targetPathManager,
		volumeBindStateManager:            volumeBindStateManager,

		driverRegistry: driverRegistry,
	}
}

func (ns *NodeServer) NodeStageVolume(ctx context.Context, request *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "node", "function", "NodeStageVolume", "request", protosanitizer.StripSecrets(request))
	var err error

	volumeID := request.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "VolumeID missing in request")
	}

	stagingTargetPath := request.GetStagingTargetPath()
	if len(stagingTargetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "StagingTargetPath missing in request")
	}

	klog.V(4).InfoS("NodeServer.NodeStageVolume: binding staging target path", "volumeID", volumeID)
	err = ns.stagingTargetPathBindStateManager.BindStagingTargetPath(volumeID, stagingTargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("can't bind staging target path: %v", err))
	}
	klog.V(4).InfoS("NodeServer.NodeStageVolume: staging target path bound", "volumeID", volumeID)

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *NodeServer) NodeUnstageVolume(ctx context.Context, request *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "node", "function", "NodeUnstageVolume", "request", protosanitizer.StripSecrets(request))

	volumeID := request.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "VolumeID missing in request")
	}

	stagingTargetPath := request.GetStagingTargetPath()
	if len(stagingTargetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "StagingTargetPath missing in request")
	}

	volumeBind, ok, err := ns.volumeBindStateManager.GetVolumeBind(volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Can't get volume bind with proxy volume handle: %v", err)
	}

	// TODO: move to device unmount
	if ok {
		// TODO: improve error messaging
		klog.V(4).InfoS("NodeServer.NodeUnstageVolume: unmounting backend device", "volumeID", volumeID, "volumeBind", volumeBind)
		csiAddress, err := ns.driverRegistry.GetAddress(volumeBind.BackendVolumeSpec.DriverName)
		if err != nil {
			// TODO: improve error handling
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can't get address of backend CSI driver: %v", err))
		}

		driverClient, err := volumecsi.NewDriverClient(csiAddress)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can't create driver client: %v", err))
		}

		// TODO: csi context
		stageUnstageSetInBackend, err := driverClient.NodeSupportsStageUnstage(ctx)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can't check whether backend driver supports STAGE_UNSTAGE_VOLUME: %v", err))
		}

		if !stageUnstageSetInBackend {
			klog.V(4).InfoS("NodeServer.NodeUnstageVolume: backend driver does not have STAGE_UNSTAGE_VOLUME capability, skipping unmount device", "volumeID", volumeID, "volumeBind", volumeBind)

			err = ns.volumeBindStateManager.DeleteVolumeBind(volumeID)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("can't delete volume bind: %v", err))
			}

			return &csi.NodeUnstageVolumeResponse{}, nil
		}

		err = driverClient.NodeUnstageVolume(ctx,
			volumeID,
			stagingTargetPath,
		)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can't node unstage volume: %v", err))
		}

		err = ns.volumeBindStateManager.DeleteVolumeBind(volumeID)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("can't delete volume bind: %v", err))
		}

		klog.V(4).InfoS("NodeServer.NodeUnstageVolume: backend device unmounted successfully", "volumeID", volumeID, "volumeBind", volumeBind)
	}

	klog.V(4).InfoS("NodeServer.NodeUnstageVolume: unbinding staging target path", "volumeID", volumeID)
	err = ns.stagingTargetPathBindStateManager.UnbindStagingTargetPath(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("can't unbind staging target path: %v", err))
	}
	klog.V(4).InfoS("NodeServer.NodeUnstageVolume: staging target path unbound", "volumeID", volumeID)

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *NodeServer) NodePublishVolume(ctx context.Context, request *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "node", "function", "NodePublishVolume", "request", protosanitizer.StripSecrets(request))
	var err error

	volumeID := request.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "VolumeID missing in request")
	}

	targetPath := request.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "TargetPath missing in request")
	}

	volumeCapability := request.GetVolumeCapability()
	if volumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapability missing in request")
	}

	// TODO: do we need this? This should be transparent for overriding volume.
	err = validateVolumeCapabilities([]*csi.VolumeCapability{volumeCapability})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Volume capability not supported: %s", err))
	}

	if volumeCapability.GetMount() == nil {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapability access type must be mount")
	}

	volumeContext := request.GetVolumeContext()
	if volumeContext == nil {
		return nil, status.Error(codes.InvalidArgument, "VolumeContext missing in request")
	}

	err = validateVolumeContext(volumeContext)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "VolumeContext invalid: %v", err)
	}

	ephemeral := volumeContext["csi.storage.k8s.io/ephemeral"]
	if ephemeral == "true" {
		return nil, status.Error(codes.InvalidArgument, "Ephemeral CSI volumes not supported")
	}

	podUID := volumeContext["csi.storage.k8s.io/pod.uid"]
	err = ns.targetPathBindManager.SetUpTargetPath(volumeID, podUID, targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Can't publish volume: %v", err)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *NodeServer) NodeUnpublishVolume(ctx context.Context, request *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "node", "function", "NodeUnpublishVolume", "request", protosanitizer.StripSecrets(request))
	var err error

	volumeID := request.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "VolumeID missing in request")
	}

	targetPath := request.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "TargetPath missing in request")
	}

	volumeBind, ok, err := ns.volumeBindStateManager.GetVolumeBind(volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Can't get volume bind with proxy volume handle: %v", err)
	}
	// TODO: move to unmounter
	if ok {
		klog.V(4).InfoS("NodeServer.NodeUnpublishVolume: unmounting backend volume", "volumeBind", volumeBind)
		csiAddress, err := ns.driverRegistry.GetAddress(volumeBind.BackendVolumeSpec.DriverName)
		if err != nil {
			// TODO: improve error handling
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can't get address of backend CSI driver: %v", err))
		}

		driverClient, err := volumecsi.NewDriverClient(csiAddress)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can't create driver client: %v", err))
		}

		// TODO: csi context
		err = driverClient.NodeUnpublishVolume(
			ctx,
			volumeBind.BackendVolumeSpec.VolumeHandle,
			targetPath,
		)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can't node unpublish volume: %v", err))
		}

		klog.V(4).InfoS("NodeServer.NodeUnpublishVolume: backend volume unmounted successfully", "volumeBind", volumeBind)
	}

	err = ns.targetPathBindManager.TearDownTargetPath(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Can't tear down target path: %v", err)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *NodeServer) NodeGetVolumeStats(ctx context.Context, request *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	klog.V(4).InfoS("Request received", "server", "node", "function", "NodeGetVolumeStats", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method NodeGetVolumeStats not implemented")
}

func (ns *NodeServer) NodeExpandVolume(ctx context.Context, request *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "node", "function", "NodeExpandVolume", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method NodeExpandVolume not implemented")
}

func (ns *NodeServer) NodeGetCapabilities(ctx context.Context, request *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	klog.V(4).InfoS("Request received", "server", "node", "function", "NodeGetCapabilities", "request", protosanitizer.StripSecrets(request))

	nodeServiceCapabilityTypes := []csi.NodeServiceCapability_RPC_Type{
		// Claim support for STAGE_UNSTAGE_VOLUME so that the driver can proxy NodeUnstageVolume calls to backend drivers.
		csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
		// Claim support for VOLUME_MOUNT_GROUP so that the CO does not attempt to perform ownership changes and delegates them to the driver.
		csi.NodeServiceCapability_RPC_VOLUME_MOUNT_GROUP,
	}

	nodeServiceCapabilities := slices.Map(nodeServiceCapabilityTypes, func(t csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
		return &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: t,
				},
			},
		}
	})

	return &csi.NodeGetCapabilitiesResponse{Capabilities: nodeServiceCapabilities}, nil
}

func (ns *NodeServer) NodeGetInfo(ctx context.Context, request *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.V(4).InfoS("Request received", "server", "node", "function", "NodeGetInfo", "request", protosanitizer.StripSecrets(request))

	return &csi.NodeGetInfoResponse{
		NodeId:             ns.nodeName,
		AccessibleTopology: nil,
	}, nil
}
