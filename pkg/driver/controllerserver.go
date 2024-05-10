package driver

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/helpers/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	storagev1listers "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

type ControllerServer struct {
	csi.UnimplementedControllerServer
	IdentityServer

	kubeClient kubernetes.Interface

	csiDriverLister        storagev1listers.CSIDriverLister
	volumeAttachmentLister storagev1listers.VolumeAttachmentLister

	podLister                   corev1listers.PodLister
	persistentVolumeClaimLister corev1listers.PersistentVolumeClaimLister
	persistentVolumeLister      corev1listers.PersistentVolumeLister

	cachesToSync []cache.InformerSynced

	eventRecorder record.EventRecorder
}

func NewControllerServer(version string) *ControllerServer {
	return &ControllerServer{
		IdentityServer: IdentityServer{
			version: version,
		},
	}
}

var _ csi.ControllerServer = &ControllerServer{}

// TODO: verify if any of these attributes need to be checked, where they come from etc
func (cs *ControllerServer) CreateVolume(ctx context.Context, request *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "CreateVolume", "request", protosanitizer.StripSecrets(request))
	var err error

	volumeID := request.GetName()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "name must be set")
	}

	volumeCapabilities := request.GetVolumeCapabilities()
	if volumeCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapabilities must be set")
	}

	err = validateVolumeCapabilities(volumeCapabilities)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Unsupported volume capabilities: %v", err))
	}

	parameters := request.GetParameters()
	err = validateVolumeParameters(parameters)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Unsupported volume parameters: %v", err))
	}

	if len(request.GetMutableParameters()) > 0 {
		return nil, status.Error(codes.InvalidArgument, "Mutable parameters cannot be set")
	}

	var accessTypeMount, accessTypeBlock bool
	var requestedFilesystemType string
	for _, vc := range volumeCapabilities {
		if vc.GetBlock() != nil {
			accessTypeBlock = true
		}
		if vc.GetMount() != nil {
			accessTypeMount = true
			requestedFilesystemType = vc.GetMount().GetFsType()
		}
	}

	if accessTypeBlock && accessTypeMount {
		return nil, status.Error(codes.InvalidArgument, "Only one access type must be specified")
	}

	// Default to mount.
	requestedAccessType := AccessTypeMount
	if accessTypeBlock {
		requestedAccessType = AccessTypeBlock
	}

	if !slices.Contains(supportedAccessTypes, requestedAccessType) {
		return nil, status.Error(codes.InvalidArgument, "Unsupported access type")
	}

	if accessTypeMount && !slices.Contains(supportedFilesystemTypes, requestedFilesystemType) {
		return nil, status.Error(codes.InvalidArgument, "Unsupported filesystem type")
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: 0,
			// TODO: not sure if this is necessary
			AccessibleTopology: nil,
		},
	}, nil
}

func (cs *ControllerServer) DeleteVolume(ctx context.Context, request *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "DeleteVolume", "request", protosanitizer.StripSecrets(request))

	// No-op.

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, request *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "ControllerPublishVolume", "request", protosanitizer.StripSecrets(request))

	// No-op.

	return &csi.ControllerPublishVolumeResponse{}, nil
}

func (cs *ControllerServer) ControllerUnpublishVolume(ctx context.Context, request *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "ControllerUnpublishVolume", "request", protosanitizer.StripSecrets(request))

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (cs *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, request *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "ValidateVolumeCapabilities", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method ValidateVolumeCapabilities not implemented")
}

func (cs *ControllerServer) ListVolumes(ctx context.Context, request *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "ListVolumes", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method ListVolumes not implemented")
}

func (cs *ControllerServer) GetCapacity(ctx context.Context, request *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "GetCapacity", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method GetCapacity not implemented")
}

func (cs *ControllerServer) ControllerGetCapabilities(ctx context.Context, request *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "ControllerGetCapabilities", "request", protosanitizer.StripSecrets(request))

	controllerServiceCapabilityTypes := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		// Set RPC_PUBLISH_UNPUBLISH_VOLUME so that we can proxy ControllerUnpublishVolume calls by deleting VAs of backing PVCs.
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	}

	controllerServiceCapabilities := slices.Map(controllerServiceCapabilityTypes, func(t csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: t,
				},
			},
		}
	})

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: controllerServiceCapabilities,
	}, nil
}

func (cs *ControllerServer) CreateSnapshot(ctx context.Context, request *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "CreateSnapshot", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method CreateSnapshot not implemented")
}

func (cs *ControllerServer) DeleteSnapshot(ctx context.Context, request *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "DeleteSnapshot", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method DeleteSnapshot not implemented")
}

func (cs *ControllerServer) ListSnapshots(ctx context.Context, request *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "ListSnapshots", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method ListSnapshots not implemented")
}

func (cs *ControllerServer) ControllerExpandVolume(ctx context.Context, request *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "ControllerExpandVolume", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method ControllerExpandVolume not implemented")
}

func (cs *ControllerServer) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "ControllerGetVolume", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method ControllerGetVolume not implemented")
}

func (cs *ControllerServer) ControllerModifyVolume(ctx context.Context, request *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	klog.V(4).InfoS("Request received", "server", "controller", "function", "ControllerModifyVolume", "request", protosanitizer.StripSecrets(request))

	return nil, status.Errorf(codes.Unimplemented, "method ControllerModifyVolume not implemented")
}
