/*
Copyright 2022.

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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	configv1alpha1 "github.com/deislabs/ratify/api/v1alpha1"
)

// VerifierReconciler reconciles a Verifier object
type VerifierReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=verifiers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=verifiers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=verifiers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Verifier object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *VerifierReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	_ = log.FromContext(ctx)

	var verifier configv1alpha1.Verifier

	if err := r.Get(ctx, req.NamespacedName, &verifier); err != nil {
		log.Log.Error(err, "unable to fetch verifier")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Log.Info("verifier " + verifier.Spec.Name)
	log.Log.Info("ArtifactTypes " + verifier.Spec.ArtifactTypes)

	myString := string(verifier.Spec.Parameters.Raw)
	log.Log.Info("Raw string " + myString)

	var p map[string]interface{}
	err := json.Unmarshal(verifier.Spec.Parameters.Raw, &p)

	if err != nil {
		// TODO: are there any verifier with no parameters
		log.Log.Error(err, "unable to decode verifier parameters")
	}

	for key, value := range p {
		switch c := value.(type) {
		case string:
			fmt.Printf("Item %q is a string, containing %q\n", key, c)
			log.Log.Info("key  " + key + ", value" + value.(string))
		case float64:
			fmt.Printf("Looks like item %q is a number, specifically %f\n", key, value)
		default:
			fmt.Printf("Not sure what type item %q is, but I think it might be %T\n", key, value)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VerifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.Verifier{}).
		Complete(r)
}
