package protection

import (
	"context"
	"fmt"
	"time"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/controllerhelpers"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (bpvcpc *Controller) sync(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.ErrorS(err, "Failed to split meta namespace cache key", "cacheKey", key)
		return fmt.Errorf("can't split meta namespace key: %w", err)
	}

	now := time.Now()
	klog.V(4).InfoS("Started syncing PersistentVolumeClaim", "PersistentVolumeClaim", klog.KRef(namespace, name), "startTime", now)
	defer func() {
		klog.V(4).InfoS("Finished syncing PersistentVolumeClaim", "PersistentVolumeClaim", klog.KRef(namespace, name), "duration", time.Since(now))
	}()

	pvc, err := bpvcpc.persistentVolumeClaimLister.PersistentVolumeClaims(namespace).Get(name)
	if apierrors.IsNotFound(err) {
		klog.V(2).InfoS("PersistentVolumeClaim has been deleted", "PersistentVolumeClaim", klog.KRef(namespace, name))
		return nil
	}
	if err != nil {
		return fmt.Errorf("can't get PersistentVolumeClaim: %w", err)
	}

	if pvc.DeletionTimestamp != nil && controllerhelpers.HasFinalizer(pvc, naming.DelayedStorageBackendPersistentVolumeClaimProtectionFinalizer) {
		if isBackendPersistentVolumeClaim(pvc) {
			klog.V(4).InfoS("Keeping PersistentVolumeClaim's finalizer because it is backing another PersistentVolumeClaim", "PersistentVolumeClaim", klog.KObj(pvc))
			return nil
		}

		err = bpvcpc.removeFinalizer(ctx, pvc)
		if err != nil {
			return fmt.Errorf("can't remove finalizer: %w", err)
		}

		return nil
	}

	if pvc.DeletionTimestamp == nil && !controllerhelpers.HasFinalizer(pvc, naming.DelayedStorageBackendPersistentVolumeClaimProtectionFinalizer) {
		err = bpvcpc.addFinalizer(ctx, pvc)
		if err != nil {
			return fmt.Errorf("can't add finalizer: %w", err)
		}

		return nil
	}

	return nil
}

func isBackendPersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) bool {
	return controllerhelpers.HasAnnotation(pvc, naming.DelayedStorageProxyPersistentVolumeClaimRefAnnotation)
}

func (bpvcpc *Controller) addFinalizer(ctx context.Context, pvc *corev1.PersistentVolumeClaim) error {
	klog.V(4).InfoS("Adding finalizer to PersistentVolumeClaim", "PersistentVolumeClaim", klog.KObj(pvc))

	patch, err := controllerhelpers.PrepareAddFinalizerPatch(pvc, naming.DelayedStorageBackendPersistentVolumeClaimProtectionFinalizer)
	if err != nil {
		return fmt.Errorf("can't prepare add finalizer patch: %w", err)
	}

	_, err = bpvcpc.kubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace).Patch(ctx, pvc.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("can't patch PersistentVolumeClaim %q: %w", naming.ObjRef(pvc), err)
	}

	klog.V(2).InfoS("Added finalizer to PersistentVolumeClaim", "PersistentVolumeClaim", klog.KObj(pvc))

	return nil
}

func (bpvcpc *Controller) removeFinalizer(ctx context.Context, pvc *corev1.PersistentVolumeClaim) error {
	klog.V(4).InfoS("Removing PersistentVolumeClaim's finalizer", "PersistentVolumeClaim", klog.KObj(pvc))

	patch, err := controllerhelpers.PrepareRemoveFinalizerPatch(pvc, naming.DelayedStorageBackendPersistentVolumeClaimProtectionFinalizer)
	if err != nil {
		return fmt.Errorf("can't prepare remove finalizer patch: %w", err)
	}

	_, err = bpvcpc.kubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace).Patch(ctx, pvc.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("can't patch PersistentVolumeClaim %q: %w", naming.ObjRef(pvc), err)
	}

	klog.V(2).InfoS("Removed PersistentVolumeClaim's finalizer", "PersistentVolumeClaim", klog.KObj(pvc))

	return nil

}
