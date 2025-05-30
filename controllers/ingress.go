package controllers

/*
Copyright 2021-2025 The k8gb Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Generated by GoLic, for more details see: https://github.com/AbsaOSS/golic
*/

import (
	"context"
	"reflect"

	"github.com/k8gb-io/k8gb/controllers/utils"

	k8gbv1beta1 "github.com/k8gb-io/k8gb/api/v1beta1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *GslbReconciler) createIngressFromGslb(gslb *k8gbv1beta1.Gslb) (*netv1.Ingress, error) {
	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gslb.Name,
			Namespace: gslb.Namespace,
		},
		Spec: k8gbv1beta1.ToV1IngressSpec(gslb.Spec.Ingress),
	}
	utils.SetCommonGslbLabels(ingress)
	err := controllerutil.SetControllerReference(gslb, ingress, r.Scheme)
	if err != nil {
		return nil, err
	}
	return ingress, err
}

func (r *GslbReconciler) saveDependentIngress(instance *k8gbv1beta1.Gslb, i *netv1.Ingress) error {
	found := &netv1.Ingress{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		log.Info().
			Str("namespace", i.Namespace).
			Str("ingress", i.Name).
			Msg("Creating a new Ingress")
		err = r.Create(context.TODO(), i)

		if err != nil {
			// Creation failed
			log.Err(err).
				Str("namespace", i.Namespace).
				Str("name", i.Name).
				Msg("Failed to create new Ingress")
			return err
		}
		// Creation was successful
		return nil
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Err(err).Msg("Failed to get Ingress")
		return err
	}

	// Update existing object with new spec and annotations
	if !ingressEqual(found, i) {
		found.Spec = i.Spec
		utils.SetCommonGslbLabels(found)
		err = r.Update(context.TODO(), found)
		if errors.IsConflict(err) {
			log.Info().
				Str("namespace", found.Namespace).
				Str("name", found.Name).
				Msg("Ingress has been modified outside of controller, retrying reconciliation")
			return nil
		}
		if err != nil {
			// Update failed
			log.Err(err).
				Str("namespace", found.Namespace).
				Str("name", found.Name).
				Msg("Failed to update Ingress")
			return err
		}
	}

	return nil
}

func ingressEqual(ing1 *netv1.Ingress, ing2 *netv1.Ingress) bool {
	return reflect.DeepEqual(ing1.Spec, ing2.Spec)
}
