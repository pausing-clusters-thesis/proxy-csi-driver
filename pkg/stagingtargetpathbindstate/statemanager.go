package stagingtargetpathbindstate

type StateManager interface {
	BindStagingTargetPath(volumeID, stagingTargetPath string) error
	UnbindStagingTargetPath(volumeID string) error
	GetBoundStagingTargetPath(volumeID string) (string, error)
}
