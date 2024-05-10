package proxy_csi_driver

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/version"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/wait"
	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"github.com/scylladb/scylla-operator/pkg/signals"
	"github.com/spf13/cobra"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

type WaitOptions struct {
	genericclioptions.ClientConfig
	genericclioptions.InClusterReflection

	PodName        string
	Volumes        []string
	SignalFilePath string

	requiredAnnotations map[string]string

	kubeClient kubernetes.Interface
}

func NewWaitOptions(streams genericclioptions.IOStreams) *WaitOptions {
	return &WaitOptions{
		ClientConfig: genericclioptions.NewClientConfig("proxy-csi-driver-wait"),

		Volumes:        []string{},
		SignalFilePath: "",

		requiredAnnotations: map[string]string{},
	}
}

func NewWaitCommand(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewWaitOptions(streams)

	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for proxy Volumes.",
		Long:  "Wait for all proxy Volumes used by the host pod to be mounted.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return err
			}

			err = o.Complete()
			if err != nil {
				return err
			}

			err = o.Run(streams, cmd)
			if err != nil {
				return err
			}

			return nil
		},

		SilenceErrors: true,
		SilenceUsage:  true,
	}

	o.ClientConfig.AddFlags(cmd)
	o.InClusterReflection.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.PodName, "pod-name", "", o.PodName, "Name of the pod.")
	cmd.Flags().StringSliceVar(&o.Volumes, "volume", o.Volumes, "Volume(s) to wait for.")
	cmd.Flags().StringVar(&o.SignalFilePath, "signal-file-path", o.SignalFilePath, "Ready file path.")

	return cmd
}

func (o *WaitOptions) Validate() error {
	var errs []error

	errs = append(errs, o.ClientConfig.Validate())
	errs = append(errs, o.InClusterReflection.Validate())

	if len(o.PodName) == 0 {
		errs = append(errs, fmt.Errorf("pod-name can't be empty"))
	} else {
		podNameValidationErrs := apimachineryvalidation.NameIsDNS1035Label(o.PodName, false)
		if len(podNameValidationErrs) != 0 {
			errs = append(errs, fmt.Errorf("invalid pod name %q: %v", o.PodName, podNameValidationErrs))
		}
	}

	if _, err := os.Stat(o.SignalFilePath); err == nil {
		errs = append(errs, fmt.Errorf("signal file %q already exists", o.SignalFilePath))
	} else if !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("can't stat signal file %q: %w", o.SignalFilePath, err))
	}

	return utilerrors.NewAggregate(errs)
}

func (o *WaitOptions) Complete() error {
	err := o.ClientConfig.Complete()
	if err != nil {
		return err
	}

	err = o.InClusterReflection.Complete()
	if err != nil {
		return err
	}

	o.kubeClient, err = kubernetes.NewForConfig(o.ProtoConfig)
	if err != nil {
		return fmt.Errorf("can't build kubernetes clientset: %w", err)
	}

	for _, v := range o.Volumes {
		o.requiredAnnotations[fmt.Sprintf(naming.DelayedStorageMountedAnnotationFormat, v)] = naming.DelayedStorageMountedAnnotationTrue
	}

	return nil
}

func (o *WaitOptions) Run(originalStreams genericclioptions.IOStreams, cmd *cobra.Command) (returnErr error) {
	klog.Infof("%s version %s", cmd.Name(), version.Get())
	cliflag.PrintFlags(cmd.Flags())

	stopCh := signals.StopChannel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-stopCh
		cancel()
	}()

	return o.Execute(ctx, originalStreams, cmd)
}

func (o *WaitOptions) Execute(cmdCtx context.Context, originalStreams genericclioptions.IOStreams, cmd *cobra.Command) error {
	identityKubeInformers := informers.NewSharedInformerFactoryWithOptions(
		o.kubeClient,
		12*time.Hour,
		informers.WithNamespace(o.Namespace),
		informers.WithTweakListOptions(
			func(options *metav1.ListOptions) {
				options.FieldSelector = fields.OneTermEqualSelector("metadata.name", o.PodName).String()
			},
		),
	)

	waitController, err := wait.NewController(
		o.Namespace,
		o.PodName,
		o.requiredAnnotations,
		o.SignalFilePath,
		o.kubeClient,
		identityKubeInformers.Core().V1().Pods(),
	)
	if err != nil {
		return fmt.Errorf("can't create wait controller: %w", err)
	}

	var wg sync.WaitGroup
	defer func() {
		klog.InfoS("Waiting for background tasks to finish")
		wg.Wait()
		klog.InfoS("Background tasks have finished")
	}()

	cmdCtx, taskCtxCancel := context.WithCancel(cmdCtx)
	defer taskCtxCancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		identityKubeInformers.Start(cmdCtx.Done())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		waitController.Run(cmdCtx)
	}()

	<-cmdCtx.Done()

	return nil
}
