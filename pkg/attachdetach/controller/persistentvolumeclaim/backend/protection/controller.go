package protection

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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	controllerName = "BackingPersistentVolumeClaimProtectionController"
)

var (
	keyFunc                                             = cache.DeletionHandlingMetaNamespaceKeyFunc
	backingPersistentVolumeClaimProtectionControllerGVK = corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim")
)

type Controller struct {
	kubeClient kubernetes.Interface

	persistentVolumeClaimLister corev1listers.PersistentVolumeClaimLister

	cachesToSync []cache.InformerSynced

	eventRecorder record.EventRecorder

	queue    workqueue.RateLimitingInterface
	handlers *socontrollerhelpers.Handlers[*corev1.PersistentVolumeClaim]
}

func NewController(
	kubeClient kubernetes.Interface,
	persistentVolumeClaimInformer corev1informers.PersistentVolumeClaimInformer,
) (*Controller, error) {
	var err error

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	bpvcpc := &Controller{
		kubeClient: kubeClient,

		persistentVolumeClaimLister: persistentVolumeClaimInformer.Lister(),

		cachesToSync: []cache.InformerSynced{
			persistentVolumeClaimInformer.Informer().HasSynced,
		},

		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "backingpersistentvolumeclaimprotection-controller"}),

		queue: workqueue.NewRateLimitingQueueWithConfig(
			workqueue.DefaultControllerRateLimiter(),
			workqueue.RateLimitingQueueConfig{
				Name: "backingpersistentvolumeclaimprotection",
			},
		),
	}
	bpvcpc.handlers, err = socontrollerhelpers.NewHandlers[*corev1.PersistentVolumeClaim](
		bpvcpc.queue,
		keyFunc,
		scheme.Scheme,
		backingPersistentVolumeClaimProtectionControllerGVK,
		kubeinterfaces.NamespacedGetList[*corev1.PersistentVolumeClaim]{
			GetFunc: func(namespace, name string) (*corev1.PersistentVolumeClaim, error) {
				return bpvcpc.persistentVolumeClaimLister.PersistentVolumeClaims(namespace).Get(name)
			},
			ListFunc: func(namespace string, selector labels.Selector) ([]*corev1.PersistentVolumeClaim, error) {
				return bpvcpc.persistentVolumeClaimLister.PersistentVolumeClaims(namespace).List(selector)
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't create handlers: %w", err)
	}

	persistentVolumeClaimInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    bpvcpc.addPersistentVolumeClaim,
		UpdateFunc: bpvcpc.updatePersistentVolumeClaim,
		DeleteFunc: bpvcpc.deletePersistentVolumeClaim,
	})

	return bpvcpc, nil
}

func (bpvcpc *Controller) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()

	klog.InfoS("Starting controller", "controller", controllerName)

	var wg sync.WaitGroup
	defer func() {
		klog.InfoS("Shutting down controller", "controller", controllerName)
		bpvcpc.queue.ShutDown()
		wg.Wait()
		klog.InfoS("Shut down controller", "controller", controllerName)
	}()

	if !cache.WaitForNamedCacheSync(controllerName, ctx.Done(), bpvcpc.cachesToSync...) {
		return
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wait.UntilWithContext(ctx, bpvcpc.runWorker, time.Second)
		}()
	}

	<-ctx.Done()
}

func (bpvcpc *Controller) runWorker(ctx context.Context) {
	for bpvcpc.processNextItem(ctx) {
	}
}

func (bpvcpc *Controller) processNextItem(ctx context.Context) bool {
	key, quit := bpvcpc.queue.Get()
	if quit {
		return false
	}
	defer bpvcpc.queue.Done(key)

	err := bpvcpc.sync(ctx, key.(string))
	// TODO: Do smarter filtering then just Reduce to handle cases like 2 conflict errors.
	err = utilerrors.Reduce(err)
	switch {
	case err == nil:
		bpvcpc.queue.Forget(key)
		return true

	case apierrors.IsConflict(err):
		klog.V(2).InfoS("Hit conflict, will retry in a bit", "Key", key, "Error", err)

	case apierrors.IsAlreadyExists(err):
		klog.V(2).InfoS("Hit already exists, will retry in a bit", "Key", key, "Error", err)

	default:
		utilruntime.HandleError(fmt.Errorf("syncing key '%v' failed: %v", key, err))
	}

	bpvcpc.queue.AddRateLimited(key)

	return true
}

func (bpvcpc *Controller) addPersistentVolumeClaim(obj interface{}) {
	bpvcpc.handlers.HandleAdd(
		obj.(*corev1.PersistentVolumeClaim),
		bpvcpc.enqueueBackingPVCDirectlyOrThroughAnnotation,
	)
}

func (bpvcpc *Controller) updatePersistentVolumeClaim(old, cur interface{}) {
	bpvcpc.handlers.HandleUpdate(
		old.(*corev1.PersistentVolumeClaim),
		cur.(*corev1.PersistentVolumeClaim),
		bpvcpc.enqueueBackingPVCDirectlyOrThroughAnnotation,
		bpvcpc.deletePersistentVolumeClaim,
	)
}

func (bpvcpc *Controller) deletePersistentVolumeClaim(obj interface{}) {
	bpvcpc.handlers.HandleDelete(
		obj.(*corev1.PersistentVolumeClaim),
		bpvcpc.enqueueBackingPVCDirectlyOrThroughAnnotation,
	)
}

func (bpvcpc *Controller) enqueueBackingPVCDirectlyOrThroughAnnotation(depth int, obj kubeinterfaces.ObjectInterface, op socontrollerhelpers.HandlerOperationType) {
	pvc := obj.(*corev1.PersistentVolumeClaim)

	// We need to check for the finalizer's presence as well because a PVC might end up having a finalizer and no longer having the annotation after it's been released.
	if controllerhelpers.HasAnnotation(pvc, naming.DelayedStorageProxyPersistentVolumeClaimRefAnnotation) || controllerhelpers.HasFinalizer(pvc, naming.DelayedStorageBackendPersistentVolumeClaimProtectionFinalizer) {
		bpvcpc.handlers.Enqueue(depth+1, obj, op)
		return
	}

	if !controllerhelpers.HasAnnotation(pvc, naming.DelayedStoragePersistentVolumeClaimBindCompletedAnnotation) {
		return
	}

	backingPVCName, ok := pvc.Annotations[naming.DelayedStorageBackendPersistentVolumeClaimRefAnnotation]
	// Sanity check.
	if !ok {
		return
	}

	backingPVC, err := bpvcpc.persistentVolumeClaimLister.PersistentVolumeClaims(pvc.Namespace).Get(backingPVCName)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	bpvcpc.handlers.Enqueue(depth+1, backingPVC, op)
}
