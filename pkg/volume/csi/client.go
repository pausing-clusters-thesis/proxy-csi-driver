package csi

import (
	"context"
	"fmt"
	"io"
	"strconv"

	csipbv1 "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/helpers/slices"
	corev1 "k8s.io/api/core/v1"
)

const (
	fsTypeBlockName = "block"
)

type DriverClient struct {
	endpoint string
}

// TODO: dynamically discover endpoint
func NewDriverClient(endpoint string) (*DriverClient, error) {
	return &DriverClient{
		endpoint: endpoint,
	}, nil
}

func (c *DriverClient) NodeUnstageVolume(
	ctx context.Context,
	volumeHandle string,
	stagingTargetPath string,
) error {
	nodeClient, closer, err := newNodeClent(ctx, c.endpoint)
	if err != nil {
		return fmt.Errorf("can't create node client: %w", err)
	}
	defer closer.Close()

	req := &csipbv1.NodeUnstageVolumeRequest{
		VolumeId:          volumeHandle,
		StagingTargetPath: stagingTargetPath,
	}

	_, err = nodeClient.NodeUnstageVolume(ctx, req)
	if err != nil {
		return fmt.Errorf("can't node unstage volume: %w", err)
	}

	return nil
}

func (c *DriverClient) NodeUnpublishVolume(
	ctx context.Context,
	volumeHandle string,
	targetPath string,
) error {
	nodeClient, closer, err := newNodeClent(ctx, c.endpoint)
	if err != nil {
		return fmt.Errorf("can't create node client: %w", err)
	}
	defer closer.Close()

	req := &csipbv1.NodeUnpublishVolumeRequest{
		VolumeId:   volumeHandle,
		TargetPath: targetPath,
	}

	_, err = nodeClient.NodeUnpublishVolume(ctx, req)
	if err != nil {
		return fmt.Errorf("can't node unpublish volume: %w", err)
	}

	return nil
}

func (c *DriverClient) NodePublishVolume(
	ctx context.Context,
	volumeHandle string,
	readOnly bool,
	stagingTargetPath string,
	targetPath string,
	accessMode corev1.PersistentVolumeAccessMode,
	publishContext map[string]string,
	volumeContext map[string]string,
	secrets map[string]string,
	fsType string,
	mountOptions []string,
	fsGroup *int64,
) error {
	nodeClient, closer, err := newNodeClent(ctx, c.endpoint)
	if err != nil {
		return fmt.Errorf("can't create node client: %w", err)
	}
	defer closer.Close()

	nodeAccessModeMapper, err := c.getNodeAccessModeMapper(ctx)
	if err != nil {
		return fmt.Errorf("can't get node access mode mapper: %w", err)
	}

	volumeCapability := &csipbv1.VolumeCapability{
		AccessType: nil,
		AccessMode: &csipbv1.VolumeCapability_AccessMode{
			Mode: nodeAccessModeMapper(accessMode),
		},
	}

	// TODO: move this to some common func?
	if fsType == fsTypeBlockName {
		volumeCapability.AccessType = &csipbv1.VolumeCapability_Block{
			Block: &csipbv1.VolumeCapability_BlockVolume{},
		}
	} else {
		mountVolume := &csipbv1.VolumeCapability_MountVolume{
			FsType:     fsType,
			MountFlags: mountOptions,
		}
		if fsGroup != nil {
			mountVolume.VolumeMountGroup = strconv.FormatInt(*fsGroup, 10 /* base */)
		}
		volumeCapability.AccessType = &csipbv1.VolumeCapability_Mount{
			Mount: mountVolume,
		}
	}

	req := &csipbv1.NodePublishVolumeRequest{
		VolumeId:          volumeHandle,
		PublishContext:    publishContext,
		StagingTargetPath: stagingTargetPath,
		TargetPath:        targetPath,
		VolumeCapability:  volumeCapability,
		Readonly:          readOnly,
		Secrets:           secrets,
		VolumeContext:     volumeContext,
	}

	_, err = nodeClient.NodePublishVolume(ctx, req)
	if err != nil {
		// TODO: check for final error and return uncertain progress error
		return fmt.Errorf("can't node publish volume: %w", err)
	}

	return nil
}

func (c *DriverClient) NodeSupportsStageUnstage(ctx context.Context) (bool, error) {
	return c.nodeSupportsCapability(ctx, csipbv1.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME)
}

func (c *DriverClient) NodeStageVolume(
	ctx context.Context,
	volumeHandle string,
	publishContext map[string]string,
	stagingTargetPath string,
	fsType string,
	accessMode corev1.PersistentVolumeAccessMode,
	secrets map[string]string,
	volumeContext map[string]string,
	mountOptions []string,
	fsGroup *int64,
) error {
	nodeClient, closer, err := newNodeClent(ctx, c.endpoint)
	if err != nil {
		return fmt.Errorf("can't create node client: %w", err)
	}
	defer closer.Close()

	nodeAccessModeMapper, err := c.getNodeAccessModeMapper(ctx)
	if err != nil {
		return fmt.Errorf("can't get node access mode mapper: %w", err)
	}

	volumeCapability := &csipbv1.VolumeCapability{
		AccessType: nil,
		AccessMode: &csipbv1.VolumeCapability_AccessMode{
			Mode: nodeAccessModeMapper(accessMode),
		},
	}

	if fsType == fsTypeBlockName {
		volumeCapability.AccessType = &csipbv1.VolumeCapability_Block{
			Block: &csipbv1.VolumeCapability_BlockVolume{},
		}
	} else {
		mountVolume := &csipbv1.VolumeCapability_MountVolume{
			FsType:     fsType,
			MountFlags: mountOptions,
		}
		if fsGroup != nil {
			mountVolume.VolumeMountGroup = strconv.FormatInt(*fsGroup, 10 /* base */)
		}
		volumeCapability.AccessType = &csipbv1.VolumeCapability_Mount{
			Mount: mountVolume,
		}
	}

	req := &csipbv1.NodeStageVolumeRequest{
		VolumeId:          volumeHandle,
		PublishContext:    publishContext,
		StagingTargetPath: stagingTargetPath,
		VolumeCapability:  volumeCapability,
		Secrets:           secrets,
		VolumeContext:     volumeContext,
	}

	_, err = nodeClient.NodeStageVolume(ctx, req)
	if err != nil {
		// TODO: check for final error and return uncertain progress error
		return fmt.Errorf("can't node stage volume: %w", err)
	}

	return nil
}

type nodeAccessModeMapper func(accessMode corev1.PersistentVolumeAccessMode) csipbv1.VolumeCapability_AccessMode_Mode

func (c *DriverClient) getNodeAccessModeMapper(ctx context.Context) (nodeAccessModeMapper, error) {
	isSingleNodeMultiWriterAccessModeSupported, err := c.NodeSupportsSingleNodeMultiWriterAccessMode(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't verify if node supports SingleNodeMultiWriter access mode: %w", err)
	}

	if isSingleNodeMultiWriterAccessModeSupported {
		return asSingleNodeMultiWriterCapableCSIAccessModeV1, nil
	}

	return asCSIAccessModeV1, nil
}

func asCSIAccessModeV1(accessMode corev1.PersistentVolumeAccessMode) csipbv1.VolumeCapability_AccessMode_Mode {
	switch accessMode {
	case corev1.ReadWriteOnce:
		return csipbv1.VolumeCapability_AccessMode_SINGLE_NODE_WRITER
	case corev1.ReadOnlyMany:
		return csipbv1.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY
	case corev1.ReadWriteMany:
		return csipbv1.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER
	case corev1.ReadWriteOncePod:
		return csipbv1.VolumeCapability_AccessMode_SINGLE_NODE_WRITER
	}
	return csipbv1.VolumeCapability_AccessMode_UNKNOWN
}

func asSingleNodeMultiWriterCapableCSIAccessModeV1(accessMode corev1.PersistentVolumeAccessMode) csipbv1.VolumeCapability_AccessMode_Mode {
	switch accessMode {
	case corev1.ReadWriteOnce:
		return csipbv1.VolumeCapability_AccessMode_SINGLE_NODE_MULTI_WRITER
	case corev1.ReadOnlyMany:
		return csipbv1.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY
	case corev1.ReadWriteMany:
		return csipbv1.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER
	case corev1.ReadWriteOncePod:
		return csipbv1.VolumeCapability_AccessMode_SINGLE_NODE_SINGLE_WRITER
	}
	return csipbv1.VolumeCapability_AccessMode_UNKNOWN
}

func (c *DriverClient) NodeSupportsVolumeMountGroup(ctx context.Context) (bool, error) {
	return c.nodeSupportsCapability(ctx, csipbv1.NodeServiceCapability_RPC_VOLUME_MOUNT_GROUP)
}

func (c *DriverClient) NodeSupportsSingleNodeMultiWriterAccessMode(ctx context.Context) (bool, error) {
	return c.nodeSupportsCapability(ctx, csipbv1.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER)
}

func (c *DriverClient) nodeSupportsCapability(ctx context.Context, capabilityType csipbv1.NodeServiceCapability_RPC_Type) (bool, error) {
	capabilities, err := c.nodeGetCapabilities(ctx)
	if err != nil {
		return false, fmt.Errorf("can't node get capabilities: %w", err)
	}

	containsCapabilityTypeFunc := func(c *csipbv1.NodeServiceCapability) bool {
		if c == nil || c.GetRpc() == nil {
			return false
		}

		return c.GetRpc().GetType() == capabilityType
	}

	return slices.ContainsFunc(capabilities, containsCapabilityTypeFunc), nil
}

func (c *DriverClient) nodeGetCapabilities(ctx context.Context) ([]*csipbv1.NodeServiceCapability, error) {
	nodeClient, closer, err := newNodeClent(ctx, c.endpoint)
	if err != nil {
		return []*csipbv1.NodeServiceCapability{}, fmt.Errorf("can't create node client: %w", err)
	}
	defer closer.Close()

	req := &csipbv1.NodeGetCapabilitiesRequest{}
	resp, err := nodeClient.NodeGetCapabilities(ctx, req)
	if err != nil {
		return []*csipbv1.NodeServiceCapability{}, fmt.Errorf("can't node get capabilities: %w", err)
	}

	capabilities := resp.GetCapabilities()
	return capabilities, nil
}

func newNodeClent(ctx context.Context, endpoint string) (csipbv1.NodeClient, io.Closer, error) {
	conn, err := connection.Connect(ctx, endpoint, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("can't create connection: %w", err)
	}

	nodeClient := csipbv1.NewNodeClient(conn)
	return nodeClient, conn, nil
}
