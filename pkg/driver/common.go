package driver

import (
	"fmt"
	"slices"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/apimachinery/pkg/util/errors"
)

type AccessType int

const (
	AccessTypeMount AccessType = iota
	AccessTypeBlock
)

var (
	volumeCapabilityAccessModes = []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_SINGLE_WRITER,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_MULTI_WRITER,
	}

	supportedFilesystemTypes = []string{
		"",
	}

	supportedAccessTypes = []AccessType{
		AccessTypeMount,
	}
)

func validateVolumeCapabilities(volumeCapabilities []*csi.VolumeCapability) error {
	var errs []error

	for _, volumeCapability := range volumeCapabilities {
		if !slices.Contains(volumeCapabilityAccessModes, volumeCapability.AccessMode.GetMode()) {
			errs = append(errs, fmt.Errorf("unsupported access mode %q", volumeCapability.AccessMode.GetMode().String()))
		}

		if volumeCapability.GetMount() == nil {
			errs = append(errs, fmt.Errorf("only filesystem volumes are supported"))
		}

		if volumeCapability.GetMount() != nil && !slices.Contains(supportedFilesystemTypes, volumeCapability.GetMount().FsType) {
			errs = append(errs, fmt.Errorf("unsupported filesystem type %q", volumeCapability.GetMount().FsType))
		}
	}

	return errors.NewAggregate(errs)
}

func validateVolumeParameters(parameters map[string]string) error {
	var errs []error
	for k := range parameters {
		switch k {
		default:
			errs = append(errs, fmt.Errorf("unsupported volume parameter key: %q", k))
		}
	}

	return errors.NewAggregate(errs)
}

func validateVolumeContext(volumeContext map[string]string) error {
	expectedVolumeAttributes := []string{
		"csi.storage.k8s.io/pod.name",
		"csi.storage.k8s.io/pod.namespace",
		"csi.storage.k8s.io/pod.uid",
		"csi.storage.k8s.io/serviceAccount.name",
		"csi.storage.k8s.io/ephemeral",
	}

	var errs []error
	for _, expectedVolumeAttribute := range expectedVolumeAttributes {
		v := volumeContext[expectedVolumeAttribute]
		if len(v) == 0 {
			errs = append(errs, fmt.Errorf("missing expected volume attribute %s", expectedVolumeAttribute))
		}
	}

	return errors.NewAggregate(errs)
}
