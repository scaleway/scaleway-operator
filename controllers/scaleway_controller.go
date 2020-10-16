package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"
	scalewaymetav1alpha1 "github.com/scaleway/scaleway-operator/apis/meta/v1alpha1"
	"github.com/scaleway/scaleway-operator/pkg/manager/scaleway"
	"github.com/scaleway/scaleway-sdk-go/scw"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	finalizerName    = "scaleway.com/finalizer"
	ignoreAnnotation = "scaleway.com/ignore"

	messageStillReconciling = "Still reconciling"

	reasonReconciling       = "Reconciling"
	reasonTransientState    = "TransientState"
	reasonResourceNotFound  = "ResourceNotFound"
	reasonPermissionsDenied = "PermissionsDenied"
	reasonOutOfStock        = "OutOfStock"
	reasonQuotasExceeded    = "QuotasExceeded"
	reasonResourceLocked    = "ResourceLocked"
	reasonInvalidArguments  = "InvalidArguments"
)

var (
	// RequeueDuration is the default requeue duration
	// It is exported to reduce it when running tests with existing data
	RequeueDuration time.Duration = time.Second * 30
)

// ScalewayReconciler is the base reconciler for Scaleway products
type ScalewayReconciler struct {
	client.Client
	ScalewayManager scaleway.Manager
	Recorder        record.EventRecorder
	Log             logr.Logger
	Scheme          *runtime.Scheme
}

// Reconcile is the global reconcile loop
func (r *ScalewayReconciler) Reconcile(req ctrl.Request, obj runtime.Object) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("object", req.String())

	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		log.Error(err, "could not find object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	objMeta, err := meta.Accessor(obj)
	if err != nil {
		log.Error(err, "failed to get object meta")
		return ctrl.Result{}, err
	}

	if ignore, ok := objMeta.GetAnnotations()[ignoreAnnotation]; ok {
		if strings.ToLower(ignore) == "true" {
			if !objMeta.GetDeletionTimestamp().IsZero() {
				controllerutil.RemoveFinalizer(obj.(controllerutil.Object), finalizerName)
			}
			r.Recorder.Event(obj, corev1.EventTypeNormal, "Ignoring", "Ignoring object based on annotation")
			return ctrl.Result{}, r.Update(ctx, obj)
		}
	}

	if objMeta.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(obj.(controllerutil.Object), finalizerName) {
			controllerutil.AddFinalizer(obj.(controllerutil.Object), finalizerName)
			if err := r.Update(ctx, obj); err != nil {
				log.Error(err, "failed to add finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		// deletion
		if controllerutil.ContainsFinalizer(obj.(controllerutil.Object), finalizerName) {
			deleted, err := r.ScalewayManager.Delete(ctx, obj)
			if err != nil {
				log.Error(err, "failed to delete")
				return ctrl.Result{}, err
			}
			if deleted {
				r.Recorder.Event(obj, corev1.EventTypeNormal, "Deleted", "Successfully deleted")
				controllerutil.RemoveFinalizer(obj.(controllerutil.Object), finalizerName)
				return ctrl.Result{}, r.Update(ctx, obj)
			}
			log.Info("still deleting")
			return ctrl.Result{RequeueAfter: RequeueDuration}, r.Status().Update(ctx, obj)
		}
		return ctrl.Result{}, nil
	}

	err = r.setOwners(ctx, log, obj, objMeta)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("reconciling object")

	ensured, err := r.ScalewayManager.Ensure(ctx, obj)
	if err != nil {
		log.Error(err, "error ensuring object")
	}

	scalewayStatus := obj.(scalewaymetav1alpha1.TypeMeta).GetStatus()
	requeueAfter, updateErr := updateStatus(&scalewayStatus, metav1.NewTime(time.Now()), err, ensured)
	obj.(scalewaymetav1alpha1.TypeMeta).SetStatus(scalewayStatus)
	err = r.Status().Update(ctx, obj)
	if err != nil {
		log.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, updateErr
}

func (r *ScalewayReconciler) setOwners(ctx context.Context, log logr.Logger, obj runtime.Object, objMeta metav1.Object) error {
	owners, err := r.ScalewayManager.GetOwners(ctx, obj)
	if err != nil {
		log.Error(err, "failed to get owners")
		return err
	}

	for _, owner := range owners {
		if err := r.Get(ctx, owner.Key, owner.Object); err == nil {
			if ownerMeta, err := meta.Accessor(owner.Object); err == nil {
				err = controllerutil.SetControllerReference(ownerMeta, objMeta, r.Scheme)
				if err != nil {
					log.Error(err, "failed to set controller reference")
					continue
				}
				err := r.Update(ctx, obj)
				if err != nil {
					log.Error(err, "failed to update controller reference")
					continue
				}
				log.Info("controller reference set")
				break
			}
		}
	}

	return nil
}

func updateStatus(status *scalewaymetav1alpha1.Status, now metav1.Time, ensureErr error, reconciled bool) (time.Duration, error) {
	reconcileStatus := corev1.ConditionTrue
	var message string
	var reason string

	var requeueAfter time.Duration

	if ensureErr != nil {
		message = ensureErr.Error()
		reconcileStatus = corev1.ConditionFalse

		switch ensureErr.(type) {
		case *scw.ResourceNotFoundError:
			reason = reasonResourceNotFound
			ensureErr = nil
		case *scw.InvalidArgumentsError:
			reason = reasonInvalidArguments
			ensureErr = nil
		case *scw.PermissionsDeniedError:
			reason = reasonPermissionsDenied
			requeueAfter = RequeueDuration * 10
			ensureErr = nil
		case *scw.OutOfStockError:
			reason = reasonOutOfStock
			requeueAfter = RequeueDuration * 4
			ensureErr = nil
		case *scw.QuotasExceededError:
			reason = reasonQuotasExceeded
			requeueAfter = RequeueDuration * 2
			ensureErr = nil
		case *scw.ResourceLockedError:
			reason = reasonResourceLocked
			requeueAfter = RequeueDuration * 10
			ensureErr = nil
		case *scw.TransientStateError:
			reason = reasonTransientState
			requeueAfter = RequeueDuration
			ensureErr = nil
		}

	} else {
		if !reconciled {
			requeueAfter = RequeueDuration
			reconcileStatus = corev1.ConditionFalse
			message = messageStillReconciling
			reason = reasonReconciling
		}
	}

	updateCondition(status, scalewaymetav1alpha1.Condition{
		Message: message,
		Reason:  reason,
		Status:  reconcileStatus,
		Type:    scalewaymetav1alpha1.Reconciled,
	}, now)

	return requeueAfter, ensureErr
}

func updateCondition(status *scalewaymetav1alpha1.Status, condition scalewaymetav1alpha1.Condition, now metav1.Time) {
	for i, c := range status.Conditions {
		if c.Type == condition.Type {
			cond := &status.Conditions[i]
			cond.LastProbeTime = now
			cond.Message = condition.Message
			cond.Reason = condition.Reason
			if cond.Status != condition.Status {
				cond.LastTransitionTime = now
			}
			cond.Status = condition.Status
			return
		}
	}
	status.Conditions = append(status.Conditions, scalewaymetav1alpha1.Condition{
		Type:               condition.Type,
		LastProbeTime:      now,
		LastTransitionTime: now,
		Message:            condition.Message,
		Reason:             condition.Reason,
		Status:             condition.Status,
	})
}
