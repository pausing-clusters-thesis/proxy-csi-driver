package pod

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/helpers/slices"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/resourceapply"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volume"
	soresourceapply "github.com/scylladb/scylla-operator/pkg/resourceapply"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	storagev1listers "k8s.io/client-go/listers/storage/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

func (pc *Controller) syncVolume(ctx context.Context, spec *volume.Spec) error {
	now := time.Now()
	klog.V(4).InfoS("Started syncing Volume", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim), "startTime", now)
	defer func() {
		klog.V(4).InfoS("Finished syncing Volume", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim), "duration", time.Since(now))
	}()

	if spec.ProxyPersistentVolume.Spec.CSI == nil {
		klog.ErrorS(nil, "Proxy PersistentVolumeSource unsupported, only CSIPersistentVolumeSource is supported", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim))
		return nil
	}

	if spec.ProxyPersistentVolume.Spec.CSI.Driver != naming.DriverName {
		klog.ErrorS(nil, "Proxy PersistentVolumeSource is not provided by us", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim))
		return nil
	}

	if spec.BackendPersistentVolume.Spec.CSI == nil {
		klog.ErrorS(nil, "Backend PersistentVolumeSource unsupported, only CSIPersistentVolumeSource is supported", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim))
		return nil
	}

	proxyVolumeAttachmentName := getAttachmentName(spec.ProxyPersistentVolume.Spec.CSI.VolumeHandle, spec.ProxyPersistentVolume.Spec.CSI.Driver, spec.Pod.Spec.NodeName)
	proxyVolumeAttachment, err := pc.volumeAttachmentLister.Get(proxyVolumeAttachmentName)
	if err != nil {
		return fmt.Errorf("can't get VolumeAttachment %q for proxy PersistentVolume %q: %w", proxyVolumeAttachmentName, naming.ObjRef(spec.ProxyPersistentVolume), err)
	}

	// TODO: we should have an abstraction for all csi/driver-registry/csi-driver related functions

	csiDriver, err := pc.csiDriverLister.Get(spec.BackendPersistentVolume.Spec.CSI.Driver)
	if err != nil {
		return fmt.Errorf("can't get CSIDriver %q: %w", spec.BackendPersistentVolume.Spec.CSI.Driver, err)
	}

	// TODO: move to CanAttach func of csi attacher/plugin? or do not return an attacher in that case
	if csiDriver.Spec.AttachRequired == nil || !*csiDriver.Spec.AttachRequired {
		klog.V(4).InfoS("Volume does not require attachment.", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim))
		return nil
	}

	// TODO: move to utils?
	if !isMultiAttachAllowed(spec.BackendPersistentVolume) {
		volumeAttachedNodes, err := getVolumeAttachedNodes(pc.volumeAttachmentLister, spec.BackendPersistentVolume.Spec.CSI)
		if err != nil {
			return fmt.Errorf("can't get volume attached nodes: %w", err)
		}

		otherVolumeAttachedNodes := slices.Filter(volumeAttachedNodes, func(nodeName string) bool {
			return nodeName != spec.Pod.Spec.NodeName
		})

		if len(otherVolumeAttachedNodes) > 0 {
			klog.V(2).InfoS("Volume is already exclusively attached and can't be attached to another node.", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim), "otherVolumeAttachedNodes", otherVolumeAttachedNodes)
			// TODO: event?
			return nil
		}
	}

	// This corresponds to GenerateAttachVolumeFunc
	// 1. get attachable plugin by spec
	// 2. get attacher from plugin
	// 3. call attacher's attach method
	// 4. generate events etc

	// TODO: extract this to csi attacher
	va := makeVolumeAttachment(spec.BackendPersistentVolume, spec.Pod.Spec.NodeName, proxyVolumeAttachment)
	va, changed, err := resourceapply.ApplyVolumeAttachment(ctx, pc.kubeClient.StorageV1(), pc.volumeAttachmentLister, pc.eventRecorder, va, soresourceapply.ApplyOptions{AllowMissingControllerRef: true})
	if err != nil {
		return fmt.Errorf("can't apply volume attachment: %w", err)
	}
	if changed {
		klog.V(2).InfoS("Applied VolumeAttachment", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim), "VolumeAttachment", klog.KObj(va))
	}

	return nil
}

// returns true only if it's DEFINITELY not allowed, otherwise inconclusive.
func isMultiAttachAllowed(pv *corev1.PersistentVolume) bool {
	if len(pv.Spec.AccessModes) == 0 {
		return true
	}

	for _, accessMode := range pv.Spec.AccessModes {
		if accessMode == corev1.ReadWriteMany || accessMode == corev1.ReadOnlyMany {
			return true
		}
	}

	return false
}

func getVolumeAttachedNodes(volumeAttachmentLister storagev1listers.VolumeAttachmentLister, pvSource *corev1.CSIPersistentVolumeSource) ([]string, error) {
	res := make([]string, 0)

	vaList, err := volumeAttachmentLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("can't list VolumeAttachments: %w", err)
	}

	for _, va := range vaList {
		if va.Spec.Attacher == pvSource.Driver && va.Spec.Source.PersistentVolumeName != nil && *va.Spec.Source.PersistentVolumeName == pvSource.VolumeHandle {
			res = append(res, va.Spec.NodeName)
		}
	}

	return res, nil
}

func makeVolumeAttachment(pv *corev1.PersistentVolume, nodeName string, owner *storagev1.VolumeAttachment) *storagev1.VolumeAttachment {
	ownerGVK := storagev1.SchemeGroupVersion.WithKind("VolumeAttachment")

	return &storagev1.VolumeAttachment{
		ObjectMeta: metav1.ObjectMeta{
			Name: getAttachmentName(pv.Spec.CSI.VolumeHandle, pv.Spec.CSI.Driver, nodeName),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: ownerGVK.GroupVersion().String(),
					Kind:       ownerGVK.Kind,
					Name:       owner.GetName(),
					UID:        owner.GetUID(),
				},
			},
			Annotations: map[string]string{
				naming.DelayedStorageProxyVolumeAttachmentRefAnnotation: owner.GetName(),
			},
		},
		Spec: storagev1.VolumeAttachmentSpec{
			Attacher: pv.Spec.CSI.Driver,
			NodeName: nodeName,
			Source: storagev1.VolumeAttachmentSource{
				PersistentVolumeName: ptr.To(pv.GetName()),
			},
		},
	}
}

// TODO: move to util
// TODO: add ref to document defining this
// getAttachmentName returns csi-<sha256(volName,csiDriverName,NodeName)>
func getAttachmentName(volumeHandle string, driverName string, nodeName string) string {
	result := sha256.Sum256([]byte(fmt.Sprintf("%s%s%s", volumeHandle, driverName, nodeName)))
	return fmt.Sprintf("csi-%x", result)
}
