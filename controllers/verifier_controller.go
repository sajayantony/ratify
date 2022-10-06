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
	//VerifiersMap map[string]vr.ReferenceVerifier
}

var (
	// a map to track of active verifiers
	VerifierMap     = map[string]vr.ReferenceVerifier{}
	verifierVersion = "1.0.0"
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

		// SusanTODO, log a message for other active verifier
		if apierrors.IsNotFound(err) {
			log.Log.Info(fmt.Sprintf("Delete event detected, removing verifier %v", req.Name))
		} else {
			log.Log.Error(err, "unable to fetch verifier")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := verifierAddOrReplace(verifier.Spec, req.Name)

	if err != nil {
		log.Log.Error(err, "unable to create verifier from verifier crd")
	}

	return ctrl.Result{}, nil
}

// TODO, do we care about the namespace, do we want to to handle objects in the ratify deployed namespace?
// creates a verifier reference from CRD spec and add store to map
func verifierAddOrReplace(spec configv1alpha1.VerifierSpec, objectName string) error {
	verifierConfig, err := specToVerifierConfig(spec)

	verifierReference, err := vf.CreateVerifierFromConfig(verifierConfig, verifierVersion, []string{spec.Address})

	if err != nil || verifierReference == nil {
		log.Log.Error(err, "unable to create verifier from verifier config")
	} else {
		VerifierMap[objectName] = verifierReference
		log.Log.Info(fmt.Sprintf("New verifier '%v' added to verifier map", verifierReference.Name()))
	}

	return err
}

// remove verifier from map
func verifierRemove(objectName string) {
	delete(VerifierMap, objectName)
}

// returns a verifier reference from spec
func specToVerifierConfig(verifierSpec configv1alpha1.VerifierSpec) (config.VerifierConfig, error) {
	log.Log.Info("verifier " + verifierSpec.Name)
	log.Log.Info("ArtifactTypes " + verifierSpec.ArtifactTypes)

	myString := string(verifierSpec.Parameters.Raw)
	log.Log.Info("Raw string " + myString)

	verifierConfig := config.VerifierConfig{}
	// SusanTODO: get json name of 'name'

	verifierConfig["name"] = verifierSpec.Name
	verifierConfig["artifactTypes"] = verifierSpec.ArtifactTypes

	if verifierSpec.Address == "" {
		//SusanTODO , handle address
		log.Log.Info("Verifier address is empty")
	}

	if string(verifierSpec.Parameters.Raw) != "" {
		var propertyMap map[string]interface{}
		err := json.Unmarshal(verifierSpec.Parameters.Raw, &propertyMap)
		if err != nil {
			log.Log.Error(err, "unable to decode verifier parameters", "Parameters.Raw", verifierSpec.Parameters.Raw)
			return config.VerifierConfig{}, err
		}
		for key, value := range propertyMap {
			verifierConfig[key] = value
		}
	}

	return verifierConfig, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VerifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.Verifier{}).
		Complete(r)
}
