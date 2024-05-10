// Copyright (c) 2024 ScyllaDB.

package proxy_csi_driver

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/version"
	"github.com/scylladb/scylla-operator/pkg/helpers"
	"github.com/scylladb/scylla-operator/pkg/signals"
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	utilstrings "k8s.io/utils/strings"
)

type WaitOptions struct {
	annotationsFilePath string

	volumes       []string
	readyFilePath string
	sleep         bool

	requiredAnnotations map[string]string
}

func NewWaitOptions(streams genericiooptions.IOStreams) *WaitOptions {
	return &WaitOptions{
		annotationsFilePath: "",

		volumes:       []string{},
		readyFilePath: "",
		sleep:         true,

		requiredAnnotations: map[string]string{},
	}
}

func NewWaitCommand(streams genericiooptions.IOStreams) *cobra.Command {
	o := NewWaitOptions(streams)

	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for delayed volumes.",
		Long:  "Wait for all delayed volumes used by the host pod to be mounted.",
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

	cmd.Flags().StringVar(&o.annotationsFilePath, "annotations-file", o.annotationsFilePath, "Path to the annotations file mounted with downward API.")
	cmd.Flags().StringSliceVar(&o.volumes, "volume", o.volumes, "Volume(s) to wait for.")
	cmd.Flags().StringVar(&o.readyFilePath, "ready-file-path", o.readyFilePath, "Ready file path.")
	cmd.Flags().BoolVar(&o.sleep, "sleep", o.sleep, "Sleep signifies whether the command should keep running after the annotation is set.")

	return cmd
}

func (o *WaitOptions) Validate() error {
	var errs []error

	if len(o.annotationsFilePath) == 0 {
		errs = append(errs, fmt.Errorf("annotations-file must be set"))
	}

	if info, err := os.Stat(o.annotationsFilePath); err != nil {
		errs = append(errs, fmt.Errorf("can't stat annotations file: %w", err))
	} else {
		if info.IsDir() {
			errs = append(errs, fmt.Errorf("annotations file must not be a directory"))
		}
	}

	if _, err := os.Stat(o.readyFilePath); err == nil {
		errs = append(errs, fmt.Errorf("ready file %q already exists", o.readyFilePath))
	} else if !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("can't stat ready file %q: %w", o.readyFilePath, err))
	}

	return utilerrors.NewAggregate(errs)
}

func (o *WaitOptions) Complete() error {
	for _, v := range o.volumes {
		o.requiredAnnotations[fmt.Sprintf(naming.DelayedStorageMountedAnnotationFormat, v)] = fmt.Sprintf("%q", utilstrings.EscapeQualifiedName(naming.DelayedStorageMountedAnnotationTrue))
	}

	return nil
}

func (o *WaitOptions) Run(originalStreams genericiooptions.IOStreams, cmd *cobra.Command) (returnErr error) {
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

func (o *WaitOptions) Execute(ctx context.Context, originalStreams genericiooptions.IOStreams, cmd *cobra.Command) error {
	err := o.waitForAnnotations(ctx)
	if err != nil {
		return fmt.Errorf("can't wait for annotations: %w", err)
	}

	klog.V(2).InfoS("Touching ready file", "path", o.readyFilePath)
	err = helpers.TouchFile(o.readyFilePath)
	if err != nil {
		return fmt.Errorf("can't touch ready file at %q: %w", o.readyFilePath, err)
	}
	klog.V(2).InfoS("Touched ready file", "path", o.readyFilePath)

	if !o.sleep {
		return nil
	}

	<-ctx.Done()

	return nil
}

func (o *WaitOptions) waitForAnnotations(ctx context.Context) error {
	klog.V(2).InfoS("Awaiting annotations to be set", "annotations", o.requiredAnnotations)

	// K8s' downwards API uses symlinks, so we have to watch the entire directory to pick up renames or atomic saves.
	annotationsFile := filepath.Clean(o.annotationsFilePath)
	annotationsDir, _ := filepath.Split(annotationsFile)
	evaluatedAnnotationsFile, _ := filepath.EvalSymlinks(annotationsFile)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("can't create new watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(annotationsDir)
	if err != nil {
		return fmt.Errorf("can't watch annotations directory: %w", err)
	}

	oneTime := make(chan struct{}, 1)
	// Make sure the changes are picked up in case the condition was met before we started the watcher.
	oneTime <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-oneTime:
			annotationsMatch, err := o.checkAnnotations()
			if err != nil {
				return fmt.Errorf("can't check annotations: %w", err)
			}
			if annotationsMatch {
				klog.V(2).InfoS("Annotations match, finishing waiting", "annotations", o.requiredAnnotations)

				return nil
			}

		case event := <-watcher.Events:
			currAnnotationsFile, _ := filepath.EvalSymlinks(o.annotationsFilePath)

			if (filepath.Clean(event.Name) == annotationsFile && (event.Has(fsnotify.Write) || event.Has(fsnotify.Create))) ||
				(len(currAnnotationsFile) != 0 && currAnnotationsFile != evaluatedAnnotationsFile) {
				evaluatedAnnotationsFile = currAnnotationsFile

				annotationsMatch, err := o.checkAnnotations()
				if err != nil {
					return fmt.Errorf("can't check annotations: %w", err)
				}
				if annotationsMatch {
					klog.V(2).InfoS("Annotations match, finishing waiting", "annotations", o.requiredAnnotations)

					return nil
				}

				break
			}

			if filepath.Clean(event.Name) == annotationsFile && event.Has(fsnotify.Remove) {
				return fmt.Errorf("watched file was removed")
			}

		case err := <-watcher.Errors:
			return fmt.Errorf("received watcher error: %w", err)

		}
	}
}

func (o *WaitOptions) checkAnnotations() (bool, error) {
	f, err := os.Open(o.annotationsFilePath)
	if err != nil {
		return false, fmt.Errorf("can't read annotations file: %w", err)
	}
	defer f.Close()

	annotations := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		key, value, found := strings.Cut(l, "=")
		if !found {
			return false, fmt.Errorf("invalid annotation: %q", l)
		}
		annotations[key] = value
	}
	err = scanner.Err()
	if err != nil {
		return false, fmt.Errorf("can't scan annotations file: %w", err)
	}

	for requiredKey, requiredVal := range o.requiredAnnotations {
		val, ok := annotations[requiredKey]
		if !ok || val != requiredVal {
			return false, nil
		}
	}

	return true, nil
}
