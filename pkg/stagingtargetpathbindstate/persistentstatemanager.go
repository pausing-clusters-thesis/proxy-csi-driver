package stagingtargetpathbindstate

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"

	"k8s.io/klog/v2"
)

const (
	stateFileExtension = "json"
)

type stagingTargetPathBind struct {
	VolumeID          string `json:"volumeID"`
	StagingTargetPath string `json:"stagingTargetPath"`
}

func (b *stagingTargetPathBind) IsValid() bool {
	return len(b.VolumeID) != 0 && len(b.StagingTargetPath) != 0
}

type PersistentStateManager struct {
	workspacePath string

	mutex sync.RWMutex
	binds map[string]*stagingTargetPathBind
}

func NewPersistentStateManager(workspacePath string) (*PersistentStateManager, error) {
	var err error

	binds := make(map[string]*stagingTargetPathBind)

	err = filepath.WalkDir(workspacePath, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if fpath != workspacePath && d.IsDir() {
			return filepath.SkipDir
		}

		if path.Ext(fpath) == fmt.Sprintf(".%s", stateFileExtension) {
			stateFile, err := parseStateFile(fpath)
			if err != nil {
				return fmt.Errorf("can't parse state file at %q: %w", fpath, err)
			}

			if !stateFile.IsValid() {
				klog.Warningf("Ignoring state file at %q because it failed validity check", fpath)
				return nil
			}

			binds[stateFile.VolumeID] = stateFile
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't recreate state from %q: %w", workspacePath, err)
	}

	sm := &PersistentStateManager{
		workspacePath: workspacePath,

		mutex: sync.RWMutex{},
		binds: binds,
	}

	return sm, nil
}

var _ StateManager = &PersistentStateManager{}

func (sm *PersistentStateManager) BindStagingTargetPath(volumeID, stagingTargetPath string) error {
	klog.V(5).InfoS("PersistentStateManager.SaveStagingTargetPathBind", "volumeID", volumeID, "stagingTargetPath", stagingTargetPath)
	var err error

	bindPath := sm.getBindPath(volumeID)
	f, err := os.Create(bindPath)
	if err != nil {
		return fmt.Errorf("can't create bind state file %q: %w", bindPath, err)
	}
	defer f.Close()

	bind := &stagingTargetPathBind{
		VolumeID:          volumeID,
		StagingTargetPath: stagingTargetPath,
	}
	err = json.NewEncoder(f).Encode(bind)
	if err != nil {
		return fmt.Errorf("can't encode bind state file at %q: %w", bindPath, err)
	}

	sm.mutex.Lock()
	sm.binds[volumeID] = bind
	sm.mutex.Unlock()

	klog.V(5).InfoS("PersistentStateManager.SaveStagingTargetPathBind: saved bind state", "volumeID", volumeID, "stagingTargetPath", stagingTargetPath)
	return nil
}

func (sm *PersistentStateManager) UnbindStagingTargetPath(volumeID string) error {
	klog.V(5).InfoS("PersistentStateManager.UnbindStagingTargetPath", "volumeID", volumeID)
	var err error

	bindPath := sm.getBindPath(volumeID)
	err = os.Remove(bindPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("can't remove bind state file at %q: %w", bindPath, err)
	}

	sm.mutex.Lock()
	delete(sm.binds, volumeID)
	sm.mutex.Unlock()

	klog.V(5).InfoS("PersistentStateManager.UnbindStagingTargetPath: staging target path unbound", "volumeID", volumeID)
	return nil
}

func (sm *PersistentStateManager) GetBoundStagingTargetPath(volumeID string) (string, error) {
	klog.V(5).InfoS("PersistentStateManager.GetStagingTargetPathBind", "volumeID", volumeID)

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	bind, ok := sm.binds[volumeID]
	if !ok {
		return "", fmt.Errorf("can't get bound staging target path")
	}

	return bind.StagingTargetPath, nil
}

func (sm *PersistentStateManager) getBindPath(volumeID string) string {
	return path.Join(sm.workspacePath, fmt.Sprintf("%s.%s", volumeID, stateFileExtension))
}

func parseStateFile(path string) (*stagingTargetPathBind, error) {
	var err error

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("can't open file at %q: %w", path, err)
	}
	defer f.Close()

	bind := &stagingTargetPathBind{}
	err = json.NewDecoder(f).Decode(bind)
	if err != nil {
		return nil, fmt.Errorf("can't decode file at %q: %w", path, err)
	}

	return bind, nil
}
