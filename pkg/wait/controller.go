package wait

import (
	"context"
	"fmt"
	"time"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/controllerhelpers"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	"github.com/scylladb/scylla-operator/pkg/controllertools"
	sohelpers "github.com/scylladb/scylla-operator/pkg/helpers"
	corev1 "k8s.io/api/core/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
)

type Controller struct {
	*controllertools.Observer

	namespace           string
	podName             string
	requiredAnnotations map[string]string
	signalFilePath      string

	podLister corev1listers.PodLister
}

func NewController(
	namespace string,
	podName string,
	requiredAnnotations map[string]string,
	signalFilePath string,
	kubeClient kubernetes.Interface,
	podInformer corev1informers.PodInformer,
) (*Controller, error) {
	controller := &Controller{
		namespace:           namespace,
		podName:             podName,
		requiredAnnotations: requiredAnnotations,
		signalFilePath:      signalFilePath,
		podLister:           podInformer.Lister(),
	}

	observer := controllertools.NewObserver(
		"proxy-csi-driver-wait",
		kubeClient.CoreV1().Events(corev1.NamespaceAll),
		controller.sync,
	)

	podHandler, err := podInformer.Informer().AddEventHandler(observer.GetGenericHandlers())
	if err != nil {
		return nil, fmt.Errorf("can't add Pod event handler: %w", err)
	}
	observer.AddCachesToSync(podHandler.HasSynced)

	controller.Observer = observer

	return controller, nil
}

func (c *Controller) sync(ctx context.Context) error {
	startTime := time.Now()
	klog.V(4).InfoS("Started syncing observer", "Name", c.Observer.Name(), "startTime", startTime)
	defer func() {
		klog.V(4).InfoS("Finished syncing observer", "Name", c.Observer.Name(), "duration", time.Since(startTime))
	}()

	annotated, err := c.evaluateAnnotations()
	if err != nil {
		return fmt.Errorf("can't evaluate annotations: %w", err)
	}

	if annotated {
		klog.V(2).InfoS("Pod has all required annotations", "SignalFile", c.signalFilePath)
		err = sohelpers.TouchFile(c.signalFilePath)
		if err != nil {
			return fmt.Errorf("can't touch signal file %q: %w", c.signalFilePath, err)
		}
	} else {
		klog.V(2).InfoS("Waiting for required annotations", "RequiredAnnotations", c.requiredAnnotations)
	}

	return nil
}

func (c *Controller) evaluateAnnotations() (bool, error) {
	pod, err := c.podLister.Pods(c.namespace).Get(c.podName)
	if err != nil {
		return false, fmt.Errorf("can't get Pod %q: %w", naming.ManualRef(c.namespace, c.podName), err)
	}

	for k, v := range c.requiredAnnotations {
		if !controllerhelpers.HasMatchingAnnotation(pod, k, v) {
			return false, nil
		}
	}

	return true, nil
}
