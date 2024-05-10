package state

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
	//MetadataFileMaxSize      = 4 * 1024
)

type PersistentStateManager struct {
	workspacePath string

	mutex sync.RWMutex
	// TODO: think about improving this
	binds                        map[string]*TargetPathBind
	targetPathToTargetPathBindID map[string]string
}

func NewPersistentStateManager(workspacePath string) (*PersistentStateManager, error) {
	var err error

	binds := make(map[string]*TargetPathBind)
	targetPathToMountStateID := make(map[string]string)

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

			binds[stateFile.GetID()] = stateFile
			targetPathToMountStateID[stateFile.TargetPath] = stateFile.GetID()
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't recreate state from %q: %w", workspacePath, err)
	}

	sm := &PersistentStateManager{
		workspacePath: workspacePath,

		mutex:                        sync.RWMutex{},
		binds:                        binds,
		targetPathToTargetPathBindID: targetPathToMountStateID,
	}

	return sm, nil
}

var _ StateManager = &PersistentStateManager{}

func (sm *PersistentStateManager) SaveTargetPathBind(targetPathBind *TargetPathBind) error {
	klog.V(5).InfoS("PersistentStateManager.SaveTargetPathBind", "targetPathBind", targetPathBind)
	var err error

	bindPath := sm.getBindPath(targetPathBind.VolumeID, targetPathBind.PodUID)
	f, err := os.Create(bindPath)
	if err != nil {
		return fmt.Errorf("can't create bind state file %q: %w", bindPath, err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(targetPathBind)
	if err != nil {
		return fmt.Errorf("can't encode bind state file at %q: %w", bindPath, err)
	}

	sm.mutex.Lock()
	bindID := targetPathBind.GetID()
	sm.binds[bindID] = targetPathBind
	sm.targetPathToTargetPathBindID[targetPathBind.TargetPath] = bindID
	sm.mutex.Unlock()

	klog.V(5).InfoS("PersistentStateManager.SaveTargetPathBind: saved bind state", "targetPathBind", targetPathBind)
	return nil
}

func (sm *PersistentStateManager) DeleteTargetPathBind(targetPath string) error {
	klog.V(5).InfoS("PersistentStateManager.DeleteTargetPathBind", "targetPath", targetPath)
	var err error

	sm.mutex.RLock()
	bindID, ok := sm.targetPathToTargetPathBindID[targetPath]
	if !ok {
		// Nothing to do.
		sm.mutex.RUnlock()
		return nil
	}
	bind, ok := sm.binds[bindID]
	if !ok {
		// Nothing to do.
		sm.mutex.RUnlock()
		return nil
	}
	sm.mutex.RUnlock()

	bindPath := sm.getBindPath(bind.VolumeID, bind.PodUID)
	err = os.Remove(bindPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("can't remove bind state file at %q: %w", bindPath, err)
	}

	sm.mutex.Lock()
	delete(sm.targetPathToTargetPathBindID, targetPath)
	delete(sm.binds, bindID)
	sm.mutex.Unlock()

	klog.V(5).InfoS("PersistentStateManager.DeleteTargetPathBind: bind deleted", "targetPath", targetPath, "volumeID", bind.VolumeID, "podUID", bind.PodUID)
	return nil
}

func (sm *PersistentStateManager) GetTargetPathBind(volumeID string, podUID string) (*TargetPathBind, error) {
	klog.V(5).InfoS("PersistentStateManager.GetTargetPathBind", "volumeID", volumeID, "podUID", podUID)

	bindID := volumeIDAndPodUIDToTargetPathBindID(volumeID, podUID)

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	bind, ok := sm.binds[bindID]
	if !ok {
		return nil, fmt.Errorf("can't get bind")
	}

	return bind, nil
}

func (sm *PersistentStateManager) getBindPath(volumeID string, podUID string) string {
	return path.Join(sm.workspacePath, fmt.Sprintf("%s-%s.%s", volumeID, podUID, stateFileExtension))
}

func parseStateFile(path string) (*TargetPathBind, error) {
	var err error

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("can't open file at %q: %w", path, err)
	}
	defer f.Close()

	bind := &TargetPathBind{}
	err = json.NewDecoder(f).Decode(bind)
	if err != nil {
		return nil, fmt.Errorf("can't decode file at %q: %w", path, err)
	}

	return bind, nil
}
