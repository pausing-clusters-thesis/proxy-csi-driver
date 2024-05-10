package proxy

import (
	"context"
	"fmt"
	"time"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/controllerhelpers"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/component-helpers/storage/volume"
	"k8s.io/klog/v2"
)

func (ppvcc *Controller) sync(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.ErrorS(err, "Failed to split meta namespace cache key", "cacheKey", key)
		return fmt.Errorf("can't split meta namespace key: %w", err)
	}

	startTime := time.Now()
	klog.V(4).InfoS("Started syncing PersistentVolumeClaim", "PersistentVolumeClaim", klog.KRef(namespace, name), "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing PersistentVolumeClaim", "PersistentVolumeClaim", klog.KRef(namespace, name), "duration", time.Since(startTime))
	}()

	pvc, err := ppvcc.persistentVolumeClaimLister.PersistentVolumeClaims(namespace).Get(name)
	if apierrors.IsNotFound(err) {
		klog.V(2).InfoS("PersistentVolumeClaim has been deleted", "PersistentVolumeClaim", klog.KRef(namespace, name))
		return nil
	}
	if err != nil {
		return fmt.Errorf("can't get PersistentVolumeClaim: %w", err)
	}

	if pvc.DeletionTimestamp != nil {
		// TODO: should we try to release here as well?

		return nil
	}

	// Sanity check.
	if !controllerhelpers.HasMatchingAnnotation(pvc, volume.AnnStorageProvisioner, naming.DriverName) {
		return fmt.Errorf("PersistentVolumeClaims provided by CSI drivers other than %q are not supported", naming.DriverName)
	}

	if controllerhelpers.HasAnnotation(pvc, naming.DelayedStoragePersistentVolumeClaimBindCompletedAnnotation) {
		// TODO: sync bound
	} else {
		err = ppvcc.syncUnbound(ctx, key, pvc)
		if err != nil {
			return fmt.Errorf("can't sync unbound PersistentVolumeClaim: %w", err)
		}
	}

	return nil
}

func (ppvcc *Controller) syncUnbound(ctx context.Context, key string, pvc *corev1.PersistentVolumeClaim) error {
	klog.V(5).InfoS("Synchronising unbound proxy PersistentVolumeClaim", "ProxyPersistentVolumeClaim", klog.KObj(pvc))

	requestedBackendPersistentVolumeClaimName, ok := pvc.Annotations[naming.DelayedStorageBackendPersistentVolumeClaimRefAnnotation]
	// Sanity check.
	if !ok {
		return fmt.Errorf("PersistentVolumeClaim %q is missing %q annotation, this should never happen", naming.ObjRef(pvc), naming.DelayedStorageBackendPersistentVolumeClaimRefAnnotation)
	}

	if len(requestedBackendPersistentVolumeClaimName) == 0 {
		// TODO: we could potentially support dynamic provisioning of backing pvcs with an additional annotation.
		klog.ErrorS(nil, "PersistentVolumeClaim must not have an empty annotation", "ProxyPersistentVolumeClaim", klog.KObj(pvc), "annotation", naming.DelayedStorageBackendPersistentVolumeClaimRefAnnotation)

		// TODO: record event instead?
		return nil
	}

	klog.V(5).InfoS("Synchronising unbound proxy PersistentVolumeClaim, backend PersistentVolumeClaim requested", "ProxyPersistentVolumeClaim", klog.KObj(pvc), "BackendPersistentVolumeClaim", naming.ManualRef(pvc.Namespace, requestedBackendPersistentVolumeClaimName))
	backendPVC, err := ppvcc.persistentVolumeClaimLister.PersistentVolumeClaims(pvc.Namespace).Get(requestedBackendPersistentVolumeClaimName)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("can't get backing PersistentVolumeClaim %q: %w", naming.ManualRef(pvc.Namespace, requestedBackendPersistentVolumeClaimName), err)
		}

		// User asked for a backend PVC which doesn't exist. We will retry later.
		klog.V(4).InfoS("Synchronising unbound proxy PersistentVolumeClaim, requested backend PersistentVolumeClaim not found", "ProxyPersistentVolumeClaim", klog.KObj(pvc), "BackendPersistentVolumeClaim", naming.ManualRef(pvc.Namespace, requestedBackendPersistentVolumeClaimName))
		return nil
	}

	klog.V(5).InfoS("Synchronising unbound proxy PersistentVolumeClaim, requested backend PersistentVolumeClaim found", "ProxyPersistentVolumeClaim", klog.KObj(pvc), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))

	boundProxyPersistentVolumeClaimName, ok := backendPVC.Annotations[naming.DelayedStorageProxyPersistentVolumeClaimRefAnnotation]
	if !ok || len(boundProxyPersistentVolumeClaimName) == 0 {
		// User asked for a Backend PVC which is not claimed.
		klog.V(5).InfoS("Synchronising unbound proxy PersistentVolumeClaim, requested backend PersistentVolumeClaim is unbound, binding", "ProxyPersistentVolumeClaim", klog.KObj(pvc), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))

		// TODO: check if backend claim satisfies proxy claim?

		err = ppvcc.bind(ctx, pvc, backendPVC)
		if err != nil {
			return fmt.Errorf("can't bind proxy PersistentVolumeClaim %q to backend PersistentVolumeClaim %q: %w", naming.ObjRef(pvc), naming.ObjRef(backendPVC), err)
		}

		return nil
	}

	if boundProxyPersistentVolumeClaimName == pvc.Name {
		// User asked for a backend PVC which is claimed by this proxy PVC.
		klog.V(5).InfoS("Synchronising unbound proxy PersistentVolumeClaim, requested backend PersistentVolumeClaim is already claimed, binding", "ProxyPersistentVolumeClaim", klog.KObj(pvc), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))

		// Finish the binding.
		err = ppvcc.bind(ctx, pvc, backendPVC)
		if err != nil {
			return fmt.Errorf("can't bind proxy PersistentVolumeClaim %q to backend PersistentVolumeClaim %q: %w", naming.ObjRef(pvc), naming.ObjRef(backendPVC), err)
		}

		return nil
	}

	// User asked for a backend PVC which is already claimed by someone else.
	klog.V(4).Info("Synchronising unbound proxy PersistentVolumeClaim, backend PersistentVolumeClaim already bound to a different claim, will retry later", "ProxyPersistentVolumeClaim", klog.KObj(pvc))
	// TODO: record event
	ppvcc.queue.AddRateLimited(key)

	return nil
}

func (ppvcc *Controller) bind(ctx context.Context, proxyPVC *corev1.PersistentVolumeClaim, backendPVC *corev1.PersistentVolumeClaim) error {
	klog.V(5).InfoS("Binding proxy and backend PersistentVolumeClaims", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))

	var err error

	err = ppvcc.bindBackendPVCToProxyPVC(ctx, proxyPVC, backendPVC)
	if err != nil {
		klog.ErrorS(err, "Binding backend PersistentVolumeClaim to proxy PersistentVolumeClaim failed", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))
		return fmt.Errorf("can't bind backend PersistentVolumeClaim to proxy PersistentVolumeClaim: %w", err)
	}

	err = ppvcc.completeProxyPVCBind(ctx, proxyPVC)
	if err != nil {
		klog.ErrorS(err, "Completing binding of proxy PersistentVolumeClaim failed", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))
		return fmt.Errorf("can't complete binding of proxy PersistentVolumeClaim: %w", err)
	}

	klog.V(4).InfoS("Successfully bound proxy and backend PersistentVolumeClaims", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))

	return nil
}

func (ppvcc *Controller) bindBackendPVCToProxyPVC(ctx context.Context, proxyPVC *corev1.PersistentVolumeClaim, backendPVC *corev1.PersistentVolumeClaim) error {
	klog.V(5).InfoS("Binding backend PersistentVolumeClaim to proxy PersistentVolumeClaim", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))
	var err error

	if isBackendPVCBoundToProxyPVC(proxyPVC, backendPVC) {
		klog.V(5).InfoS("Backend PersistentVolumeClaim is already bound to proxy PersistentVolumeClaim", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))
		return nil
	}

	backendPVCCopy := backendPVC.DeepCopy()

	if backendPVCCopy.Annotations == nil {
		backendPVCCopy.Annotations = map[string]string{}
	}
	backendPVCCopy.Annotations[naming.DelayedStorageProxyPersistentVolumeClaimRefAnnotation] = proxyPVC.Name

	_, err = ppvcc.kubeClient.CoreV1().PersistentVolumeClaims(backendPVCCopy.Namespace).Update(ctx, backendPVCCopy, metav1.UpdateOptions{})
	if err != nil {
		klog.ErrorS(err, "Updating backend PersistentVolumeClaim: binding to proxy PersistentVolumeClaim failed", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))
		return fmt.Errorf("can't update backend PersistentVolumeClaim %q: %w", naming.ObjRef(backendPVC), err)
	}

	klog.V(4).InfoS("Backend PersistentVolumeClaim successfully bound to proxy PersistentVolumeClaim", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC), "BackendPersistentVolumeClaim", naming.ObjRef(backendPVC))
	return nil
}

func (ppvcc *Controller) completeProxyPVCBind(ctx context.Context, proxyPVC *corev1.PersistentVolumeClaim) error {
	klog.V(5).InfoS("Completing binding, marking proxy PersistentVolumeClaim as bound", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC))
	var err error

	if metav1.HasAnnotation(proxyPVC.ObjectMeta, naming.DelayedStoragePersistentVolumeClaimBindCompletedAnnotation) {
		klog.V(5).InfoS("Proxy PersistentVolumeClaim has already completed binding", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC))
		return nil
	}

	proxyPVCCopy := proxyPVC.DeepCopy()
	metav1.SetMetaDataAnnotation(&proxyPVCCopy.ObjectMeta, naming.DelayedStoragePersistentVolumeClaimBindCompletedAnnotation, "")

	_, err = ppvcc.kubeClient.CoreV1().PersistentVolumeClaims(proxyPVCCopy.Namespace).Update(ctx, proxyPVCCopy, metav1.UpdateOptions{})
	if err != nil {
		klog.ErrorS(err, "Completing proxy PersistentVolumeClaim binding failed", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC))
		return fmt.Errorf("can't update proxy PersistentVolumeClaim %q: %w", naming.ObjRef(proxyPVC), err)
	}

	klog.V(4).InfoS("Proxy PersistentVolumeClaim binding completed", "ProxyPersistentVolumeClaim", klog.KObj(proxyPVC))
	return nil
}

func isBackendPVCBoundToProxyPVC(proxyPVC *corev1.PersistentVolumeClaim, backendPVC *corev1.PersistentVolumeClaim) bool {
	proxyPersistentVolumeClaimRefAnnotation := backendPVC.Annotations[naming.DelayedStorageProxyPersistentVolumeClaimRefAnnotation]

	return proxyPersistentVolumeClaimRefAnnotation == proxyPVC.Name
}
