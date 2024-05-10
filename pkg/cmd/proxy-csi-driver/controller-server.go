package proxy_csi_driver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	backendpvc "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/attachdetach/controller/persistentvolumeclaim/backend"
	backendpvcprotection "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/attachdetach/controller/persistentvolumeclaim/backend/protection"
	proxypvc "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/attachdetach/controller/persistentvolumeclaim/proxy"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/attachdetach/controller/pod"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/driver"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/version"
	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	sogenericclioptions "github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"github.com/scylladb/scylla-operator/pkg/signals"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

type ControllerServerOptions struct {
	// TODO: leader election: how to sync with external-provisioner?

	sogenericclioptions.ClientConfig

	kubeClient kubernetes.Interface

	csiAddress string

	concurrentSyncs int
}

func NewDefaultControllerServer() *ControllerServerOptions {
	return &ControllerServerOptions{
		csiAddress: "",

		concurrentSyncs: defaultConcurrentSyncs,
	}
}

func NewControllerServerCommand(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewDefaultControllerServer()

	cmd := &cobra.Command{
		Use: "controller-server",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := version.Get()
			klog.InfoS("Binary info",
				"Command", cmd.Name(),
				"Revision", version.OptionalToString(v.Revision),
				"RevisionTime", version.OptionalToString(v.RevisionTime),
				"Modified", version.OptionalToString(v.Modified),
				"GoVersion", version.OptionalToString(v.GoVersion),
			)
			cliflag.PrintFlags(cmd.Flags())

			stopCh := signals.StopChannel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				<-stopCh
				cancel()
			}()

			err := o.Validate()
			if err != nil {
				return err
			}

			err = o.Complete()
			if err != nil {
				return err
			}

			err = o.Run(ctx, streams, cmd)
			if err != nil {
				return err
			}

			return nil
		},

		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().StringVarP(&o.csiAddress, "csi-address", "", o.csiAddress, "Path to the CSI endpoint socket.")
	cmd.Flags().IntVarP(&o.concurrentSyncs, "concurrent-syncs", "", o.concurrentSyncs, "The number of DelayedVolume objects that are allowed to sync concurrently.")

	return cmd
}

func (o *ControllerServerOptions) Validate() error {
	var err error
	var errs []error

	err = o.ClientConfig.Validate()
	if err != nil {
		errs = append(errs, err)
	}

	if len(o.csiAddress) == 0 {
		errs = append(errs, fmt.Errorf("csi-address must not be empty"))
	}

	err = utilerrors.NewAggregate(errs)
	if err != nil {
		return err
	}

	return nil
}

func (o *ControllerServerOptions) Complete() error {
	err := o.ClientConfig.Complete()
	if err != nil {
		return err
	}

	o.kubeClient, err = kubernetes.NewForConfig(o.ProtoConfig)
	if err != nil {
		return fmt.Errorf("can't build kubernetes clientset: %w", err)
	}

	return nil
}

func (o *ControllerServerOptions) Run(ctx context.Context, streams genericclioptions.IOStreams, cmd *cobra.Command) error {
	return o.run(ctx, streams)
}

func (o *ControllerServerOptions) run(ctx context.Context, streams genericclioptions.IOStreams) error {
	// Attacher-detacher
	kubeInformers := informers.NewSharedInformerFactory(o.kubeClient, resyncPeriod)

	pc, err := pod.NewController(
		o.kubeClient,
		kubeInformers.Storage().V1().VolumeAttachments(),
		kubeInformers.Storage().V1().CSIDrivers(),
		kubeInformers.Core().V1().Pods(),
		kubeInformers.Core().V1().PersistentVolumeClaims(),
		kubeInformers.Core().V1().PersistentVolumes(),
	)
	if err != nil {
		return fmt.Errorf("can't create pod controller: %w", err)
	}

	ppvcc, err := proxypvc.NewController(
		o.kubeClient,
		kubeInformers.Core().V1().PersistentVolumeClaims(),
	)
	if err != nil {
		return fmt.Errorf("can't create proxy persistent volume claim controller: %w", err)
	}

	bpvcc, err := backendpvc.NewController(
		o.kubeClient,
		kubeInformers.Core().V1().PersistentVolumeClaims(),
	)
	if err != nil {
		return fmt.Errorf("can't create backend persistent volume claim controller: %w", err)
	}

	bpvcpc, err := backendpvcprotection.NewController(
		o.kubeClient,
		kubeInformers.Core().V1().PersistentVolumeClaims(),
	)
	if err != nil {
		return fmt.Errorf("can't create backend persistent volume claim protection controller: %w", err)
	}

	// Server
	if err := os.Remove(o.csiAddress); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("can't remove file at %q: %w", o.csiAddress, err)
	}

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "unix", o.csiAddress)
	if err != nil {
		return fmt.Errorf("can't listen on %q using unix protocol: %w", o.csiAddress, err)
	}

	defer func() {
		cleanupErr := os.Remove(o.csiAddress)
		if cleanupErr != nil {
			klog.ErrorS(cleanupErr, "Failed to clean up socket", "path", o.csiAddress)
		}
	}()

	cs := driver.NewControllerServer(version.Get().String())

	server := grpc.NewServer()

	csi.RegisterIdentityServer(server, cs)
	csi.RegisterControllerServer(server, cs)

	var eg errgroup.Group

	eg.Go(func() error {
		kubeInformers.Start(ctx.Done())

		return nil
	})

	eg.Go(func() error {
		pc.Run(ctx, o.concurrentSyncs)

		return nil
	})

	eg.Go(func() error {
		ppvcc.Run(ctx, o.concurrentSyncs)

		return nil
	})

	eg.Go(func() error {
		bpvcc.Run(ctx, o.concurrentSyncs)

		return nil
	})

	eg.Go(func() error {
		bpvcpc.Run(ctx, o.concurrentSyncs)

		return nil
	})

	eg.Go(func() error {
		klog.InfoS("Listening for connections", "address", listener.Addr())
		err = server.Serve(listener)
		if err != nil && !errors.Is(grpc.ErrServerStopped, err) {
			return fmt.Errorf("can't serve: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()

		server.GracefulStop()

		return nil
	})

	err = eg.Wait()
	if err != nil {
		return err
	}

	return nil
}
