package utils

import (
	operatorv1alpha1 "RedisBackupOperator/api/v1alpha1"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func MakeJob(backupInstance operatorv1alpha1.RedisBackup) *batch.Job {

	backofflimit := int32(1)
	activedeadlineseconds := int64(1800)
	ttlsecondsafterfinished := int32(3600)
	volumedir := corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: "redisfailover-persistent-keep-data-rfr-wqj-0",
			},
		},
	}

	container := corev1.Container{}
	container.Name = "redis-backup"
	container.Image = "redis:6.2.6"
	//container.Env = append(initcontainer.Env, corev1.EnvVar{Name: "registrypassword", Value: buildInstance.Spec.Registry.Password})
	container.Command = append(container.Command, "/bin/bash", "-c")
	shell := "echo cao;touch test.txt"
	container.Args = append(container.Args, shell)
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
