/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"github.com/go-logr/logr"
	gink "github.com/onsi/ginkgo/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// All of reconciliers of this operator must implement those
type CloudflareAccessReconciler interface {
	reconcile.Reconciler

	// if defined, [override] will be registered instead of natural reconcilier
	SetupWithManager(mgr ctrl.Manager, override reconcile.Reconciler) error

	//
	GetReconcilierLogger(ctx context.Context) logr.Logger
}

//
//
//

// Its purpose is to store all errors that happened during controllers reconciliations.
type ReconcilierErrorTracker []error

// Clears all errors tracked
func (rt *ReconcilierErrorTracker) Clear() {
	if rt != nil {
		*rt = (*rt)[:0]
	}
}

// Ginkgo Only: tests if [ReconcilierErrorTracker] contains tracked errors from reconciliers. If so, Fails().
func (rt *ReconcilierErrorTracker) TestEmpty() {
	if len(*rt) > 0 {
		gink.Fail("Reconciliation failed")
	}
}

//
//
//

// Allows to intercept all inner errors and push them into logger
type ReconcilerWithLoggedErrors struct {
	// [Inner] is the inner reconcilier that will have its errors logged
	Inner CloudflareAccessReconciler

	// Setting this would allow storage of intercepted errors by reconciliation logic. Used for testing purposes.
	ErrTracker *ReconcilierErrorTracker
}

func (rw *ReconcilerWithLoggedErrors) SetupWithManager(mgr ctrl.Manager) error {
	return rw.Inner.SetupWithManager(mgr, rw) //nolint:wrapcheck
}

//
//
//

func (rw *ReconcilerWithLoggedErrors) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	result, err := rw.Inner.Reconcile(ctx, req)

	if err != nil {
		// HERE, May add more logs about error
		// log := rw.Inner.GetReconcilierLogger(ctx)
		// log.Error(err, "reconciliation failed")

		//
		rw.maybeTrackErrors(err)
	}

	return result, err //nolint:wrapcheck
}

func (rw *ReconcilerWithLoggedErrors) maybeTrackErrors(err error) {
	if rw.ErrTracker != nil {
		*rw.ErrTracker = append(*rw.ErrTracker, err)
	}
}
