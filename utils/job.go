package utils

import (
	operatorv1alpha1 "RedisBackupOperator/api/v1alpha1"
	"RedisBackupOperator/utils/constants"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

func MakeJob(backupPvc string, backupInstance operatorv1alpha1.RedisBackup) *batch.Job {

	backupName := time.Now().Format("20060102150405") + ".tar.gz"
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
	container.Image = "172.16.5.171/mysql/mysql-operator-sidecar-5.7:v1.2.2"
	//container.Env = append(initcontainer.Env, corev1.EnvVar{Name: "registrypassword", Value: buildInstance.Spec.Registry.Password})
	container.Command = append(container.Command, "/bin/bash", "-c")
	shell := "tar -zcvf " + backupName + " *.rdb;"
	rcloneShell := "rclone --config /tmp/rclone.conf moveto " + backupName + " s3://velero/" + backupName + " --s3-access-key-id " + constants.S3AccessKeyId + " --s3-secret-access-key " + constants.S3SecretAccessKey + " --s3-region " + constants.S3Region + " --s3-endpoint " + constants.S3Endpoint + " --s3-provider " + constants.S3Provider + ";"
	container.Args = append(container.Args, shell+rcloneShell)
	log.Log.Info("backup create succeed: " + backupName)
	log.Log.Info("rclone command is: " + shell + rcloneShell)
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
