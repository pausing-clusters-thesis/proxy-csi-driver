package pod

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/controllerhelpers"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/helpers/maps"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/helpers/slices"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	k8svolume "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/thirdparty/k8s.io/kubernetes/pkg/volume"
	k8svolumeutil "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/thirdparty/k8s.io/kubernetes/pkg/volume/util"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volume"
	volumebindstate "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volume/bindstate"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/volume/csi"
	sonaming "github.com/scylladb/scylla-operator/pkg/naming"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	storagehelpers "k8s.io/component-helpers/storage/volume"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

func (pc *Controller) syncVolume(ctx context.Context, key string, spec *volume.Spec) error {
	now := time.Now()
	klog.V(4).InfoS("Started syncing Volume", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim), "startTime", now)
	defer func() {
		klog.V(4).InfoS("Finished syncing Volume", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim), "duration", time.Since(now))
	}()

	if spec.ProxyPersistentVolume.Spec.CSI == nil {
		klog.ErrorS(nil, "Proxy PersistentVolumeSource unsupported, only CSIPersistentVolumeSource is supported", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim))
		return nil
	}

	if spec.ProxyPersistentVolume.Spec.CSI.Driver != naming.DriverName {
		klog.ErrorS(nil, "Proxy PersistentVolumeSource is not provided by us", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim))
		return nil
	}

	targetPath, err := pc.targetPathBindManager.GetTargetPath(spec.ProxyPersistentVolume.Spec.CSI.VolumeHandle, string(spec.Pod.GetUID()))
	if err != nil {
		return fmt.Errorf("can't get target path: %w", err)
	}

	if spec.BackendPersistentVolume.Spec.CSI == nil {
		klog.ErrorS(nil, "Backend PersistentVolumeSource unsupported, only CSIPersistentVolumeSource is supported", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim))
		return nil
	}

	// TODO: we should have an abstraction for all csi/driver-registry/csi-driver related functions

	csiDriver, err := pc.csiDriverLister.Get(spec.BackendPersistentVolume.Spec.CSI.Driver)
	if err != nil {
		return fmt.Errorf("can't get CSIDriver %q: %w", spec.BackendPersistentVolume.Spec.CSI.Driver, err)
	}

	backendVolumeMode, err := getVolumeMode(spec.BackendPersistentVolume)
	if err != nil {
		return fmt.Errorf("can't get volume mode of PersistentVolume %q: %w", sonaming.ObjRef(spec.BackendPersistentVolume), err)
	}

	if backendVolumeMode != corev1.PersistentVolumeFilesystem {
		return fmt.Errorf("unsupported persistent volume mode: %q", backendVolumeMode)
	}

	// TODO: move to plugin abstraction?

	node, err := pc.nodeLister.Get(pc.nodeName)
	if err != nil {
		return fmt.Errorf("can't get Node %q: %w", pc.nodeName, err)
	}

	err = storagehelpers.CheckNodeAffinity(spec.BackendPersistentVolume, node.Labels)
	if err != nil {
		return fmt.Errorf("volume does not have affinity corresponding to the node: %w", err)
	}

	// TODO: volumemanager creates a mounter abstraction here

	// TODO: checkMountOptionSupport

	// TODO: check if volume is mounted elsewhere?

	// TODO: volumemanager creates an attacher abstraction here

	// TODO: volumemanager creates a volumeDeviceMounter abstraction here

	if csiDriver.Spec.AttachRequired != nil && *csiDriver.Spec.AttachRequired {
		volumeAttachmentName := getVolumeAttachmentName(spec.BackendPersistentVolume.Spec.CSI.VolumeHandle, csiDriver.Name, pc.nodeName)
		attached, err := pc.isVolumeAttached(volumeAttachmentName)
		if err != nil {
			return fmt.Errorf("can't verify volume attachment: %w", err)
		}
		if !attached {
			klog.InfoS("Waiting for volume to attach, will try again later", "Volume", spec.Volume.Name, "Pod", klog.KObj(spec.Pod), "ProxyPersistentVolumeClaim", klog.KObj(spec.ProxyPersistentVolumeClaim), "BackendPersistentVolumeClaim", klog.KObj(spec.BackendPersistentVolumeClaim), "volumeAttachmentName", volumeAttachmentName)
			// FIXME: fix handlers instead
			pc.queue.Add(key)
			return nil
		}
	}

	// FIXME: check if device is globally mounted? We'd need some cache.
	// FIXME: maybe we can use the state manager for everything?
	// FIXME: OR we can just rely on idempotency

	volumeBind := &volumebindstate.VolumeBind{
		ProxyVolumeHandle: spec.ProxyPersistentVolume.Spec.CSI.VolumeHandle,
		BackendVolumeSpec: volumebindstate.BackendVolumeSpec{
			VolumeHandle: spec.BackendPersistentVolume.Spec.CSI.VolumeHandle,
			DriverName:   spec.BackendPersistentVolume.Spec.CSI.Driver,
		},
	}
	err = pc.volumeBindStateManager.SaveVolumeBind(volumeBind)
	if err != nil {
		return fmt.Errorf("can't save volume bind state: %w", err)
	}

	var fsGroup *int64
	var fsGroupChangePolicy *corev1.PodFSGroupChangePolicy
	podSecurityContext := spec.Pod.Spec.SecurityContext
	if podSecurityContext != nil {
		if podSecurityContext.FSGroup != nil {
			fsGroup = podSecurityContext.FSGroup
		}
		if podSecurityContext.FSGroupChangePolicy != nil {
			fsGroupChangePolicy = podSecurityContext.FSGroupChangePolicy
		}
	}

	deviceMounterArgs := volume.DeviceMounterArgs{
		FsGroup: fsGroup,
		// TODO: SELinuxLabel
	}
	err = pc.mountDevice(
		ctx,
		spec,
		deviceMounterArgs,
	)
	if err != nil {
		// TODO: generate event
		// TODO: mark device error state
		// TODO: other DVs referencing this device shouldn't attempt to mount it globally again?

		// TODO: add operation context to error?
		return fmt.Errorf("can't mount device: %w", err)
	}

	// TODO: mark device as mounted in state

	// TODO: support volume resize?

	// TODO: mount options from volume?

	// TODO: add mounter args
	mounterArgs := volume.MounterArgs{
		FsUser:              k8svolumeutil.FsUserFrom(spec.Pod),
		FsGroup:             fsGroup,
		FSGroupChangePolicy: fsGroupChangePolicy,
		// TODO: where should this come from?
		DesiredSize: nil,
		// TODO: selinux labels
	}

	err = pc.setUpAt(ctx, spec, targetPath, mounterArgs)
	if err != nil {
		return fmt.Errorf("can't set up volume: %w", err)
	}

	patch, err := controllerhelpers.PrepareSetAnnotationPatch(spec.Pod, fmt.Sprintf(naming.DelayedStorageMountedAnnotationFormat, spec.Volume.Name), ptr.To(naming.DelayedStorageMountedAnnotationTrue))
	if err != nil {
		return fmt.Errorf("can't prepare set annotation patch: %w", err)
	}

	_, err = pc.kubeClient.CoreV1().Pods(spec.Pod.Namespace).Patch(ctx, spec.Pod.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("can't patch pod %q: %w", naming.ObjRef(spec.Pod), err)
	}

	return nil
}

// TODO: move to volume util
func getVolumeMode(pv *corev1.PersistentVolume) (corev1.PersistentVolumeMode, error) {
	if pv.Spec.VolumeMode == nil {
		return "", fmt.Errorf("nil volume mode")
	}

	return *pv.Spec.VolumeMode, nil
}

// FIXME: this is copied from k8s, need to investigate whether this getting out of sync will cause problems?
// FIXME: perhaps we could list all volume attachments referring this volume?
// getVolumeAttachmentName returns csi-<sha256(volName,csiDriverName,NodeName)>
func getVolumeAttachmentName(volName, csiDriverName, nodeName string) string {
	result := sha256.Sum256([]byte(fmt.Sprintf("%s%s%s", volName, csiDriverName, nodeName)))
	return fmt.Sprintf("csi-%x", result)
}

func (pc *Controller) isVolumeAttached(volumeAttachmentName string) (bool, error) {
	va, err := pc.volumeAttachmentLister.Get(volumeAttachmentName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("can't get VolumeAttachment %q: %w", volumeAttachmentName, err)
	}

	if va.DeletionTimestamp != nil {
		return false, fmt.Errorf("VolumeAttachment %q is being deleted", volumeAttachmentName)
	}

	if va.Status.AttachError != nil {
		return false, fmt.Errorf("volume attachment failed: %s", va.Status.AttachError.Message)
	}

	if va.Status.Attached {
		return true, nil
	}

	return false, nil
}

func (pc *Controller) mountDevice(ctx context.Context, spec *volume.Spec, deviceMounterArgs volume.DeviceMounterArgs) error {
	// Check if node stage/unstage is supported.

	csiAddress, err := pc.driverRegistry.GetAddress(spec.BackendPersistentVolume.Spec.CSI.Driver)
	if err != nil {
		return fmt.Errorf("driver %q is not supported: %w", spec.BackendPersistentVolume.Spec.CSI.Driver, err)
	}

	// TODO: create csi operation context

	driverClient, err := csi.NewDriverClient(csiAddress)
	if err != nil {
		return fmt.Errorf("can't create driver client: %w", err)
	}

	nodeSupportsStageUnstage, err := driverClient.NodeSupportsStageUnstage(ctx)
	if err != nil {
		return fmt.Errorf("can't verify whether driver supports node stage/unstage: %w", err)
	}

	publishContext, err := pc.getPublishContext(spec.BackendPersistentVolume)
	if err != nil {
		return fmt.Errorf("can't get publish context: %w", err)
	}

	nodeStageSecrets := map[string]string{}
	if spec.BackendPersistentVolume.Spec.CSI.NodeStageSecretRef != nil && nodeSupportsStageUnstage {
		nodeStageSecrets, err = pc.getCredentialsFromSecret(spec.BackendPersistentVolume.Spec.CSI.NodeStageSecretRef)
		if err != nil {
			return fmt.Errorf("can't get credentials from secret: %w", err)
		}
	}

	mountOptions := spec.BackendPersistentVolume.Spec.MountOptions

	// TODO: support selinux options

	// TODO: move this to state manager
	// Store volume metadata for UnmountDevice. Keep it around even if the
	// driver does not support NodeStage, UnmountDevice still needs it.
	//if err = mkdirAllWithPathCheck(deviceMountPath, 0750); err != nil {
	//	return fmt.Errorf("can't create dir %q: %w", deviceMountPath, err)
	//}

	//dataDir := filepath.Dir(deviceMountPath)
	//data := map[string]string{
	//	volDataKey.volHandle:  pv.Spec.CSI.VolumeHandle,
	//	volDataKey.driverName: pv.Spec.CSI.Driver,
	//}

	// TODO: selinux label

	//err = saveVolumeData(dataDir, volDataFileName, data)
	//if err != nil {
	//	return fmt.Errorf("can't save volume data: %w", err)
	//}
	//defer func() {
	//	// TODO
	//	//// Only if there was an error and volume operation was considered
	//	//// finished, we should remove the directory.
	//	//if err != nil && volumetypes.IsOperationFinishedError(err) {
	//	//	// clean up metadata
	//	//	if err := removeMountDir(c.plugin, deviceMountPath); err != nil {
	//	//		klog.Error(log("attacher.MountDevice failed to remove mount dir after error [%s]: %v", deviceMountPath, err))
	//	//	}
	//	//}
	//}()

	if !nodeSupportsStageUnstage {
		return nil
	}

	deviceMountPath, err := pc.getDeviceMountPath(spec)
	if err != nil {
		return fmt.Errorf("can't get device mount path: %w", err)
	}

	accessMode := corev1.ReadWriteOnce
	if spec.BackendPersistentVolume.Spec.AccessModes != nil {
		accessMode = spec.BackendPersistentVolume.Spec.AccessModes[0]
	}

	var nodeStageFSGroupArg *int64
	driverSupportsCSIVolumeMountGroup, err := driverClient.NodeSupportsVolumeMountGroup(ctx)
	if err != nil {
		// TODO: transient operation failure
		return fmt.Errorf("can't verify if the driver's node service has VOLUME_MOUNT_GROUP capability: %w", err)
	}

	if driverSupportsCSIVolumeMountGroup {
		// TODO
		//klog.V(3).Infof("Driver %s supports applying FSGroup (has VOLUME_MOUNT_GROUP node capability). Delegating FSGroup application to the driver through NodeStageVolume.", csiSource.Driver)
		nodeStageFSGroupArg = deviceMounterArgs.FsGroup
	}

	err = driverClient.NodeStageVolume(ctx,
		spec.BackendPersistentVolume.Spec.CSI.VolumeHandle,
		publishContext,
		deviceMountPath,
		spec.BackendPersistentVolume.Spec.CSI.FSType,
		accessMode,
		nodeStageSecrets,
		spec.BackendPersistentVolume.Spec.CSI.VolumeAttributes,
		mountOptions,
		nodeStageFSGroupArg,
	)
	if err != nil {
		return fmt.Errorf("can't node stage volume: %w", err)
	}

	return nil
}

func (pc *Controller) getDeviceMountPath(spec *volume.Spec) (string, error) {
	deviceMountPath, err := pc.stagingTargetPathBindStateManager.GetBoundStagingTargetPath(spec.ProxyPersistentVolume.Spec.CSI.VolumeHandle)
	if err != nil {
		return "", fmt.Errorf("can't get bound staging target path: %w", err)
	}

	return deviceMountPath, nil
}

// TODO: part of attacher? or move to state manager
// saveVolumeData persists parameter data as json file at the provided location
func saveVolumeData(dir string, fileName string, data map[string]string) error {
	dataFilePath := filepath.Join(dir, fileName)
	klog.V(4).InfoS("Saving volume data", "path", dataFilePath, "data", data)

	file, err := os.Create(dataFilePath)
	if err != nil {
		return fmt.Errorf("can't create volume data file at %q: %w", dataFilePath, err)
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		return fmt.Errorf("can't json encode data at %q: %w", dataFilePath, err)
	}

	klog.V(4).InfoS("Volume data saved", "path", dataFilePath)

	return nil
}

func (pc *Controller) getCredentialsFromSecret(secretRef *corev1.SecretReference) (map[string]string, error) {
	credentials := map[string]string{}
	secret, err := pc.kubeClient.CoreV1().Secrets(secretRef.Namespace).Get(context.TODO(), secretRef.Name, metav1.GetOptions{})
	if err != nil {
		return credentials, fmt.Errorf("can't get secret %q: %w", sonaming.ManualRef(secretRef.Namespace, secretRef.Name))
	}
	for key, value := range secret.Data {
		credentials[key] = string(value)
	}

	return credentials, nil
}

// TODO: part of attacher
func (pc *Controller) getPublishContext(pv *corev1.PersistentVolume) (map[string]string, error) {
	// TODO: move to a shared method of csi handler
	csiDriver, err := pc.csiDriverLister.Get(pv.Spec.CSI.Driver)
	if err != nil {
		return nil, fmt.Errorf("can't get CSIDriver %q: %w", sonaming.ManualRef("", pv.Spec.CSI.Driver), err)
	}

	if csiDriver.Spec.AttachRequired == nil || !*csiDriver.Spec.AttachRequired {
		return map[string]string{}, nil
	}

	volumeAttachmentName := getVolumeAttachmentName(pv.Spec.CSI.VolumeHandle, csiDriver.Name, pc.nodeName)

	va, err := pc.volumeAttachmentLister.Get(volumeAttachmentName)
	if err != nil {
		return nil, fmt.Errorf("can't get VolumeAttachment %q: %w", volumeAttachmentName, err)
	}

	return va.Status.AttachmentMetadata, nil
}

// TODO: make this a part of mounter struct?
func (pc *Controller) setUpAt(ctx context.Context, spec *volume.Spec, targetPath string, mounterArgs volume.MounterArgs) error {
	csiAddress, err := pc.driverRegistry.GetAddress(spec.BackendPersistentVolume.Spec.CSI.Driver)
	if err != nil {
		return fmt.Errorf("driver %q is not supported: %w", spec.BackendPersistentVolume.Spec.CSI.Driver, err)
	}

	// TODO: create csi operation context

	driverClient, err := csi.NewDriverClient(csiAddress)
	if err != nil {
		return fmt.Errorf("can't create driver client: %w", err)
	}

	//volumeLifecycleMode := storagev1.VolumeLifecyclePersistent
	driverSupportsVolumeLifecyclePersistent, err := pc.driverSupportsVolumeLifecyclePersistent(spec.BackendPersistentVolume)
	if err != nil {
		return fmt.Errorf("can't verify that driver supports volume lifecycle persistent")
	}
	if !driverSupportsVolumeLifecyclePersistent {
		return fmt.Errorf("driver does not support volume lifecycle persistent")
	}

	fsGroupPolicy, err := pc.getFSGroupPolicy(spec.BackendPersistentVolume)
	if err != nil {
		return fmt.Errorf("can't get FSGroup policy: %w", err)
	}

	//volumeHandle := pv.Spec.CSI.VolumeHandle
	// TODO: where should this come from?
	readOnly := false
	accessMode := corev1.ReadWriteOnce

	var (
		fsType             string
		volAttribs         map[string]string
		nodePublishSecrets map[string]string
		publishContext     map[string]string
		mountOptions       []string
		deviceMountPath    string
		secretRef          *corev1.SecretReference
	)

	fsType = spec.BackendPersistentVolume.Spec.CSI.FSType
	volAttribs = spec.BackendPersistentVolume.Spec.CSI.VolumeAttributes
	secretRef = spec.BackendPersistentVolume.Spec.CSI.NodePublishSecretRef

	// TODO: wtf? Make this better. This is copied from kube.
	if len(spec.BackendPersistentVolume.Spec.AccessModes) > 0 {
		accessMode = spec.BackendPersistentVolume.Spec.AccessModes[0]
	}

	mountOptions = spec.BackendPersistentVolume.Spec.MountOptions

	nodeSupportsStageUnstage, err := driverClient.NodeSupportsStageUnstage(ctx)
	if err != nil {
		return fmt.Errorf("can't verify whether driver supports node stage/unstage: %w", err)
	}

	if nodeSupportsStageUnstage {
		// TODO: this should have a common place, it doesn't make sense
		deviceMountPath, err = pc.getDeviceMountPath(spec)
		if err != nil {
			return fmt.Errorf("can't get device mount path: %w", err)
		}
	}

	publishContext, err = pc.getPublishContext(spec.BackendPersistentVolume)
	if err != nil {
		return fmt.Errorf("can't get publish context: %w", err)
	}

	// TODO: We don't have to create parent dir since target path should already exist. Maybe we should verify it?

	nodePublishSecrets = map[string]string{}
	if secretRef != nil {
		nodePublishSecrets, err = pc.getCredentialsFromSecret(secretRef)
		if err != nil {
			// TODO: transient operation failure
			return fmt.Errorf("can't get credentials from secret: %w", err)
		}
	}

	podInfoEnabled, err := pc.isPodInfoOnMountEnabled(spec.BackendPersistentVolume)
	if err != nil {
		// TODO: transient operation failure
		return fmt.Errorf("can't verify whether pod info on mount is enabled: %w", err)
	}
	if podInfoEnabled {
		volAttribs = maps.Merge(volAttribs, getPodInfoOnMountAttributes(spec.Pod, storagev1.VolumeLifecyclePersistent))
	}

	// TODO: service token attributes (important)

	var nodePublishFSGroupArg *int64
	driverSupportsCSIVolumeMountGroup, err := driverClient.NodeSupportsVolumeMountGroup(ctx)
	if err != nil {
		// TODO: transient operation failure
		return fmt.Errorf("can't verify if the driver's node service has VOLUME_MOUNT_GROUP capability: %w", err)
	}

	if driverSupportsCSIVolumeMountGroup {
		// TODO
		nodePublishFSGroupArg = mounterArgs.FsGroup
	}

	// TODO: selinux label mount

	//volData := map[string]string{
	//	volDataKey.specVolID:           pv.GetName(),
	//	volDataKey.volHandle:           volumeHandle,
	//	volDataKey.driverName:          pv.Spec.CSI.Driver,
	//	volDataKey.nodeName:            pc.nodeName,
	//	volDataKey.volumeLifecycleMode: string(volumeLifecycleMode),
	//	volDataKey.attachmentID:        getVolumeAttachmentName(pv.Spec.CSI.VolumeHandle, pv.Spec.CSI.Driver, pc.nodeName),
	//}
	//
	//parentDir := filepath.Dir(targetPath)
	// TODO: have our own way of doing this
	//err = saveVolumeData(parentDir, volDataFileName, volData)
	//defer func() {
	//	// TODO: remove saved data on fail
	//}()
	//if err != nil {
	//	// TODO: transient error
	//	return fmt.Errorf("can't save volume data: %w", err)
	//}

	// TODO: add selinux labels to voldata

	err = driverClient.NodePublishVolume(
		ctx,
		spec.BackendPersistentVolume.Spec.CSI.VolumeHandle,
		readOnly,
		deviceMountPath,
		targetPath,
		accessMode,
		publishContext,
		volAttribs,
		nodePublishSecrets,
		fsType,
		mountOptions,
		nodePublishFSGroupArg,
	)
	if err != nil {
		// TODO: check for operation finished error
		return fmt.Errorf("can't mount volume: %w", err)
	}

	// TODO: check selinux label mount

	// TODO: check if driver supports csi volume mount group, otherwise do it from the controller
	if !driverSupportsCSIVolumeMountGroup && supportsFSGroup(spec.BackendPersistentVolume, fsType, mounterArgs.FsGroup, fsGroupPolicy, readOnly) {
		// Driver doesn't support applying FSGroup. In this case kubelet would do it instead, so our controller has to as well.

		// TODO: either make it nicer or write our own replacement of this func
		err = k8svolume.SetVolumeOwnership(targetPath, mounterArgs.FsGroup, mounterArgs.FSGroupChangePolicy, nil, readOnly)
		if err != nil {
			// TODO: uncertain progress error
			return fmt.Errorf("can't set volume ownership: %w", err)
		}
	}

	return nil
}

func supportsFSGroup(pv *corev1.PersistentVolume, fsType string, fsGroup *int64, driverPolicy storagev1.FSGroupPolicy, readOnly bool) bool {
	if fsGroup == nil || driverPolicy == storagev1.NoneFSGroupPolicy || readOnly {
		return false
	}

	if driverPolicy == storagev1.FileFSGroupPolicy {
		return true
	}

	if len(fsType) == 0 {
		return false
	}

	if pv.Spec.AccessModes == nil {
		return false
	}

	hasReadWriteOnce := slices.Contains(pv.Spec.AccessModes, corev1.ReadWriteOnce)
	if !hasReadWriteOnce {
		return false
	}

	return true
}

// TODO: move to csi util
func getPodInfoOnMountAttributes(pod *corev1.Pod, volumeLifecycleMode storagev1.VolumeLifecycleMode) map[string]string {
	res := map[string]string{
		"csi.storage.k8s.io/pod.name":            pod.Name,
		"csi.storage.k8s.io/pod.namespace":       pod.Namespace,
		"csi.storage.k8s.io/pod.uid":             string(pod.UID),
		"csi.storage.k8s.io/serviceAccount.name": pod.Spec.ServiceAccountName,
		"csi.storage.k8s.io/ephemeral":           strconv.FormatBool(volumeLifecycleMode == storagev1.VolumeLifecycleEphemeral),
	}
	return res
}

// TODO: part of csi driver plugin abstraction? something like that
func (pc *Controller) isPodInfoOnMountEnabled(pv *corev1.PersistentVolume) (bool, error) {
	// TODO: check if is csi
	// TODO: move to a shared method of csi handler, this should be coming from the csi driver registry
	csiDriver, err := pc.csiDriverLister.Get(pv.Spec.CSI.Driver)
	if err != nil {
		return false, fmt.Errorf("can't get CSIDriver %q: %w", sonaming.ManualRef("", pv.Spec.CSI.Driver), err)
	}

	if csiDriver.Spec.PodInfoOnMount != nil && *csiDriver.Spec.PodInfoOnMount {
		return true, nil
	}

	return false, nil
}

func (pc *Controller) getFSGroupPolicy(pv *corev1.PersistentVolume) (storagev1.FSGroupPolicy, error) {
	// TODO: check if is csi
	// TODO: move to a shared method of csi handler, this should be coming from the csi driver registry
	csiDriver, err := pc.csiDriverLister.Get(pv.Spec.CSI.Driver)
	if err != nil {
		return "", fmt.Errorf("can't get CSIDriver %q: %w", sonaming.ManualRef("", pv.Spec.CSI.Driver), err)
	}

	if csiDriver.Spec.FSGroupPolicy == nil || *csiDriver.Spec.FSGroupPolicy == "" {
		return "", fmt.Errorf("csi driver %q has no FSGroupPolicy", pv.Spec.CSI.Driver)
	}

	return *csiDriver.Spec.FSGroupPolicy, nil
}

// TODO: part of mounter
func (pc *Controller) driverSupportsVolumeLifecyclePersistent(pv *corev1.PersistentVolume) (bool, error) {
	// TODO: check if is csi
	// TODO: move to a shared method of csi handler, this should be coming from the csi driver registry

	csiDriver, err := pc.csiDriverLister.Get(pv.Spec.CSI.Driver)
	if err != nil {
		return false, fmt.Errorf("can't get CSIDriver %q: %w", sonaming.ManualRef("", pv.Spec.CSI.Driver), err)
	}

	// TODO: volume lifecycle mode is persistent in our case, if we wanted to support ephemeral we'd have to change the volume source

	return slices.Contains(csiDriver.Spec.VolumeLifecycleModes, storagev1.VolumeLifecyclePersistent), nil
}
