package state

import "fmt"

type TargetPathBind struct {
	VolumeID   string `json:"volumeID"`
	PodUID     string `json:"podUID"`
	TargetPath string `json:"targetPath"`
}

func (b *TargetPathBind) GetID() string {
	return volumeIDAndPodUIDToTargetPathBindID(b.VolumeID, b.PodUID)
}

func volumeIDAndPodUIDToTargetPathBindID(volumeID string, podUID string) string {
	return fmt.Sprintf("%s-%s", volumeID, podUID)
}

func (b *TargetPathBind) IsValid() bool {
	return len(b.VolumeID) != 0 && len(b.PodUID) != 0 && len(b.TargetPath) != 0
}

type StateManager interface {
	SaveTargetPathBind(*TargetPathBind) error
	DeleteTargetPathBind(targetPath string) error
	GetTargetPathBind(volumeID, podUID string) (*TargetPathBind, error)
}
