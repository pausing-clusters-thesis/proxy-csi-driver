package proxy

import (
	"context"
	"fmt"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (bpvcc *Controller) sync(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.ErrorS(err, "Failed to split meta namespace cache key", "cacheKey", key)
		return fmt.Errorf("can't split meta namespace key: %w", err)
	}

	backendPVC, err := bpvcc.persistentVolumeClaimLister.PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("can't get PersistentVolumeClaim: %w", err)
		}

		klog.V(2).InfoS("PersistentVolumeClaim has been deleted", "PersistentVolumeClaim", klog.KRef(namespace, name))
		return nil
	}

	proxyPVCName, hasProxyPVCNameAnnotation := backendPVC.Annotations[naming.DelayedStorageProxyPersistentVolumeClaimRefAnnotation]
	if !hasProxyPVCNameAnnotation {
		// Sanity check. It shouldn't have been queued.
		klog.ErrorS(nil, "PersistentVolumeClaim is missing an annotation, this should never happen", "BackendPersistentVolumeClaim", klog.KObj(backendPVC), "Annotation", naming.DelayedStorageProxyPersistentVolumeClaimRefAnnotation)
		return nil
	}

	// Backend PVC is bound to a claim.
	_, err = bpvcc.persistentVolumeClaimLister.PersistentVolumeClaims(namespace).Get(proxyPVCName)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("can't get proxy PersistentVolumeClaim %q: %w", naming.ManualRef(namespace, proxyPVCName), err)
		}

		_, err = bpvcc.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, proxyPVCName, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("can't get proxy PersistentVolumeClaim %q: %w", naming.ManualRef(namespace, proxyPVCName), err)
			}

			// Bound proxy PVC is not found. Release backend PVC to make it reclaimable.
			err = bpvcc.releaseBackendPVC(ctx, backendPVC)
			if err != nil {
				return fmt.Errorf("can't release backend PersistentVolumeClaim %q: %w", naming.ObjRef(backendPVC), err)
			}
		}
	}

	// TODO: there are cases to be handled here potentially
	// - proxyPVC is missing backendPVC request annotation (user removed it?)
	// - proxyPVC's backendPVC request annotation doesn't match this backendPVC (user changed it? Shouldn't be possible?)
	// - proxyPVC is missing bind completed annotation

	return nil
}

func (bpvcc *Controller) releaseBackendPVC(ctx context.Context, backendPVC *corev1.PersistentVolumeClaim) error {
	klog.V(5).InfoS("Releasing backend PersistentVolumeClaim", "BackendPersistentVolumeClaim", klog.KObj(backendPVC))
	var err error

	if !metav1.HasAnnotation(backendPVC.ObjectMeta, naming.DelayedStorageProxyPersistentVolumeClaimRefAnnotation) {
		klog.V(5).InfoS("Backend PersistentVolumeClaim has already been released", "BackendPersistentVolumeClaim", klog.KObj(backendPVC))
		return nil
	}

	backendPVCCopy := backendPVC.DeepCopy()
	delete(backendPVCCopy.Annotations, naming.DelayedStorageProxyPersistentVolumeClaimRefAnnotation)

	_, err = bpvcc.kubeClient.CoreV1().PersistentVolumeClaims(backendPVC.Namespace).Update(ctx, backendPVCCopy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("can't update PersistentVolumeClaim %q: %w", naming.ObjRef(backendPVC), err)
	}

	klog.V(4).InfoS("Released backend PersistentVolumeClaim", "BackendPersistentVolumeClaim", klog.KObj(backendPVC))
	return nil
}
