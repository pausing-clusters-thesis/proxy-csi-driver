package secuirtycontext

import (
	corev1 "k8s.io/api/core/v1"
)

// DetermineEffectiveRunAsUser returns a pointer of UID from the provided pod's
// and container's security context and a bool value to indicate if it is absent.
// Container's runAsUser take precedence in cases where both are set.
func DetermineEffectiveRunAsUser(pod *corev1.Pod, container *corev1.Container) (*int64, bool) {
	var runAsUser *int64
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsUser != nil {
		runAsUser = new(int64)
		*runAsUser = *pod.Spec.SecurityContext.RunAsUser
	}
	if container.SecurityContext != nil && container.SecurityContext.RunAsUser != nil {
		runAsUser = new(int64)
		*runAsUser = *container.SecurityContext.RunAsUser
	}
	if runAsUser == nil {
		return nil, false
	}
	return runAsUser, true
}
