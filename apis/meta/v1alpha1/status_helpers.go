package v1alpha1

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// IsReady returns true is the Ready condition is True
func (s Status) IsReady() bool {
	for _, c := range s.Conditions {
		if c.Type == Ready && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// IsReconciled returns true is the Ready condition is True
func (s Status) IsReconciled() bool {
	for _, c := range s.Conditions {
		if c.Type == Reconciled && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// IsReconciledAfter returns true is the Ready condition is True
func (s Status) IsReconciledAfter(now time.Time) bool {
	for _, c := range s.Conditions {
		if c.Type == Reconciled && c.Status == corev1.ConditionTrue && c.LastTransitionTime.After(now) {
			return true
		}
	}
	return false
}
