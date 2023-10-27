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
	"encoding/base64"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"

	operatorv1alpha1 "RedisBackupOperator/api/v1alpha1"
	redis "RedisBackupOperator/utils/redis"
)

// RedisCloneReconciler reconciles a RedisClone object
type RedisCloneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewRedisCloneReconciler builds and return new NewRedisCloneReconciler instance
func NewRedisCloneReconciler(client client.Client, Scheme *runtime.Scheme) *RedisCloneReconciler {
	RedisSyncCloneReconciler := &RedisCloneReconciler{
		client,
		Scheme,
	}

	return RedisSyncCloneReconciler
}

//+kubebuilder:rbac:groups=operator.bobfintech.com,resources=redisclones,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.bobfintech.com,resources=redisclones/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.bobfintech.com,resources=redisclones/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RedisClone object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *RedisCloneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	cloneInstance := &operatorv1alpha1.RedisClone{}
	err := r.Client.Get(ctx, req.NamespacedName, cloneInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			msg := fmt.Sprintf("RedisClone clone CR [%s] not found. Might be deleted.", req.NamespacedName)
			log.Log.Info(msg)
			return reconcile.Result{}, nil
		}
		msg := fmt.Sprintf("RedisClone unable to get clone CR %s: %v", req.NamespacedName, err)
		log.Log.Info(msg)
		return reconcile.Result{}, nil
	}

	cloneConditions := operatorv1alpha1.CloneCondition{
		SourceCluster: cloneInstance.Spec.SourceCluster,
		DestCluster:   cloneInstance.Spec.DestCluster,
		SourceIPPort:  cloneInstance.Spec.SourceIP + ":" + cloneInstance.Spec.SourcePort,
		ClusterType:   cloneInstance.Spec.ClusterType,
		StartTime:     getCurrentTime(),
		OffSet:        "",
		CloneLag:      "",
		CloneStatus:   "down",
		Replicas:      cloneInstance.Spec.Replicas,
	}

	replicationInfo := make(map[string]string)
	//获取目标集群DestCluster节点Ip
	podlist := corev1.PodList{}
	error := r.Client.List(ctx, &podlist, client.MatchingLabels{
		"app.kubernetes.io/component":  "redis",
		"app.kubernetes.io/managed-by": "redis-operator",
		"app.kubernetes.io/name":       cloneInstance.Spec.DestCluster,
		"app.kubernetes.io/part-of":    "redis-failover"})
	if error != nil {
		msg := fmt.Sprintf("RedisClone Get Dest Redis pod error! %s:%s", req.NamespacedName, cloneInstance.Spec.DestCluster)
		log.Log.Error(error, msg)
		return ctrl.Result{}, error
	}

	if len(podlist.Items) == 0 {
		msg := fmt.Sprintf("RedisClone can't find Dest Redis pods! %s:%s", req.NamespacedName, cloneInstance.Spec.DestCluster)
		log.Log.Info(msg)
	} else {
		DestPodIP := podlist.Items[0].Status.PodIP
		//解析源端SourceCluster密码
		redisPass, _ := base64.StdEncoding.DecodeString(cloneInstance.Spec.SourcePasswd)
		//获取复制信息
		infoReplication, _ := redis.SlaveIsReady(DestPodIP, "6379", string(redisPass))

		if infoReplication != "" {
			//过滤目标端DestPod的复制信息
			lines := strings.Split(infoReplication, "\n")
			for _, line := range lines {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					replicationInfo[parts[0]] = parts[1]
				}
			}
			cloneConditions.OffSet = replicationInfo["slave_repl_offset"]
			cloneConditions.CloneLag = replicationInfo["master_last_io_seconds_ago"]
			cloneConditions.CloneStatus = replicationInfo["master_link_status"]
		}
	}
	//更新clone CR的Conditions
	cloneInstance.Status.Conditions = cloneConditions
	err = r.Status().Update(context.TODO(), cloneInstance)
	if err != nil {
		msg := fmt.Sprintf("RedisClone update CR conditions %s: %v", req.NamespacedName, err)
		log.Log.Info(msg)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *RedisCloneReconciler) CloneInfoSync(ctx context.Context, namespace, name string) error {
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	cloneInstance := &operatorv1alpha1.RedisClone{}
	err := r.Client.Get(ctx, namespacedName, cloneInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			msg := fmt.Sprintf("CloneInfoSync clone CR [%s] not found. Might be deleted.", namespacedName)
			log.Log.Info(msg)
			return nil
		}
		msg := fmt.Sprintf("CloneInfoSync unable to get clone CR %s: %v", namespacedName, err)
		log.Log.Info(msg)
		return nil
	}

	cloneConditions := cloneInstance.Status.Conditions
	if cloneConditions.StartTime == "" {
		cloneConditions.SourceCluster = cloneInstance.Spec.SourceCluster
		cloneConditions.DestCluster = cloneInstance.Spec.DestCluster
		cloneConditions.Replicas = cloneInstance.Spec.Replicas
		cloneConditions.ClusterType = cloneInstance.Spec.ClusterType
		cloneConditions.SourceIPPort = cloneInstance.Spec.SourceIP + ":" + cloneInstance.Spec.SourcePort
		cloneConditions.StartTime = getCurrentTime()
	}
	replicationInfo := make(map[string]string)
	//获取目标集群DestCluster节点Ip
	podlist := corev1.PodList{}
	error := r.Client.List(ctx, &podlist, client.MatchingLabels{
		"app.kubernetes.io/component":  "redis",
		"app.kubernetes.io/managed-by": "redis-operator",
		"app.kubernetes.io/name":       cloneInstance.Spec.DestCluster,
		"app.kubernetes.io/part-of":    "redis-failover"})
	if error != nil {
		msg := fmt.Sprintf("CloneInfoSync Get Dest Redis pod error! %s:%s", namespacedName, cloneInstance.Spec.DestCluster)
		log.Log.Error(error, msg)
		return error
	}

	if len(podlist.Items) == 0 {
		cloneConditions.CloneLag = "-1"
		cloneConditions.CloneStatus = "down"
		msg := fmt.Sprintf("CloneInfoSync can't find Dest Redis pods! %s:%s", namespacedName, cloneInstance.Spec.DestCluster)
		log.Log.Info(msg)
	} else {
		DestPodIP := podlist.Items[0].Status.PodIP
		//解析源端SourceCluster密码
		redisPass, _ := base64.StdEncoding.DecodeString(cloneInstance.Spec.SourcePasswd)
		//获取复制信息
		infoReplication, _ := redis.SlaveIsReady(DestPodIP, "6379", string(redisPass))

		if infoReplication != "" {
			//过滤目标端DestPod的复制信息
			lines := strings.Split(infoReplication, "\n")
			for _, line := range lines {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					replicationInfo[parts[0]] = parts[1]
				}
			}
			cloneConditions.OffSet = replicationInfo["slave_repl_offset"]
			cloneConditions.CloneLag = replicationInfo["master_last_io_seconds_ago"]
			cloneConditions.CloneStatus = replicationInfo["master_link_status"]
		} else {
			cloneConditions.CloneLag = "-1"
			cloneConditions.CloneStatus = "down"
		}
	}
	//更新clone CR的Conditions
	cloneInstance.Status.Conditions = cloneConditions
	err = r.Status().Update(context.TODO(), cloneInstance)
	if err != nil {
		if strings.Contains(err.Error(), "the object has been modified") {
			return nil
		} else {
			msg := fmt.Sprintf("CloneInfoSync update CR conditions %s: %v", namespacedName, err)
			log.Log.Info(msg)
		}
	}
	return nil
}

func getCurrentTime() string {
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc).Format(timeLayout)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisCloneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.RedisClone{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 5}).
		Complete(r)
}
