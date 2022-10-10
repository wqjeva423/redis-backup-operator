package utils

import (
	operatorv1alpha1 "RedisBackupOperator/api/v1alpha1"
	"RedisBackupOperator/utils/constants"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

func MakeJob(backupPvc, clusterName string, backupInstance operatorv1alpha1.RedisBackup) *batch.Job {

	backupName := "redis-" + clusterName + "-" + time.Now().Format("20060102150405") + ".tar.gz"
	backofflimit := int32(1)
	activedeadlineseconds := int64(1800)
	ttlsecondsafterfinished := int32(3600)
	volumedir := corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: backupPvc,
			},
		},
	}

	container := corev1.Container{}
	container.Name = "redis-backup"
	container.Image = "172.16.5.171/redis/s3cmd:latest"
	//container.Image = "d3fk/s3cmd:latest"

	//container.Env = append(initcontainer.Env, corev1.EnvVar{Name: "registrypassword", Value: buildInstance.Spec.Registry.Password})
	boolTrue := true
	container.Env = []corev1.EnvVar{
		{
			Name: "S3Region",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: backupInstance.Spec.BackupSecretName,
					},
					Key:      "S3Region",
					Optional: &boolTrue,
				},
			},
		},
		{
			Name: "S3Endpoint",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: backupInstance.Spec.BackupSecretName,
					},
					Key:      "S3Endpoint",
					Optional: &boolTrue,
				},
			},
		},
		{
			Name: "S3AccessKey",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: backupInstance.Spec.BackupSecretName,
					},
					Key:      "S3AccessKey",
					Optional: &boolTrue,
				},
			},
		},
		{
			Name: "S3SecretKey",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: backupInstance.Spec.BackupSecretName,
					},
					Key:      "S3SecretKey",
					Optional: &boolTrue,
				},
			},
		},
		{
			Name: "S3Bucket",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: backupInstance.Spec.BackupSecretName,
					},
					Key:      "S3Bucket",
					Optional: &boolTrue,
				},
			},
		},
	}

	container.Command = append(container.Command, "sh", "-c")
	shell1 := "cd " + constants.DataDir + ";tar -zcvf /tmp/" + backupName + " *.*;"
	shell2 := "s3cmd --no-ssl --region=$(S3Region) --host=$(S3Endpoint) --host-bucket=$(S3Endpoint) --access_key=$(S3AccessKey) --secret_key=$(S3SecretKey) put /tmp/" + backupName + " s3://$(S3Bucket)/;"
	shell3 := "rm -rf /tmp/" + backupName + ";"
	container.Args = append(container.Args, shell1+shell2+shell3)
	log.Log.Info("backup file is: " + backupName)
	log.Log.Info("rclone command is: " + shell1 + shell2 + shell3)
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{Name: "data", MountPath: "/data"})

	job := &batch.Job{}
	job.Name = backupInstance.Name + "-job"
	job.Namespace = backupInstance.Namespace
	job.Spec.ActiveDeadlineSeconds = &activedeadlineseconds
	job.Spec.BackoffLimit = &backofflimit
	job.Spec.TTLSecondsAfterFinished = &ttlsecondsafterfinished
	job.Spec.Template.Spec.RestartPolicy = "Never"
	job.Spec.Template.Spec.Containers = append(job.Spec.Template.Spec.Containers, container)
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, volumedir)
	//if buildInstance.Spec.NodeSelector != nil {
	//	job.Spec.Template.Spec.NodeSelector = buildInstance.Spec.NodeSelector
	//}

	return job
}
