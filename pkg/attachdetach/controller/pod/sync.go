package pod

import (
	"context"
	"fmt"
	"time"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/controllerhelpers"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volume"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/cache"
	volumehelpers "k8s.io/component-helpers/storage/volume"
	"k8s.io/klog/v2"
)

func (pc *Controller) sync(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.ErrorS(err, "Failed to split meta namespace cache key", "cacheKey", key)
		return fmt.Errorf("can't split meta namespace key: %w", err)
	}

	now := time.Now()
	klog.V(4).InfoS("Started syncing Pod", "Pod", klog.KRef(namespace, name), "startTime", now)
	defer func() {
		klog.V(4).InfoS("Finished syncing Pod", "Pod", klog.KRef(namespace, name), "duration", time.Since(now))
	}()

	pod, err := pc.podLister.Pods(namespace).Get(name)
	if err != nil {
		return fmt.Errorf("can't get pod %q: %w", naming.ManualRef(namespace, name), err)
	}

	if pod.DeletionTimestamp != nil {
		// TODO: do we have to do anything here?
		return nil
	}

	// TODO: doing it sequentially for starters, we can consider introducing concurrency later
	var errs []error
	for _, v := range pod.Spec.Volumes {
		if v.PersistentVolumeClaim == nil {
			continue
		}

		pvc, err := pc.persistentVolumeClaimLister.PersistentVolumeClaims(pod.Namespace).Get(v.PersistentVolumeClaim.ClaimName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				errs = append(errs, fmt.Errorf("can't get persistent volume claim %q: %w", naming.ManualRef(pod.Namespace, v.PersistentVolumeClaim.ClaimName), err))
			}

			continue
		}

		if !controllerhelpers.HasMatchingAnnotation(pvc, volumehelpers.AnnStorageProvisioner, naming.DriverName) {
			continue
		}

		// Sanity check, this should never happen.
		if len(pvc.Spec.VolumeName) == 0 {
			klog.ErrorS(nil, "Proxy PersistentVolumeClaim does not have a bound PersistentVolume, this should never happen", "Pod", klog.KObj(pod), "PersistentVolumeClaim", klog.KObj(pvc), "Volume", v.Name)
			continue
		}

		pv, err := pc.persistentVolumeLister.Get(pvc.Spec.VolumeName)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't get persistent volume %q: %w", pvc.Spec.VolumeName, err))
			continue
		}

		if !controllerhelpers.HasAnnotation(pvc, naming.DelayedStoragePersistentVolumeClaimBindCompletedAnnotation) {
			klog.V(4).InfoS("Proxy PersistentVolumeClaim does not have a bound backend PersistentVolumeClaim yet", "Pod", klog.KObj(pod), "PersistentVolumeClaim", klog.KObj(pvc), "Volume", v.Name)
			continue
		}

		backendPVCName, ok := pvc.Annotations[naming.DelayedStorageBackendPersistentVolumeClaimRefAnnotation]
		// Sanity check, this should never happen.
		if !ok || len(backendPVCName) == 0 {
			klog.ErrorS(nil, "Proxy PersistentVolumeClaim is marked as bound but it doesn't have a backend PersistentVolumeClaim, this should never happen", "Pod", klog.KObj(pod), "PersistentVolumeClaim", klog.KObj(pvc), "Volume", v.Name)
			continue
		}

		backendPVC, err := pc.persistentVolumeClaimLister.PersistentVolumeClaims(pod.Namespace).Get(backendPVCName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				errs = append(errs, fmt.Errorf("can't get persistent volume claim %q: %w", naming.ManualRef(pod.Namespace, backendPVCName), err))
			}

			klog.V(2).InfoS("backend PersistentVolumeClaim not found", "Pod", klog.KObj(pod), "ProxyPersistentVolumeClaim", klog.KObj(pvc), "BackendPersistentVolumeClaim", klog.KRef(pod.Namespace, backendPVCName), "Volume", v.Name)
			continue
		}

		if len(backendPVC.Spec.VolumeName) == 0 {
			klog.V(2).InfoS("Backend PersistentVolumeClaim does not have a bound PersistentVolume yet", "Pod", klog.KObj(pod), "ProxyPersistentVolumeClaim", klog.KObj(pvc), "BackendPersistentVolumeClaim", klog.KObj(backendPVC), "Volume", v.Name)
			continue
		}

		backendPV, err := pc.persistentVolumeLister.Get(backendPVC.Spec.VolumeName)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't get persistent volume %q: %w", backendPVC.Spec.VolumeName, err))
			continue
		}

		spec := &volume.Spec{
			Volume:                       &v,
			Pod:                          pod,
			ProxyPersistentVolumeClaim:   pvc,
			ProxyPersistentVolume:        pv,
			BackendPersistentVolumeClaim: backendPVC,
			BackendPersistentVolume:      backendPV,
		}
		err = pc.syncVolume(ctx, spec)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't sync volume %q: %w", v.Name, err))
		}
	}

	err = utilerrors.NewAggregate(errs)
	if err != nil {
		return err
	}

	return nil
}
