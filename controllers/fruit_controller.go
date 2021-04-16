/*


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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	eatv1 "github.com/sm43/fruit-controller/api/v1"
)

// FruitReconciler reconciles a Fruit object
type FruitReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=eat.sm43.dev,resources=fruits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eat.sm43.dev,resources=fruits/status,verbs=get;update;patch

func (r *FruitReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("fruit", req.NamespacedName)

	var fruit eatv1.Fruit
	if err := r.Get(ctx, req.NamespacedName, &fruit); err != nil {
		log.Error(err, "failed to get fruit")
		return ctrl.Result{}, err
	}

	desiredFruit := fruitDeployment(fruit)
	if err := r.Create(ctx, desiredFruit); err != nil {
		log.Error(err, "failed to create fruit deployment")
		return ctrl.Result{}, err
	}

	fruit.Status.Created = true
	if err := r.Status().Update(ctx, &fruit); err != nil {
		log.Error(err, "failed to update fruit status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *FruitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eatv1.Fruit{}).
		Complete(r)
}

func fruitDeployment(fruit eatv1.Fruit) *appsv1.Deployment {
	replicas := int32(fruit.Spec.Number)
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fruit.ObjectMeta.Name,
			Namespace: fruit.ObjectMeta.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"eatv1/fruit": fruit.ObjectMeta.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fruit.ObjectMeta.Name,
					Namespace: fruit.ObjectMeta.Namespace,
					Labels: map[string]string{
						"eatv1/fruit": fruit.ObjectMeta.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "fruit-container",
							Image: "nginx",
						},
					},
				},
			},
		},
	}
	d.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         fruit.APIVersion,
			Kind:               fruit.Kind,
			Name:               fruit.Namespace,
			UID:                fruit.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return d
}
