/*
Copyright The Ratify Authors.
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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	configv1alpha1 "github.com/deislabs/ratify/api/v1alpha1"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	sf "github.com/deislabs/ratify/pkg/referrerstore/factory"
	"github.com/deislabs/ratify/pkg/referrerstore/types"
)

// StoreReconciler reconciles a Store object
type StoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	// a map to track active stores
	StoreMap = map[string]referrerstore.ReferrerStore{}
	// default version of the store
	storeVersion = "1.0.0"
)

//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=stores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=stores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=stores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *StoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var store configv1alpha1.Store

	if err := r.Get(ctx, req.NamespacedName, &store); err != nil {

		if apierrors.IsNotFound(err) {
			log.Log.Info(fmt.Sprintf("deletion detected, removing store %v", req.Name))
			storeRemove(req.Name)
		} else {
			log.Log.Error(err, "unable to fetch store")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := storeAddOrReplace(store.Spec, req.Name)

	if err != nil {
		log.Log.Error(err, "unable to create store from store crd")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.Store{}).
		Complete(r)
}

// Creates a store reference from CRD spec and add store to map
func storeAddOrReplace(spec configv1alpha1.StoreSpec, objectName string) error {
	storeConfig, err := specToStoreConfig(spec)

	storeReference, err := sf.CreateStoreFromConfig(storeConfig, storeVersion, []string{spec.Address})

	if err != nil || storeReference == nil {
		log.Log.Error(err, "unable to create store from store config")
	} else {
		StoreMap[objectName] = storeReference
		log.Log.Info(fmt.Sprintf("new store '%v' added to verifier map", storeReference.Name()))
	}

	return err
}

// Remove store from map
func storeRemove(objectName string) {
	delete(StoreMap, objectName)
}

// Returns a store reference from spec
func specToStoreConfig(storeSpec configv1alpha1.StoreSpec) (config.StorePluginConfig, error) {

	myString := string(storeSpec.Parameters.Raw)
	log.Log.Info("Raw string " + myString)

	storeConfig := config.StorePluginConfig{}

	storeConfig[types.Name] = storeSpec.Name

	if string(storeSpec.Parameters.Raw) != "" {
		var propertyMap map[string]interface{}
		err := json.Unmarshal(storeSpec.Parameters.Raw, &propertyMap)
		if err != nil {
			log.Log.Error(err, "unable to decode store parameters", "Parameters.Raw", storeSpec.Parameters.Raw)
			return config.StorePluginConfig{}, err
		}
		for key, value := range propertyMap {
			storeConfig[key] = value
		}
	}

	return storeConfig, nil
}
