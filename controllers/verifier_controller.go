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

	configv1alpha1 "github.com/deislabs/ratify/api/v1alpha1"
	vr "github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/config"
	vf "github.com/deislabs/ratify/pkg/verifier/factory"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// VerifierReconciler reconciles a Verifier object
type VerifierReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	verifiersMap = map[string]vr.ReferenceVerifier{}
)

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

		// SusanTODO, log a message if this is the last verifier
		if apierrors.IsNotFound(err) {
			log.Log.Info("Removing verifier " + req.Name)
			delete(verifiersMap, req.Name)
		} else {
			log.Log.Error(err, "unable to fetch verifier")

		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Log.Info("verifier " + verifier.Spec.Name)
	log.Log.Info("ArtifactTypes " + verifier.Spec.ArtifactTypes)

	myString := string(verifier.Spec.Parameters.Raw)
	log.Log.Info("Raw string " + myString)

	// setup verifier config map
	// SusanTODO: get json name of 'name'
	verifierConfig := config.VerifierConfig{}
	verifierConfig["name"] = verifier.Spec.Name
	verifierConfig["artifactTypes"] = verifier.Spec.ArtifactTypes

	if verifier.Spec.Address == "" {
		//SusanTODO , handle address
		log.Log.Info("Verifier addresss is empty")
	}

	var propertyMap map[string]interface{}
	err := json.Unmarshal(verifier.Spec.Parameters.Raw, &propertyMap)
	if err != nil {
		log.Log.Error(err, "unable to decode verifier parameters", "Parameters.Raw", verifier.Spec.Parameters.Raw)
	}

	if propertyMap == nil {
		log.Log.Info("verifier propertyMap is empty")
	} else {
		for key, value := range propertyMap {
			verifierConfig[key] = value
		}
	}

	// SusanTODO: how do we get version from the Crd
	verifierReference, err := vf.CreateVerifierFromConfig(verifierConfig, "1.0.0", []string{verifier.Spec.Address})

	if err != nil || verifierReference == nil {
		log.Log.Error(err, "unable to create verifier from verifier config")
	} else {
		verifiersMap[req.Name] = verifierReference
		log.Log.Info("New verifier created")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VerifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.Verifier{}).
		Complete(r)
}
