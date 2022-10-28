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
	"RedisBackupOperator/utils/constants"
	rediswqj "RedisBackupOperator/utils/redis"
	"context"
	"encoding/json"
	"fmt"
	batch "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"reflect"
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
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
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

	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{Name: backupInstance.Spec.ClusterName + "-redis-auth", Namespace: req.Namespace}
	if err := r.Get(ctx, secretKey, secret); err != nil {
		msg := fmt.Sprintf("get redisfailover cluster secret %s/%s failed", req.Namespace, backupInstance.Spec.ClusterName+"-redis-auth")
		log.Log.Error(err, msg)
		return ctrl.Result{Requeue: true}, err
	}
	redisPass := string(secret.Data["password"])

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
			if podlist.Items[k].ObjectMeta.Labels["redisfailovers-role"] == "master" {
				if v.Status.Phase == "Running" {
					backupPvc = v.Spec.Volumes[0].PersistentVolumeClaim.ClaimName
					msg := fmt.Sprintf("Cluster %s: Backup on Master Pod %s,Pvc is %s", backupInstance.Spec.ClusterName, v.ObjectMeta.Name, v.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
					log.Log.Info(msg)
					break
				} else {
					msg := fmt.Sprintf("Cluster %s: No Running Pod to Backup!", backupInstance.Spec.ClusterName)
					log.Log.Info(msg)
					return ctrl.Result{}, nil
				}
			}
		}

		if backupPvc == "" {
			msg := fmt.Sprintf("Cluster %s: No master found!", backupInstance.Spec.ClusterName)
			log.Log.Info(msg)
			return ctrl.Result{}, nil
		}

		// 判断是否周期执行
		if len(backupInstance.Spec.BackupSchedule) == 0 {
			// 如果BackupSchedule为空，则创建单次备份JOB
			backupJob := &batch.Job{}
			jobName := backupInstance.Name + "-" + backupInstance.Spec.ClusterName + "-job"
			backupJobGet := types.NamespacedName{
				Namespace: req.Namespace,
				Name:      jobName,
			}

			keyName := backupInstance.Name
			hostPort := "rfs-" + backupInstance.Spec.ClusterName + ":26379"

			err = r.Client.Get(ctx, backupJobGet, backupJob)
			if errors.IsNotFound(err) {
				job := utils.MakeCronJob(keyName, backupInstance.Spec.BackupSchedule, "", backupPvc, backupInstance.Spec.ClusterName, *backupInstance)
				if err := controllerutil.SetControllerReference(backupInstance, job, r.Scheme); err != nil {
					msg := fmt.Sprintf("set controllerReference for Job %s/%s failed", req.Namespace, jobName)
					log.Log.Error(err, msg)
					return ctrl.Result{Requeue: true}, err
				}

				msg := fmt.Sprintf("create job %s/%s", req.Namespace, jobName)
				log.Log.Info(msg)
				err = r.Client.Create(context.TODO(), job)
				if err != nil {
					msg := fmt.Sprintf("create job %s/%s error!", req.Namespace, jobName)
					log.Log.Error(err, msg)
					return ctrl.Result{Requeue: true}, err
				}
			} else if err != nil {
				msg := fmt.Sprintf("failed to get job %s/%s error!", req.Namespace, jobName)
				log.Log.Error(err, msg)
				return ctrl.Result{Requeue: true}, err
			} else {
				if !backupJob.ObjectMeta.CreationTimestamp.IsZero() {
					res := rediswqj.RedisConn(keyName, hostPort, redisPass)

					if len(res) > 0 {
						var condition []operatorv1alpha1.BackupCondition
						unmarshalErr := json.Unmarshal([]byte(res), &condition)
						if unmarshalErr != nil {
							log.Log.Error(unmarshalErr, "unmarshal json failed")
						}

						backupInstance.Status.Conditions = condition
						err := r.Status().Update(context.TODO(), backupInstance)

						if err != nil {
							fmt.Println(err)
							return ctrl.Result{}, err
						}
					}
					//if !backupJob.ObjectMeta.CreationTimestamp.IsZero() {
					//		if len(backupInstance.Status.Conditions) == 0 || (len(backupInstance.Status.Conditions) > 0 && condition[0].ObjectName != backupInstance.Status.Conditions[0].ObjectName) {
					//			}
					//	}
				}
				return ctrl.Result{}, nil
			}
		} else {
			var backupCronJob client.Object
			// 判断Cronjob版本是v1 or v1beta1
			cronJobVersion := getCronJobVersion()
			if cronJobVersion == constants.CronJobVersionV1beta1 {
				// cronjob版本为v1beta1，则使用v1beta1 APi
				backupCronJob = &batchv1beta1.CronJob{}
			} else {
				// cronjob版本为v1，使用v1 API
				backupCronJob = &batch.CronJob{}
			}

			cronJobName := backupInstance.Name + "-" + backupInstance.Spec.ClusterName + "-cronjob"
			backupCronJobGet := types.NamespacedName{
				Namespace: req.Namespace,
				Name:      cronJobName,
			}

			keyName := "backupCronjob"
			hostPort := "rfs-" + backupInstance.Spec.ClusterName + ":26379"

			err = r.Client.Get(ctx, backupCronJobGet, backupCronJob)
			//var found interface{}
			if errors.IsNotFound(err) {
				cronjob := utils.MakeCronJob(keyName, backupInstance.Spec.BackupSchedule, cronJobVersion, backupPvc, backupInstance.Spec.ClusterName, *backupInstance)
				if err := controllerutil.SetControllerReference(backupInstance, cronjob, r.Scheme); err != nil {
					msg := fmt.Sprintf("set controllerReference for CronJob %s/%s failed", req.Namespace, cronJobName)
					log.Log.Error(err, msg)
					return ctrl.Result{Requeue: true}, err
				}

				msg := fmt.Sprintf("create cronjob %s/%s", req.Namespace, cronJobName)
				log.Log.Info(msg)
				err = r.Client.Create(context.TODO(), cronjob)
				if err != nil {
					msg := fmt.Sprintf("create cronjob %s/%s error!", req.Namespace, cronJobName)
					log.Log.Error(err, msg)
					return ctrl.Result{Requeue: true}, err
				}
			} else if err != nil {
				msg := fmt.Sprintf("failed to get cronjob %s/%s error!", req.Namespace, cronJobName)
				log.Log.Error(err, msg)
				return ctrl.Result{Requeue: true}, err
			} else {
				cronjob := utils.MakeCronJob(keyName, backupInstance.Spec.BackupSchedule, cronJobVersion, backupPvc, backupInstance.Spec.ClusterName, *backupInstance)
				if cronJobVersion == constants.CronJobVersionV1beta1 {
					_, found, err := utils.ExistCronJobV1beta1(cronJobName, req.Namespace, r.Client)
					if err != nil {
						msg := fmt.Sprintf("found cronjob %s/%s error!", req.Namespace, cronJobName)
						log.Log.Error(err, msg)
						return ctrl.Result{Requeue: true}, err
					}
					backupPvcOld := found.Spec.JobTemplate.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName
					if !reflect.DeepEqual(backupPvc, backupPvcOld) {
						msg := fmt.Sprintf("rebuild cronjob %s/%s", req.Namespace, cronJobName)
						log.Log.Info(msg)
						if err := r.Client.Delete(context.TODO(), cronjob, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
							msg := fmt.Sprintf("rebuild cronjob %s/%s error", req.Namespace, cronJobName)
							log.Log.Error(err, msg)
							return ctrl.Result{Requeue: true}, err
						}
						return ctrl.Result{}, nil
					}
				} else {
					_, found, err := utils.ExistCronJob(cronJobName, req.Namespace, r.Client)
					if err != nil {
						msg := fmt.Sprintf("found cronjob %s/%s error!", req.Namespace, cronJobName)
						log.Log.Error(err, msg)
						return ctrl.Result{Requeue: true}, err
					}
					backupPvcOld := found.Spec.JobTemplate.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName
					if !reflect.DeepEqual(backupPvc, backupPvcOld) {
						msg := fmt.Sprintf("rebuild cronjob %s/%s", req.Namespace, cronJobName)
						log.Log.Info(msg)
						if err := r.Client.Delete(context.TODO(), cronjob, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
							msg := fmt.Sprintf("rebuild cronjob %s/%s error", req.Namespace, cronJobName)
							log.Log.Error(err, msg)
							return ctrl.Result{Requeue: true}, err
						}
						return ctrl.Result{}, nil
					}
				}

				res := rediswqj.RedisConn(keyName, hostPort, redisPass)
				if len(res) > 0 {
					var condition []operatorv1alpha1.BackupCondition
					unmarshalErr := json.Unmarshal([]byte(res), &condition)
					if unmarshalErr != nil {
						log.Log.Error(unmarshalErr, "unmarshal json failed")
					}

					if len(backupInstance.Status.Conditions) == 0 ||
						(len(backupInstance.Status.Conditions) > 0 &&
							condition[0].FilePath != backupInstance.Status.Conditions[len(backupInstance.Status.Conditions)-1].FilePath) {

						backupInstance.Status.Conditions = append(backupInstance.Status.Conditions, condition[0])
						err := r.Status().Update(context.TODO(), backupInstance)

						if err != nil {
							fmt.Println(err)
							return ctrl.Result{}, err
						}
					}
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
			if singleAPIResource.Kind == constants.CronJobKind {
				version = groupVersion
			}
		}
	}
	return version
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	cronJobVersion := getCronJobVersion()
	if cronJobVersion == constants.CronJobVersionV1beta1 {
		return ctrl.NewControllerManagedBy(mgr).
			For(&operatorv1alpha1.RedisBackup{}).
			Owns(&batchv1beta1.CronJob{}).
			Owns(&batch.Job{}).
			Complete(r)
	} else {
		return ctrl.NewControllerManagedBy(mgr).
			For(&operatorv1alpha1.RedisBackup{}).
			Owns(&batch.CronJob{}).
			Owns(&batch.Job{}).
			Complete(r)
	}
}
