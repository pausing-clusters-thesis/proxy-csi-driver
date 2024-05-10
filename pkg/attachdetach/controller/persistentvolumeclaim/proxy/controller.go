package proxy

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	"k8s.io/component-helpers/storage/volume"
	"k8s.io/klog/v2"
)

const (
	controllerName = "ProxyPersistentVolumeClaimController"
)

var (
	keyFunc                                 = cache.DeletionHandlingMetaNamespaceKeyFunc
	proxyPersistentVolumeClaimControllerGVK = corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim")
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

	ppvcc := &Controller{
		kubeClient: kubeClient,

		persistentVolumeClaimLister: persistentVolumeClaimInformer.Lister(),

		cachesToSync: []cache.InformerSynced{
			persistentVolumeClaimInformer.Informer().HasSynced,
		},

		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "proxypersistentvolumeclaim-controller"}),

		queue: workqueue.NewRateLimitingQueueWithConfig(
			workqueue.DefaultControllerRateLimiter(),
			workqueue.RateLimitingQueueConfig{
				Name: "proxypersistentvolumeclaim",
			},
		),
	}

	ppvcc.handlers, err = socontrollerhelpers.NewHandlers[*corev1.PersistentVolumeClaim](
		ppvcc.queue,
		keyFunc,
		scheme.Scheme,
		proxyPersistentVolumeClaimControllerGVK,
		kubeinterfaces.NamespacedGetList[*corev1.PersistentVolumeClaim]{
			GetFunc: func(namespace, name string) (*corev1.PersistentVolumeClaim, error) {
				return ppvcc.persistentVolumeClaimLister.PersistentVolumeClaims(namespace).Get(name)
			},
			ListFunc: func(namespace string, selector labels.Selector) ([]*corev1.PersistentVolumeClaim, error) {
				return ppvcc.persistentVolumeClaimLister.PersistentVolumeClaims(namespace).List(selector)
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't create handlers: %w", err)
	}

	persistentVolumeClaimInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ppvcc.addPersistentVolumeClaim,
		UpdateFunc: ppvcc.updatePersistentVolumeClaim,
		DeleteFunc: ppvcc.deletePersistentVolumeClaim,
	})

	return ppvcc, nil
}

func (ppvcc *Controller) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()

	klog.InfoS("Starting controller", "controller", controllerName)

	var wg sync.WaitGroup
	defer func() {
		klog.InfoS("Shutting down controller", "controller", controllerName)
		ppvcc.queue.ShutDown()
		wg.Wait()
		klog.InfoS("Shut down controller", "controller", controllerName)
	}()

	if !cache.WaitForNamedCacheSync(controllerName, ctx.Done(), ppvcc.cachesToSync...) {
		return
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wait.UntilWithContext(ctx, ppvcc.runWorker, time.Second)
		}()
	}

	<-ctx.Done()
}

func (ppvcc *Controller) runWorker(ctx context.Context) {
	for ppvcc.processNextItem(ctx) {
	}
}

func (ppvcc *Controller) processNextItem(ctx context.Context) bool {
	key, quit := ppvcc.queue.Get()
	if quit {
		return false
	}
	defer ppvcc.queue.Done(key)

	err := ppvcc.sync(ctx, key.(string))
	// TODO: Do smarter filtering then just Reduce to handle cases like 2 conflict errors.
	err = utilerrors.Reduce(err)
	switch {
	case err == nil:
		ppvcc.queue.Forget(key)
		return true

	case apierrors.IsConflict(err):
		klog.V(2).InfoS("Hit conflict, will retry in a bit", "Key", key, "Error", err)

	case apierrors.IsAlreadyExists(err):
		klog.V(2).InfoS("Hit already exists, will retry in a bit", "Key", key, "Error", err)

	default:
		utilruntime.HandleError(fmt.Errorf("syncing key '%v' failed: %v", key, err))
	}

	ppvcc.queue.AddRateLimited(key)

	return true
}

func (ppvcc *Controller) addPersistentVolumeClaim(obj interface{}) {
	ppvcc.handlers.HandleAdd(
		obj.(*corev1.PersistentVolumeClaim),
		ppvcc.enqueueProxyPVCDirectlyOrThroughAnnotation,
	)
}

func (ppvcc *Controller) updatePersistentVolumeClaim(old, cur interface{}) {
	ppvcc.handlers.HandleUpdate(
		old.(*corev1.PersistentVolumeClaim),
		cur.(*corev1.PersistentVolumeClaim),
		ppvcc.enqueueProxyPVCDirectlyOrThroughAnnotation,
		ppvcc.deletePersistentVolumeClaim,
	)
}

func (ppvcc *Controller) deletePersistentVolumeClaim(obj interface{}) {
	ppvcc.handlers.HandleDelete(
		obj.(*corev1.PersistentVolumeClaim),
		ppvcc.enqueueProxyPVCDirectlyOrThroughAnnotation,
	)
}

func (ppvcc *Controller) enqueueProxyPVCDirectlyOrThroughAnnotation(depth int, obj kubeinterfaces.ObjectInterface, op socontrollerhelpers.HandlerOperationType) {
	pvc := obj.(*corev1.PersistentVolumeClaim)

	if isProxyPersistentVolumeClaim(pvc) {
		ppvcc.handlers.Enqueue(depth+1, obj, op)
		return
	}

	isBackedByPVC := func(proxyPVC *corev1.PersistentVolumeClaim) bool {
		return controllerhelpers.HasMatchingAnnotation(proxyPVC, volume.AnnStorageProvisioner, naming.DriverName) &&
			controllerhelpers.HasMatchingAnnotation(proxyPVC, naming.DelayedStorageBackendPersistentVolumeClaimRefAnnotation, pvc.Name)
	}
	ppvcc.handlers.EnqueueAllFunc(ppvcc.handlers.EnqueueWithFilterFunc(isBackedByPVC))(depth+1, obj, op)
}

func isProxyPersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) bool {
	return controllerhelpers.HasMatchingAnnotation(pvc, volume.AnnStorageProvisioner, naming.DriverName) &&
		metav1.HasAnnotation(pvc.ObjectMeta, naming.DelayedStorageBackendPersistentVolumeClaimRefAnnotation)
}
