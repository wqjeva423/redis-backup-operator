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
	rediswqj "RedisBackupOperator/utils/redis"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "RedisBackupOperator/api/v1alpha1"
)

// RedisoperateReconciler reconciles a Redisoperate object
type RedisoperateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operator.bobfintech.com,resources=redisoperates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.bobfintech.com,resources=redisoperates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.bobfintech.com,resources=redisoperates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Redisoperate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *RedisoperateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	operateInstance := &operatorv1alpha1.Redisoperate{}
	err := r.Get(ctx, req.NamespacedName, operateInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
	}

	sentinellist := corev1.PodList{}
	error := r.Client.List(ctx, &sentinellist, client.MatchingLabels{
		"app.kubernetes.io/component":  "sentinel",
		"app.kubernetes.io/managed-by": "redis-operator",
		"app.kubernetes.io/name":       operateInstance.Spec.ClusterName,
		"app.kubernetes.io/part-of":    "redis-failover"})
	if error != nil {
		msg := fmt.Sprintf("RedisOperate Get Sentinel pod error!")
		log.Log.Error(error, msg)
	}

	if operateInstance.Spec.FailoverFlag == 1 {
		log.Log.Info("RedisOperate Cluster " + operateInstance.Spec.ClusterName + ": start to failover")
		sentinelIP := sentinellist.Items[0].Status.PodIP
		res, err := rediswqj.FailOverOperate(sentinelIP, "mymaster")
		if err != nil {
			msg := fmt.Sprintf("RedisOperate Cluster %s: %s", operateInstance.Spec.ClusterName, res)
			log.Log.Error(err, msg)
		} else {
			msg := fmt.Sprintf("RedisOperate Cluster %s: %s", operateInstance.Spec.ClusterName, res)
			log.Log.Info(msg)
		}

		original := operateInstance.DeepCopy()
		operateInstance.Spec.FailoverFlag = 0
		if err := r.Patch(ctx, operateInstance, client.MergeFrom(original)); err != nil {
			log.Log.Error(err, "RedisOperate Cluster "+operateInstance.Spec.ClusterName+": Failed to patch spec")
			return ctrl.Result{}, err
		}

		msg := fmt.Sprintf("RedisOperate Cluster %s: Successfully patched spec.FailoverFlag to 0", operateInstance.Spec.ClusterName)
		log.Log.Info(msg)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisoperateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Redisoperate{}).
		Complete(r)
}
