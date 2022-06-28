package controllers

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	FinalizerNode corev1.FinalizerName = "geth-operator-node"
)

// RequeueIfError requeues if an error is found.
func RequeueIfError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// RequeueImmediately requeues immediately when Requeue is True and no duration is specified.
func RequeueImmediately() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

// RequeueAfterInterval requeues after a duration when duration > 0 is specified.
func RequeueAfterInterval(interval time.Duration, err error) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: interval}, err
}

// NoRequeue does not requeue when Requeue is False.
func NoRequeue() (ctrl.Result, error) {
	return RequeueIfError(nil)
}

// RemoveString removes a specific string from a splice and returns the rest.
func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// ContainsString is a helper functions to check and remove string from a slice of strings.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func IgnoreAlreadyExists(err error) error {
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
