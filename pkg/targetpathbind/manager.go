package targetpathbind

import (
	"fmt"
	"os"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/targetpathbind/state"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
)

type TargetPathBindManager struct {
	stateManager state.StateManager
}

func NewTargetPathBindManager(stateManager state.StateManager) *TargetPathBindManager {
	return &TargetPathBindManager{
		stateManager: stateManager,
	}
}

func (tpm *TargetPathBindManager) SetUpTargetPath(volumeID string, podUID string, targetPath string) error {
	klog.V(5).InfoS("TargetPathBindManager.SetUpTargetPath", "volumeID", volumeID, "targetPath", targetPath)
	var err error

	klog.V(5).InfoS("TargetPathBindManager.SetUpTargetPath: setting up path", "volumeID", volumeID, "targetPath", targetPath)
	err = tpm.setUpPath(targetPath)
	if err != nil {
		klog.V(5).InfoS("TargetPathBindManager.SetUpTargetPath: path setup failed", "volumeID", volumeID, "targetPath", targetPath)
		return fmt.Errorf("cannot set up path: %w", err)
	}

	bind := &state.TargetPathBind{
		VolumeID:   volumeID,
		PodUID:     podUID,
		TargetPath: targetPath,
	}
	err = tpm.stateManager.SaveTargetPathBind(bind)
	if err != nil {
		errs := []error{
			fmt.Errorf("can't save target path in state: %w", err),
		}

		err = tpm.tearDownPath(targetPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't tear down path %q: %w", targetPath, err))
		}

		return errors.NewAggregate(errs)
	}

	klog.V(5).InfoS("TargetPathBindManager.SetUpTargetPath: target path set up", "volumeID", volumeID, "targetPath", targetPath)
	return nil
}

func (tpm *TargetPathBindManager) TearDownTargetPath(targetPath string) error {
	klog.V(5).InfoS("TargetPathBindManager.TearDownTargetPath", "targetPath", targetPath)
	var err error

	klog.V(5).InfoS("TargetPathBindManager.TearDownTargetPath: tearing down path", "targetPath", targetPath)
	err = tpm.tearDownPath(targetPath)
	if err != nil {
		return fmt.Errorf("can't tear down path: %w", err)
	}

	err = tpm.stateManager.DeleteTargetPathBind(targetPath)
	if err != nil {
		return fmt.Errorf("can't delete mount stateManager: %w", err)
	}

	klog.V(5).InfoS("TargetPathBindManager.TearDownTargetPath: volume unmounted", "targetPath", targetPath)
	return nil
}

func (tpm *TargetPathBindManager) GetTargetPath(volumeID string, podUID string) (string, error) {
	klog.V(5).InfoS("TargetPathBindManager.GetTargetPathBind", "volumeID", volumeID, "podUID", podUID)
	var err error

	bind, err := tpm.stateManager.GetTargetPathBind(volumeID, podUID)
	if err != nil {
		return "", fmt.Errorf("can't get bind: %w", err)
	}

	return bind.TargetPath, nil
}

func (tpm *TargetPathBindManager) setUpPath(path string) error {
	var err error

	var perm os.FileMode = 0440
	err = os.Mkdir(path, perm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("can't create directory at %q: %w", path, err)
	}

	fileInfo, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("can't get file info for %q: %w", path, err)
	}

	if fileInfo.Mode().Perm() != perm.Perm() {
		err = os.Chmod(path, perm)
		if err != nil {
			return fmt.Errorf("can't change mode for %q: %w", path, err)
		}
	}

	return nil
}

func (tpm *TargetPathBindManager) tearDownPath(path string) error {
	var err error

	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("can't remove path %q: %w", path, err)
	}

	return nil
}
