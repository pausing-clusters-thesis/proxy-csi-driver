package util

import (
	podutil "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/thirdparty/k8s.io/kubernetes/pkg/api/v1/pod"
	securitycontext "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/thirdparty/k8s.io/kubernetes/pkg/secuirtycontext"
	corev1 "k8s.io/api/core/v1"
)

// FsUserFrom returns FsUser of pod, which is determined by the runAsUser
// attributes.
func FsUserFrom(pod *corev1.Pod) *int64 {
	var fsUser *int64
	// Exclude ephemeral containers because SecurityContext is not allowed.
	podutil.VisitContainers(&pod.Spec, podutil.InitContainers|podutil.Containers, func(container *corev1.Container, containerType podutil.ContainerType) bool {
		runAsUser, ok := securitycontext.DetermineEffectiveRunAsUser(pod, container)
		// One container doesn't specify user or there are more than one
		// non-root UIDs.
		if !ok || (fsUser != nil && *fsUser != *runAsUser) {
			fsUser = nil
			return false
		}
		if fsUser == nil {
			fsUser = runAsUser
		}
		return true
	})
	return fsUser
}
