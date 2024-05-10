package proxy_csi_driver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/driver"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/stagingtargetpathbindstate"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/targetpathbind"
	targetpathbindstate "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/targetpathbind/state"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/version"
	volumebindstate "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volume/bindstate"
	proxycsi "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volume/csi"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volumemanager/controller/pod"
	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	sogenericclioptions "github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"github.com/scylladb/scylla-operator/pkg/signals"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

// TODO: move to common place or move servers to dedicated packages
const (
	defaultConcurrentSyncs = 50
	resyncPeriod           = 12 * time.Hour
)

type NodeServerOptions struct {
	sogenericclioptions.ClientConfig

	kubeClient kubernetes.Interface

	csiAddress        string
	backendCSIAddress []string
	nodeName          string
	stateDir          string

	stagingTargetPathBindStateManagerDir string
	targetPathBindStateManagerDir        string
	volumeBindStateManagerDir            string

	concurrentSyncs int
}

func NewDefaultNodeServer() *NodeServerOptions {
	return &NodeServerOptions{
		csiAddress:        "",
		backendCSIAddress: []string{},
		nodeName:          "",
		stateDir:          "",

		stagingTargetPathBindStateManagerDir: "",
		targetPathBindStateManagerDir:        "",

		concurrentSyncs: defaultConcurrentSyncs,
	}
}

func NewNodeServerCommand(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewDefaultNodeServer()

	cmd := &cobra.Command{
		Use: "node-server",
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
	cmd.Flags().StringSliceVarP(&o.backendCSIAddress, "backend-csi-address", "", o.backendCSIAddress, "Path to the CSI endpoint of supported backend CSI driver socket.")
	cmd.Flags().StringVarP(&o.nodeName, "node-name", "", o.nodeName, "Name of the node the server is running on.")
	cmd.Flags().StringVarP(&o.stateDir, "state-dir", "", o.stateDir, "Path to a directory where local state is kept.")
	cmd.Flags().IntVarP(&o.concurrentSyncs, "concurrent-syncs", "", o.concurrentSyncs, "The number of DelayedVolume objects that are allowed to sync concurrently.")

	return cmd
}

func (o *NodeServerOptions) Validate() error {
	var err error
	var errs []error

	err = o.ClientConfig.Validate()
	if err != nil {
		errs = append(errs, err)
	}

	if len(o.csiAddress) == 0 {
		errs = append(errs, fmt.Errorf("csi-address must not be empty"))
	}

	for _, addr := range o.backendCSIAddress {
		_, err = os.Stat(addr)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't stat backend CSI address %q: %w", addr, err))
		}
	}

	if len(o.nodeName) == 0 {
		errs = append(errs, fmt.Errorf("node-name must not be empty"))
	}

	if len(o.stateDir) == 0 {
		errs = append(errs, fmt.Errorf("state-dir must not be empty"))
	}

	_, err = os.Stat(o.stateDir)
	if err != nil {
		errs = append(errs, fmt.Errorf("can't stat state-dir: %w", err))
	}

	err = utilerrors.NewAggregate(errs)
	if err != nil {
		return err
	}

	return nil
}

func (o *NodeServerOptions) Complete() error {
	err := o.ClientConfig.Complete()
	if err != nil {
		return err
	}

	o.kubeClient, err = kubernetes.NewForConfig(o.ProtoConfig)
	if err != nil {
		return fmt.Errorf("can't build kubernetes clientset: %w", err)
	}

	o.stagingTargetPathBindStateManagerDir = path.Join(o.stateDir, "stagingtargetpath")
	err = os.Mkdir(o.stagingTargetPathBindStateManagerDir, 0777)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("can't make staging target path bind state manager dir: %w", err)
	}

	o.targetPathBindStateManagerDir = path.Join(o.stateDir, "targetpath")
	err = os.Mkdir(o.targetPathBindStateManagerDir, 0777)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("can't make target path bind state manager dir: %w", err)
	}

	o.volumeBindStateManagerDir = path.Join(o.stateDir, "bind")
	err = os.Mkdir(o.volumeBindStateManagerDir, 0777)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("can't make volume bind state manager dir: %w", err)
	}

	return nil
}

func (o *NodeServerOptions) Run(ctx context.Context, streams genericclioptions.IOStreams, cmd *cobra.Command) error {
	return o.run(ctx, streams)
}

func (o *NodeServerOptions) run(ctx context.Context, streams genericclioptions.IOStreams) error {
	var err error

	// Common
	stagingTargetPathBindStateManager, err := stagingtargetpathbindstate.NewPersistentStateManager(o.stagingTargetPathBindStateManagerDir)
	if err != nil {
		return fmt.Errorf("can't create staging target path bind state manager: %w", err)
	}

	targetPathBindStateManager, err := targetpathbindstate.NewPersistentStateManager(o.targetPathBindStateManagerDir)
	if err != nil {
		return fmt.Errorf("can't create persistent state manager: %w", err)
	}

	targetPathBindManager := targetpathbind.NewTargetPathBindManager(targetPathBindStateManager)

	volumeBindStateManager, err := volumebindstate.NewPersistentStateManager(o.volumeBindStateManagerDir)
	if err != nil {
		return fmt.Errorf("can't create persistent volume bind state manager: %w", err)
	}

	driverRegistry := proxycsi.NewDriverRegistry()

	var errs []error
	for _, addr := range o.backendCSIAddress {
		err = driverRegistry.RegisterDriver(ctx, addr)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't register CSI driver with address %q: %w", addr, err))
		}
	}

	err = utilerrors.NewAggregate(errs)
	if err != nil {
		return err
	}

	// Volume manager
	kubeInformers := informers.NewSharedInformerFactory(o.kubeClient, resyncPeriod)
	hostInformers := informers.NewSharedInformerFactoryWithOptions(o.kubeClient, resyncPeriod, informers.WithTweakListOptions(
		func(options *metav1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector("spec.nodeName", o.nodeName).String()
		},
	))

	// TODO: rename package
	pc, err := pod.NewController(
		o.kubeClient,
		kubeInformers.Core().V1().Nodes(),
		kubeInformers.Storage().V1().VolumeAttachments(),
		kubeInformers.Storage().V1().CSIDrivers(),
		hostInformers.Core().V1().Pods(),
		kubeInformers.Core().V1().PersistentVolumeClaims(),
		kubeInformers.Core().V1().PersistentVolumes(),
		o.nodeName,
		stagingTargetPathBindStateManager,
		targetPathBindManager,
		volumeBindStateManager,
		driverRegistry,
	)
	if err != nil {
		return fmt.Errorf("can't create pod controller: %w", err)
	}

	// Server
	err = os.Remove(o.csiAddress)
	if err != nil && !os.IsNotExist(err) {
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

	d := driver.NewNodeServer(
		version.Get().String(),
		o.nodeName,
		stagingTargetPathBindStateManager,
		targetPathBindManager,
		volumeBindStateManager,
		driverRegistry,
	)

	server := grpc.NewServer()

	csi.RegisterIdentityServer(server, d)
	csi.RegisterNodeServer(server, d)

	var eg errgroup.Group

	eg.Go(func() error {
		kubeInformers.Start(ctx.Done())

		return nil
	})

	eg.Go(func() error {
		hostInformers.Start(ctx.Done())

		return nil
	})

	eg.Go(func() error {
		pc.Run(ctx, o.concurrentSyncs)

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
