package bindstate

type BackendVolumeSpec struct {
	VolumeHandle string `json:"volumeHandle"`
	DriverName   string `json:"driverName"`
}

type VolumeBind struct {
	ProxyVolumeHandle string            `json:"proxyVolumeHandle"`
	BackendVolumeSpec BackendVolumeSpec `json:"backendVolumeSpec"`
}

func (b *VolumeBind) IsValid() bool {
	return b.ProxyVolumeHandle != "" && b.BackendVolumeSpec.VolumeHandle != "" && b.BackendVolumeSpec.DriverName != ""
}

type StateManager interface {
	SaveVolumeBind(volumeBind *VolumeBind) error
	DeleteVolumeBind(proxyVolumeHandle string) error
	GetVolumeBind(proxyVolumeHandle string) (*VolumeBind, bool, error)
}
