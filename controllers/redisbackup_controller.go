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
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	cronJobKind           = "CronJob"
	cronJobVersionV1beta1 = "batch/v1beta1"
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

	podlist := corev1.PodList{}
	backupPvc := ""
	error := r.Client.List(ctx, &podlist, client.MatchingLabels{
		"app.kubernetes.io/component":  "redis",
		"app.kubernetes.io/managed-by": "redis-operator",
		"app.kubernetes.io/name":       backupInstance.Spec.ClusterName,
		"app.kubernetes.io/part-of":    "redis-failover"})
	if error != nil {
		msg := fmt.Sprintf("Get Redis pod error!")
		log.Log.Error(error, msg)
	} else {
		for k, v := range podlist.Items {
			if podlist.Items[k].ObjectMeta.Labels["redisfailovers-role"] == "slave" {
				if v.Status.Phase == "Running" {
					backupPvc = v.Spec.Volumes[0].PersistentVolumeClaim.ClaimName
					msg := fmt.Sprintf("Backup Slave Pod is %s,Pvc is %s", v.ObjectMeta.Name, v.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
					log.Log.Info(msg)
					break
				} else {
					backupPvc = ""
					msg := fmt.Sprintf("Slave Pod %s is not running", v.ObjectMeta.Name)
					log.Log.Info(msg)
				}
			}
		}
	}
	if backupPvc == "" {
		for k, v := range podlist.Items {
			if podlist.Items[k].ObjectMeta.Labels["redisfailovers-role"] == "master" {
				if v.Status.Phase == "Running" {
					backupPvc = v.Spec.Volumes[0].PersistentVolumeClaim.ClaimName
					msg := fmt.Sprintf("Backup on Master Pod %s,Pvc is %s", v.ObjectMeta.Name, v.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
					log.Log.Info(msg)
					break
				} else {
					msg := fmt.Sprintf("No Running Pod to Backup!")
					log.Log.Info(msg)
				}
			}
		}
	}

	// 判断是否周期执行
	if len(backupInstance.Spec.BackupSchedule) == 0 {
		// 如果BackupSchedule为空，则创建单次备份JOB
		backupJob := &batch.Job{}
		backupJobGet := types.NamespacedName{
			Namespace: req.Namespace,
			Name:      backupInstance.Name + "-job",
		}

		err = r.Client.Get(ctx, backupJobGet, backupJob)
		if errors.IsNotFound(err) {
			job := utils.MakeJob(backupPvc, backupInstance.Spec.ClusterName, *backupInstance)
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
	} else {
		// 判断Cronjob版本是v1 or v1beta1
		cronJobVersion := getCronJobVersion()
		if cronJobVersion == cronJobVersionV1beta1 {

			// cronjob版本为v1beta1，则使用v1beta1 APi
			backupCronJob := &batchv1beta1.CronJob{}
			backupCronJobGet := types.NamespacedName{
				Namespace: req.Namespace,
				Name:      backupInstance.Name + "-cronjob",
			}
			err = r.Client.Get(ctx, backupCronJobGet, backupCronJob)
			if errors.IsNotFound(err) {
				cronjob := utils.MakeCronJobV1beta1(backupPvc, backupInstance.Spec.ClusterName, *backupInstance)
				if err := controllerutil.SetControllerReference(backupInstance, cronjob, r.Scheme); err != nil {
					msg := fmt.Sprintf("set controllerReference for CronJob %s/%s failed", cronjob.Namespace, cronjob.Name)
					log.Log.Error(err, msg)
					return ctrl.Result{Requeue: true}, err
				}

				msg := fmt.Sprintf("create cronjob %s/%s", cronjob.Namespace, cronjob.Name)
				log.Log.Info(msg)
				err = r.Client.Create(context.TODO(), cronjob)
				if err != nil {
					msg := fmt.Sprintf("create cronjob %s/%s error!", cronjob.Namespace, cronjob.Name)
					log.Log.Error(err, msg)
					return ctrl.Result{Requeue: true}, err
				}
			}
		} else {
			// cronjob版本为v1，使用v1 API
			backupCronJob := &batch.CronJob{}
			backupCronJobGet := types.NamespacedName{
				Namespace: req.Namespace,
				Name:      backupInstance.Name + "-cronjob",
			}

			err = r.Client.Get(ctx, backupCronJobGet, backupCronJob)
			if errors.IsNotFound(err) {
				cronjob := utils.MakeCronJob(backupPvc, backupInstance.Spec.ClusterName, *backupInstance)
				if err := controllerutil.SetControllerReference(backupInstance, cronjob, r.Scheme); err != nil {
					msg := fmt.Sprintf("set controllerReference for CronJob %s/%s failed", cronjob.Namespace, cronjob.Name)
					log.Log.Error(err, msg)
					return ctrl.Result{Requeue: true}, err
				}

				msg := fmt.Sprintf("create cronjob %s/%s", cronjob.Namespace, cronjob.Name)
				log.Log.Info(msg)
				err = r.Client.Create(context.TODO(), cronjob)
				if err != nil {
					msg := fmt.Sprintf("create cronjob %s/%s error!", cronjob.Namespace, cronjob.Name)
					log.Log.Error(err, msg)
					return ctrl.Result{Requeue: true}, err
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

func getCronJobVersion() string {
	var version string
	config := ctrl.GetConfigOrDie()
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Log.Error(err, "get discoveryClient failed")
	}
	// 获取所有分组和资源数据
	APIResourceListSlice, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		log.Log.Error(err, "get APIResources failed")
	}
	// APIResourceListSlice是个切片，里面的每个元素代表一个GroupVersion及其资源
	for _, singleAPIResourceList := range APIResourceListSlice {
		groupVersion := singleAPIResourceList.GroupVersion
		// APIResources字段是个切片，里面是当前GroupVersion下的所有资源
		for _, singleAPIResource := range singleAPIResourceList.APIResources {
			if singleAPIResource.Kind == cronJobKind {
				version = groupVersion
			}
		}
	}
	return version
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.RedisBackup{}).
		Complete(r)
}
