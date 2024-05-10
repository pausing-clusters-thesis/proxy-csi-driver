package bindstate

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
	binds map[string]*VolumeBind
}

func NewPersistentStateManager(workspacePath string) (*PersistentStateManager, error) {
	var err error

	binds := make(map[string]*VolumeBind)

	err = filepath.WalkDir(workspacePath, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if fpath != workspacePath && d.IsDir() {
			return filepath.SkipDir
		}

		if path.Ext(fpath) == fmt.Sprintf(".%s", stateFileExtension) {
			volumeBind, err := parseStateFile(fpath)
			if err != nil {
				return fmt.Errorf("can't parse state file at %q: %w", fpath, err)
			}

			if !volumeBind.IsValid() {
				klog.Warningf("Ignoring state file at %q because it failed validity check", fpath)
				return nil
			}

			binds[volumeBind.ProxyVolumeHandle] = volumeBind
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

func (sm *PersistentStateManager) SaveVolumeBind(volumeBind *VolumeBind) error {
	klog.V(5).InfoS("PersistentStateManager.SaveVolumeBind", "volumeBind", volumeBind)
	var err error

	// TODO: we need to provide a mechanism for checking if proxy volume is already bound
	// In theory we should be protected from it by pvc<->pvc 1:1 binding though.

	bindPath := sm.getBindPath(volumeBind.ProxyVolumeHandle)
	f, err := os.Create(bindPath)
	if err != nil {
		return fmt.Errorf("can't create bind state file %q: %w", bindPath, err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(volumeBind)
	if err != nil {
		return fmt.Errorf("can't encode bind state file at %q: %w", bindPath, err)
	}

	sm.mutex.Lock()
	sm.binds[volumeBind.ProxyVolumeHandle] = volumeBind
	sm.mutex.Unlock()

	klog.V(5).InfoS("PersistentStateManager.SaveVolumeBind: saved bind state", "volumeBind", volumeBind)
	return nil
}

func (sm *PersistentStateManager) DeleteVolumeBind(proxyVolumeHandle string) error {
	klog.V(5).InfoS("PersistentStateManager.DeleteVolumeBind", "proxyVolumeHandle", proxyVolumeHandle)
	var err error

	bindPath := sm.getBindPath(proxyVolumeHandle)
	err = os.Remove(bindPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("can't remove bind state file at %q: %w", bindPath, err)
	}

	sm.mutex.Lock()
	delete(sm.binds, proxyVolumeHandle)
	sm.mutex.Unlock()

	klog.V(5).InfoS("PersistentStateManager.DeleteVolumeBind: bind deleted", "proxyVolumeHandle", proxyVolumeHandle)
	return nil
}

func (sm *PersistentStateManager) GetVolumeBind(proxyVolumeHandle string) (*VolumeBind, bool, error) {
	klog.V(5).InfoS("PersistentStateManager.GetVolumeBind", "proxyVolumeHandle", proxyVolumeHandle)

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	bind, ok := sm.binds[proxyVolumeHandle]
	if !ok {
		return nil, false, nil
	}

	return bind, true, nil
}

func (sm *PersistentStateManager) getBindPath(proxyVolumeHandle string) string {
	return path.Join(sm.workspacePath, fmt.Sprintf("%s.%s", proxyVolumeHandle, stateFileExtension))
}

func parseStateFile(path string) (*VolumeBind, error) {
	var err error

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("can't open file at %q: %w", path, err)
	}
	defer f.Close()

	bind := &VolumeBind{}
	err = json.NewDecoder(f).Decode(bind)
	if err != nil {
		return nil, fmt.Errorf("can't decode file at %q: %w", path, err)
	}

	return bind, nil
}
