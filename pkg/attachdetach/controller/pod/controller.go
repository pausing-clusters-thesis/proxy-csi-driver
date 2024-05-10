package pod

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/controllerhelpers"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/scheme"
	socontrollerhelpers "github.com/scylladb/scylla-operator/pkg/controllerhelpers"
	"github.com/scylladb/scylla-operator/pkg/kubeinterfaces"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1informers "k8s.io/client-go/informers/core/v1"
	storagev1informers "k8s.io/client-go/informers/storage/v1"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	storagev1listers "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	volumehelpers "k8s.io/component-helpers/storage/volume"
	"k8s.io/klog/v2"
)

const (
	controllerName = "PodController"
)

var (
	keyFunc          = cache.DeletionHandlingMetaNamespaceKeyFunc
	podControllerGVK = corev1.SchemeGroupVersion.WithKind("Pod")
)

type Controller struct {
	kubeClient kubernetes.Interface

	csiDriverLister        storagev1listers.CSIDriverLister
	volumeAttachmentLister storagev1listers.VolumeAttachmentLister

	podLister                   corev1listers.PodLister
	persistentVolumeClaimLister corev1listers.PersistentVolumeClaimLister
	persistentVolumeLister      corev1listers.PersistentVolumeLister

	cachesToSync []cache.InformerSynced

	eventRecorder record.EventRecorder

	queue    workqueue.RateLimitingInterface
	handlers *socontrollerhelpers.Handlers[*corev1.Pod]
}

func NewController(
	kubeClient kubernetes.Interface,
	volumeAttachmentInformer storagev1informers.VolumeAttachmentInformer,
	csiDriverInformer storagev1informers.CSIDriverInformer,
	podInformer corev1informers.PodInformer,
	persistentVolumeClaimInformer corev1informers.PersistentVolumeClaimInformer,
	persistentVolumeInformer corev1informers.PersistentVolumeInformer,
) (*Controller, error) {
	var err error

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	pc := &Controller{
		kubeClient: kubeClient,

		csiDriverLister:        csiDriverInformer.Lister(),
		volumeAttachmentLister: volumeAttachmentInformer.Lister(),

		podLister:                   podInformer.Lister(),
		persistentVolumeClaimLister: persistentVolumeClaimInformer.Lister(),
		persistentVolumeLister:      persistentVolumeInformer.Lister(),

		cachesToSync: []cache.InformerSynced{
			csiDriverInformer.Informer().HasSynced,
			volumeAttachmentInformer.Informer().HasSynced,
			podInformer.Informer().HasSynced,
			persistentVolumeClaimInformer.Informer().HasSynced,
			persistentVolumeInformer.Informer().HasSynced,
		},

		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "pod-controller"}),

		queue: workqueue.NewRateLimitingQueueWithConfig(
			workqueue.DefaultControllerRateLimiter(),
			workqueue.RateLimitingQueueConfig{
				Name: "pod",
			},
		),
	}

	pc.handlers, err = socontrollerhelpers.NewHandlers[*corev1.Pod](
		pc.queue,
		keyFunc,
		scheme.Scheme,
		podControllerGVK,
		kubeinterfaces.NamespacedGetList[*corev1.Pod]{
			GetFunc: func(namespace, name string) (*corev1.Pod, error) {
				return pc.podLister.Pods(namespace).Get(name)
			},
			ListFunc: func(namespace string, selector labels.Selector) ([]*corev1.Pod, error) {
				return pc.podLister.Pods(namespace).List(selector)
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't create handlers: %w", err)
	}

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    pc.addPod,
		UpdateFunc: pc.updatePod,
		DeleteFunc: pc.deletePod,
	})

	persistentVolumeClaimInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    pc.addPersistentVolumeClaim,
		UpdateFunc: pc.updatePersistentVolumeClaim,
		DeleteFunc: pc.deletePersistentVolumeClaim,
	})

	persistentVolumeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    pc.addPersistentVolume,
		UpdateFunc: pc.updatePersistentVolume,
		DeleteFunc: pc.deletePersistentVolume,
	})

	volumeAttachmentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    pc.addVolumeAttachment,
		UpdateFunc: pc.updateVolumeAttachment,
		DeleteFunc: pc.deleteVolumeAttachment,
	})

	return pc, nil
}

func (pc *Controller) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()

	klog.InfoS("Starting controller", "controller", controllerName)

	var wg sync.WaitGroup
	defer func() {
		klog.InfoS("Shutting down controller", "controller", controllerName)
		pc.queue.ShutDown()
		wg.Wait()
		klog.InfoS("Shut down controller", "controller", controllerName)
	}()

	if !cache.WaitForNamedCacheSync(controllerName, ctx.Done(), pc.cachesToSync...) {
		return
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wait.UntilWithContext(ctx, pc.runWorker, time.Second)
		}()
	}

	<-ctx.Done()
}

func (pc *Controller) runWorker(ctx context.Context) {
	for pc.processNextItem(ctx) {
	}
}

func (pc *Controller) processNextItem(ctx context.Context) bool {
	key, quit := pc.queue.Get()
	if quit {
		return false
	}
	defer pc.queue.Done(key)

	err := pc.sync(ctx, key.(string))
	// TODO: Do smarter filtering then just Reduce to handle cases like 2 conflict errors.
	err = utilerrors.Reduce(err)
	switch {
	case err == nil:
		pc.queue.Forget(key)
		return true

	case apierrors.IsConflict(err):
		klog.V(2).InfoS("Hit conflict, will retry in a bit", "Key", key, "Error", err)

	case apierrors.IsAlreadyExists(err):
		klog.V(2).InfoS("Hit already exists, will retry in a bit", "Key", key, "Error", err)

	default:
		utilruntime.HandleError(fmt.Errorf("syncing key '%v' failed: %v", key, err))
	}

	pc.queue.AddRateLimited(key)

	return true
}

func (pc *Controller) addPod(obj interface{}) {
	pc.handlers.HandleAdd(
		obj.(*corev1.Pod),
		pc.enqueuePodWithBoundProxyVolume,
	)
}

func (pc *Controller) updatePod(old, cur interface{}) {
	pc.handlers.HandleUpdate(
		old.(*corev1.Pod),
		cur.(*corev1.Pod),
		pc.enqueuePodWithBoundProxyVolume,
		pc.deletePod,
	)
}

func (pc *Controller) deletePod(obj interface{}) {
	pc.handlers.HandleDelete(
		obj.(*corev1.Pod),
		pc.enqueuePodWithBoundProxyVolume,
	)
}

func (pc *Controller) enqueuePodWithBoundProxyVolume(depth int, untypedObj kubeinterfaces.ObjectInterface, op socontrollerhelpers.HandlerOperationType) {
	pc.handlers.EnqueueWithFilterFunc(pc.podHasProxyVolume)(depth+1, untypedObj, op)
}

func (pc *Controller) podHasProxyVolume(pod *corev1.Pod) bool {
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}

		pvc, err := pc.persistentVolumeClaimLister.PersistentVolumeClaims(pod.Namespace).Get(volume.PersistentVolumeClaim.ClaimName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// TODO: log?
				continue
			}

			utilruntime.HandleError(err)
			return false
		}

		if !isBoundProxyPersistentVolumeClaim(pvc) {
			continue
		}

		return true
	}

	return false
}

func (pc *Controller) addPersistentVolumeClaim(obj interface{}) {
	pc.handlers.HandleAdd(
		obj.(*corev1.PersistentVolumeClaim),
		pc.enqueuePodsThroughPersistentVolumeClaim,
	)
}

func (pc *Controller) updatePersistentVolumeClaim(old, cur interface{}) {
	pc.handlers.HandleUpdate(
		old.(*corev1.PersistentVolumeClaim),
		cur.(*corev1.PersistentVolumeClaim),
		pc.enqueuePodsThroughPersistentVolumeClaim,
		pc.deletePersistentVolumeClaim,
	)
}

func (pc *Controller) deletePersistentVolumeClaim(obj interface{}) {
	pc.handlers.HandleDelete(
		obj.(*corev1.PersistentVolumeClaim),
		pc.enqueuePodsThroughPersistentVolumeClaim,
	)
}

func (pc *Controller) enqueuePodsThroughPersistentVolumeClaim(depth int, obj kubeinterfaces.ObjectInterface, op socontrollerhelpers.HandlerOperationType) {
	pvc := obj.(*corev1.PersistentVolumeClaim)

	if !controllerhelpers.HasAnnotation(pvc, naming.DelayedStorageBackendPersistentVolumeClaimRefAnnotation) {
		// PVC is not a proxy PVC, but it can still be backingproxy volume bound to a backend volume., so we check for a proxy ref.
		proxyPVCRef, hasProxyPVCRef := pvc.Annotations[naming.DelayedStorageProxyPersistentVolumeClaimRefAnnotation]
		if !hasProxyPVCRef {
			return
		}

		proxyPVC, err := pc.persistentVolumeClaimLister.PersistentVolumeClaims(pvc.Namespace).Get(proxyPVCRef)
		if err != nil {
			utilruntime.HandleError(err)
			return
		}

		pvc = proxyPVC
	}

	if !isBoundProxyPersistentVolumeClaim(pvc) {
		return
	}

	podReferencesPersistentVolumeClaim := func(pod *corev1.Pod) bool {
		for _, volume := range pod.Spec.Volumes {
			pvSrc := volume.VolumeSource.PersistentVolumeClaim
			if pvSrc == nil {
				continue
			}

			if pvSrc.ClaimName == pvc.Name {
				return true
			}
		}

		return false
	}
	pc.handlers.EnqueueAllFunc(pc.handlers.EnqueueWithFilterFunc(podReferencesPersistentVolumeClaim))(depth+1, pvc, op)
}

func isBoundProxyPersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) bool {
	return controllerhelpers.HasMatchingAnnotation(pvc, volumehelpers.AnnStorageProvisioner, naming.DriverName) &&
		controllerhelpers.HasAnnotation(pvc, naming.DelayedStorageBackendPersistentVolumeClaimRefAnnotation) &&
		controllerhelpers.HasAnnotation(pvc, naming.DelayedStoragePersistentVolumeClaimBindCompletedAnnotation)
}

func (pc *Controller) addPersistentVolume(obj interface{}) {
	pc.handlers.HandleAdd(
		obj.(*corev1.PersistentVolume),
		pc.enqueuePodsThroughPersistentVolume,
	)
}

func (pc *Controller) updatePersistentVolume(old, cur interface{}) {
	pc.handlers.HandleUpdate(
		old.(*corev1.PersistentVolume),
		cur.(*corev1.PersistentVolume),
		pc.enqueuePodsThroughPersistentVolume,
		pc.deletePersistentVolume,
	)
}

func (pc *Controller) deletePersistentVolume(obj interface{}) {
	pc.handlers.HandleDelete(
		obj.(*corev1.PersistentVolume),
		pc.enqueuePodsThroughPersistentVolume,
	)
}

func (pc *Controller) enqueuePodsThroughPersistentVolume(depth int, obj kubeinterfaces.ObjectInterface, op socontrollerhelpers.HandlerOperationType) {
	pv := obj.(*corev1.PersistentVolume)

	if pv.Spec.ClaimRef == nil {
		return
	}

	pvc, err := pc.persistentVolumeClaimLister.PersistentVolumeClaims(pv.Namespace).Get(pv.Name)
	if err != nil {
		return
	}

	pc.enqueuePodsThroughPersistentVolumeClaim(depth+1, pvc, op)
}

func (pc *Controller) addVolumeAttachment(obj interface{}) {
	pc.handlers.HandleAdd(
		obj.(*storagev1.VolumeAttachment),
		pc.enqueuePodsThroughVolumeAttachment,
	)
}

func (pc *Controller) updateVolumeAttachment(old, cur interface{}) {
	pc.handlers.HandleUpdate(
		old.(*storagev1.VolumeAttachment),
		cur.(*storagev1.VolumeAttachment),
		pc.enqueuePodsThroughVolumeAttachment,
		pc.deleteVolumeAttachment,
	)
}

func (pc *Controller) deleteVolumeAttachment(obj interface{}) {
	pc.handlers.HandleDelete(
		obj.(*storagev1.VolumeAttachment),
		pc.enqueuePodsThroughVolumeAttachment,
	)
}

func (pc *Controller) enqueuePodsThroughVolumeAttachment(depth int, obj kubeinterfaces.ObjectInterface, op socontrollerhelpers.HandlerOperationType) {
	va := obj.(*storagev1.VolumeAttachment)

	if va.Spec.Source.PersistentVolumeName == nil {
		return
	}

	pv, err := pc.persistentVolumeLister.Get(*va.Spec.Source.PersistentVolumeName)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	if pv.Spec.ClaimRef == nil {
		return
	}

	pvc, err := pc.persistentVolumeClaimLister.PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name)
	if err != nil {
		return
	}

	pc.enqueuePodsThroughPersistentVolumeClaim(depth+1, pvc, op)
}
