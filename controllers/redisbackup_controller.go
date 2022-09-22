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
	operatorv1alpha1 "RedisBackupOperator/api/v1alpha1"
	"RedisBackupOperator/utils"
	"context"
	"fmt"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RedisBackupReconciler reconciles a RedisBackup object
type RedisBackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operator.bobfintech.com,resources=redisbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.bobfintech.com,resources=redisbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.bobfintech.com,resources=redisbackups/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RedisBackup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *RedisBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	backupInstance := &operatorv1alpha1.RedisBackup{}
	err := r.Get(ctx, req.NamespacedName, backupInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
	}
	backupJob := &batch.Job{}
	backupJobGet := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      backupInstance.Name + "-job",
	}

	podlist := corev1.PodList{}
	masterPvc := ""
	backupPvc := ""
	error := r.Client.List(ctx, &podlist, client.MatchingLabels{"app.kubernetes.io/component": "redis", "app.kubernetes.io/managed-by": "redis-operator", "app.kubernetes.io/name": backupInstance.Spec.ClusterName, "app.kubernetes.io/part-of": "redis-failover"})
	if error != nil {
		fmt.Println(error)
	} else {
		for k, v := range podlist.Items {
			fmt.Println(podlist.Items[k].ObjectMeta.Labels["redisfailovers-role"])
			if podlist.Items[k].ObjectMeta.Labels["redisfailovers-role"] == "master" {
				masterPvc = v.Spec.Volumes[0].PersistentVolumeClaim.ClaimName
			}
			if podlist.Items[k].ObjectMeta.Labels["redisfailovers-role"] == "slave" {
				if v.Status.Phase == "Running" {
					backupPvc = v.Spec.Volumes[0].PersistentVolumeClaim.ClaimName
					fmt.Println(k, v.ObjectMeta.Name, v.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
					break
				} else {
					backupPvc = ""
					fmt.Println("Slave Pod: " + v.ObjectMeta.Name + " is not running")
				}
			}
			if backupPvc == "" {
				if v.Status.Phase == "Running" {
					backupPvc = masterPvc
				} else {
					fmt.Println("no Running pod to backup")
				}
			}
		}
	}
	fmt.Println("backup PVC is : " + backupPvc)

	err = r.Client.Get(ctx, backupJobGet, backupJob)
	if errors.IsNotFound(err) {
		job := utils.MakeJob(backupPvc, *backupInstance)
		if err := controllerutil.SetControllerReference(backupInstance, job, r.Scheme); err != nil {
			msg := fmt.Sprintf("set controllerReference for Job %s/%s failed", job.Namespace, job.Name)
			log.Log.Error(err, msg)
			return ctrl.Result{Requeue: true}, err
		}

		msg := fmt.Sprintf("create job %s/%s", job.Namespace, job.Name)
		log.Log.Info(msg)
		err = r.Client.Create(context.TODO(), job)
		if err != nil {
			msg := fmt.Sprintf("create job %s/%s error!", job.Namespace, job.Name)
			log.Log.Error(err, msg)
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.RedisBackup{}).
		Complete(r)
}
