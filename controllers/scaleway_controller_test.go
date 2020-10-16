package controllers

import (
	"testing"
	"time"

	scalewaymetav1alpha1 "github.com/scaleway/scaleway-operator/apis/meta/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func compareStatus(s1, s2 *scalewaymetav1alpha1.Status) bool {
	if len(s1.Conditions) != len(s2.Conditions) {
		return false
	}
	for i := range s1.Conditions {
		c1 := s1.Conditions[i]
		c2 := s2.Conditions[i]
		if !c1.LastProbeTime.Equal(&c2.LastProbeTime) ||
			!c1.LastTransitionTime.Equal(&c2.LastTransitionTime) ||
			c1.Message != c2.Message ||
			c1.Reason != c2.Reason ||
			c1.Status != c2.Status ||
			c1.Type != c2.Type {
			return false
		}
	}
	return true
}

func Test_updateStatus(t *testing.T) {
	now := metav1.NewTime(time.Now())
	before := metav1.NewTime(time.Now().Add(-5 * time.Second))

	notFoundErr := &scw.ResourceNotFoundError{
		Resource:   "dummy",
		ResourceID: "dummyID",
	}

	tStateErr := &scw.TransientStateError{
		CurrentState: "dummy-state",
		Resource:     "dummy",
		ResourceID:   "dummyID",
	}

	permDeniedErr := &scw.PermissionsDeniedError{}

	oosErr := &scw.OutOfStockError{
		Resource: "dummy",
	}

	quotasExceededError := &scw.QuotasExceededError{
		Details: []scw.QuotasExceededErrorDetail{
			{
				Resource: "dummy",
				Current:  2,
				Quota:    3,
			},
		},
	}

	resLockedErr := &scw.ResourceLockedError{
		Resource:   "dummy",
		ResourceID: "dummyID",
	}

	invalidArgErr := &scw.InvalidArgumentsError{
		Details: []scw.InvalidArgumentsErrorDetail{
			{
				ArgumentName: "dummy-arg",
				HelpMessage:  "I NEED SOMEBODY! HELP! NOT JUST ANYBODY",
				Reason:       "I'm feeling down",
			},
		},
	}

	cases := []struct {
		status     *scalewaymetav1alpha1.Status
		ensureErr  error
		reconciled bool
		err        error
		newStatus  *scalewaymetav1alpha1.Status
	}{
		{
			&scalewaymetav1alpha1.Status{},
			nil,
			true,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: now,
						Message:            "",
						Reason:             "",
						Status:             corev1.ConditionTrue,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{},
			nil,
			false,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: now,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
			nil,
			true,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: now,
						Message:            "",
						Reason:             "",
						Status:             corev1.ConditionTrue,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
			nil,
			false,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: before,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
			notFoundErr,
			false,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: before,
						Message:            notFoundErr.Error(),
						Reason:             reasonResourceNotFound,
						Status:             corev1.ConditionFalse,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
			tStateErr,
			false,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: before,
						Message:            tStateErr.Error(),
						Reason:             reasonTransientState,
						Status:             corev1.ConditionFalse,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
			permDeniedErr,
			false,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: before,
						Message:            permDeniedErr.Error(),
						Reason:             reasonPermissionsDenied,
						Status:             corev1.ConditionFalse,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
			oosErr,
			false,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: before,
						Message:            oosErr.Error(),
						Reason:             reasonOutOfStock,
						Status:             corev1.ConditionFalse,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
			quotasExceededError,
			false,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: before,
						Message:            quotasExceededError.Error(),
						Reason:             reasonQuotasExceeded,
						Status:             corev1.ConditionFalse,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
			resLockedErr,
			false,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: before,
						Message:            resLockedErr.Error(),
						Reason:             reasonResourceLocked,
						Status:             corev1.ConditionFalse,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            messageStillReconciling,
						Reason:             reasonReconciling,
						Status:             corev1.ConditionFalse,
					},
				},
			},
			invalidArgErr,
			false,
			nil,
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Reconciled,
						LastProbeTime:      now,
						LastTransitionTime: before,
						Message:            invalidArgErr.Error(),
						Reason:             reasonInvalidArguments,
						Status:             corev1.ConditionFalse,
					},
				},
			},
		},
	}
	for _, c := range cases {
		_, err := updateStatus(c.status, now, c.ensureErr, c.reconciled)
		if err != c.err {
			t.Errorf("Got error %+v instead of %+v", err, c.err)
		}
		if !compareStatus(c.newStatus, c.status) {
			t.Errorf("Got status %+v instead of %+v", *c.status, *c.newStatus)
		}
	}
}

func Test_updateCondition(t *testing.T) {
	now := metav1.NewTime(time.Now())
	before := metav1.NewTime(time.Now().Add(-5 * time.Second))

	cases := []struct {
		status    *scalewaymetav1alpha1.Status
		condition scalewaymetav1alpha1.Condition
		newStatus *scalewaymetav1alpha1.Status
	}{
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{},
			},
			scalewaymetav1alpha1.Condition{
				Type:    scalewaymetav1alpha1.Ready,
				Message: "message",
				Reason:  "reason",
				Status:  corev1.ConditionTrue,
			},
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Ready,
						LastProbeTime:      now,
						LastTransitionTime: now,
						Message:            "message",
						Reason:             "reason",
						Status:             corev1.ConditionTrue,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Ready,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            "message",
						Reason:             "reason",
						Status:             corev1.ConditionTrue,
					},
				},
			},
			scalewaymetav1alpha1.Condition{
				Type:    scalewaymetav1alpha1.Ready,
				Message: "message",
				Reason:  "reason",
				Status:  corev1.ConditionTrue,
			},
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Ready,
						LastProbeTime:      now,
						LastTransitionTime: before,
						Message:            "message",
						Reason:             "reason",
						Status:             corev1.ConditionTrue,
					},
				},
			},
		},
		{
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Ready,
						LastProbeTime:      before,
						LastTransitionTime: before,
						Message:            "message",
						Reason:             "reason",
						Status:             corev1.ConditionTrue,
					},
				},
			},
			scalewaymetav1alpha1.Condition{
				Type:    scalewaymetav1alpha1.Ready,
				Message: "message",
				Reason:  "reason",
				Status:  corev1.ConditionFalse,
			},
			&scalewaymetav1alpha1.Status{
				Conditions: []scalewaymetav1alpha1.Condition{
					{
						Type:               scalewaymetav1alpha1.Ready,
						LastProbeTime:      now,
						LastTransitionTime: now,
						Message:            "message",
						Reason:             "reason",
						Status:             corev1.ConditionFalse,
					},
				},
			},
		},
	}

	for _, c := range cases {
		updateCondition(c.status, c.condition, now)
		if !compareStatus(c.newStatus, c.status) {
			t.Errorf("Got status %+v instead of %+v", *c.status, *c.newStatus)
		}
	}
}
