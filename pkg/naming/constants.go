package naming

const (
	DriverName = "proxy.csi.scylladb.com"
)

const (
	DelayedStorageBackendPersistentVolumeClaimRefAnnotation    = "proxy.csi.scylladb.com/backend-pvc-ref"
	DelayedStorageProxyPersistentVolumeClaimRefAnnotation      = "proxy.csi.scylladb.com/proxy-pvc-ref"
	DelayedStoragePersistentVolumeClaimBindCompletedAnnotation = "proxy.csi.scylladb.com/delayed-pvc-bind-completed"

	DelayedStorageBackendPersistentVolumeClaimProtectionFinalizer = "proxy.csi.scylladb.com/backend-pvc-protection"

	DelayedStorageProxyVolumeAttachmentRefAnnotation = "proxy.csi.scylladb.com/proxy-va-ref"

	DelayedStorageMountedAnnotationFormat = "proxy.csi.scylladb.com/volume-mounted-%s"
	DelayedStorageMountedAnnotationTrue   = "true"
)
