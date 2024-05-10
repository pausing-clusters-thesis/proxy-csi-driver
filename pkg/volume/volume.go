package volume

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Spec struct {
	Volume                       *corev1.Volume
	Pod                          *corev1.Pod
	ProxyPersistentVolumeClaim   *corev1.PersistentVolumeClaim
	ProxyPersistentVolume        *corev1.PersistentVolume
	BackendPersistentVolumeClaim *corev1.PersistentVolumeClaim
	BackendPersistentVolume      *corev1.PersistentVolume
}

type DeviceMounterArgs struct {
	FsGroup *int64
	//SELinuxLabel string
}

type MounterArgs struct {
	// When FsUser is set, the ownership of the volume will be modified to be
	// owned and writable by FsUser. Otherwise, there is no side effects.
	// Currently only supported with projected service account tokens.
	FsUser              *int64
	FsGroup             *int64
	FSGroupChangePolicy *corev1.PodFSGroupChangePolicy
	DesiredSize         *resource.Quantity
	//SELinuxLabel        string
}
