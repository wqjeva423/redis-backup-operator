package utils

//func MakeCronJobT(cronJobVersion, backupPvc, clusterName string, backupInstance operatorv1alpha1.RedisBackup) client.Object {
//
//	backupName := "redis-bak-" + clusterName + "-$(date '+%Y%m%d%H%M%S').tar.gz"
//	backupImg := GetEnvDefault("backup_image", "d3fk/s3cmd:latest")
//	backofflimit := int32(1)
//	activedeadlineseconds := int64(1800)
//	ttlsecondsafterfinished := int32(3600)
//	volumedir := corev1.Volume{
//		Name: "data",
//		VolumeSource: corev1.VolumeSource{
//			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
//				ClaimName: backupPvc,
//			},
//		},
//	}
//
//	container := corev1.Container{}
//	container.Name = "redis-backup"
//	container.Image = backupImg
//
//	//container.Env = append(initcontainer.Env, corev1.EnvVar{Name: "registrypassword", Value: buildInstance.Spec.Registry.Password})
//	boolTrue := true
//	container.Env = []corev1.EnvVar{
//		{
//			Name: "S3Region",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3Region",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3Endpoint",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3Endpoint",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3AccessKey",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3AccessKey",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3SecretKey",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3SecretKey",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3Bucket",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3Bucket",
//					Optional: &boolTrue,
//				},
//			},
//		},
//	}
//
//	container.Command = append(container.Command, "sh", "-c")
//	shTar := "cd /data;tar -zcvf /tmp/" + backupName + " *.*;"
//	shS3put := "s3cmd --no-ssl --region=$(S3Region) --host=$(S3Endpoint) --host-bucket=$(S3Endpoint) --access_key=$(S3AccessKey) --secret_key=$(S3SecretKey) put /tmp/redis-bak-* " + backupInstance.Spec.BackupURL + ";"
//	shS3expire := "s3cmd --no-ssl --region=$(S3Region) --host=$(S3Endpoint) --host-bucket=$(S3Endpoint) --access_key=$(S3AccessKey) --secret_key=$(S3SecretKey) expire " + backupInstance.Spec.BackupURL + " --expiry-day=" + backupInstance.Spec.ExpireDays + " --expiry-prefix=redis-bak-" + clusterName + ";"
//	shClean := "rm -rf /tmp/redis-bak-*;"
//	container.Args = append(container.Args, shTar+shS3put+shS3expire+shClean)
//	log.Log.Info("backup file is: " + backupName)
//	//log.Log.Info("rclone command is: " + shTar + shS3put + shS3expire + shClean)
//	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{Name: "data", MountPath: "/data"})
//
//	if cronJobVersion == constants.CronJobVersionV1beta1 {
//		cronJob := &batchv1beta1.CronJob{}
//		cronJob.Name = backupInstance.Name + "-" + backupInstance.Spec.ClusterName + "-cronjob"
//		cronJob.Namespace = backupInstance.Namespace
//		cronJob.Spec.Schedule = backupInstance.Spec.BackupSchedule
//		cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &activedeadlineseconds
//		cronJob.Spec.JobTemplate.Spec.BackoffLimit = &backofflimit
//		cronJob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished = &ttlsecondsafterfinished
//		cronJob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = "Never"
//		cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers = append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers, container)
//		cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes, volumedir)
//		//if buildInstance.Spec.NodeSelector != nil {
//		//	job.Spec.Template.Spec.NodeSelector = buildInstance.Spec.NodeSelector
//		//}
//
//		return cronJob
//	}else{
//		cronJob := &batch.CronJob{}
//		cronJob.Name = backupInstance.Name + "-" + backupInstance.Spec.ClusterName + "-cronjob"
//		cronJob.Namespace = backupInstance.Namespace
//		cronJob.Spec.Schedule = backupInstance.Spec.BackupSchedule
//		cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &activedeadlineseconds
//		cronJob.Spec.JobTemplate.Spec.BackoffLimit = &backofflimit
//		cronJob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished = &ttlsecondsafterfinished
//		cronJob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = "Never"
//		cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers = append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers, container)
//		cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes, volumedir)
//
//		return cronJob
//	}
//}
//
//func MakeCronJob(backupPvc, clusterName string, backupInstance operatorv1alpha1.RedisBackup) *batch.CronJob {
//
//	backupName := "redis-bak-" + clusterName + "-$(date '+%Y%m%d%H%M%S').tar.gz"
//	backupImg := GetEnvDefault("backup_image", "d3fk/s3cmd:latest")
//	backofflimit := int32(1)
//	activedeadlineseconds := int64(1800)
//	ttlsecondsafterfinished := int32(3600)
//	volumedir := corev1.Volume{
//		Name: "data",
//		VolumeSource: corev1.VolumeSource{
//			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
//				ClaimName: backupPvc,
//			},
//		},
//	}
//
//	container := corev1.Container{}
//	container.Name = "redis-backup"
//	container.Image = backupImg
//
//	//container.Env = append(initcontainer.Env, corev1.EnvVar{Name: "registrypassword", Value: buildInstance.Spec.Registry.Password})
//	boolTrue := true
//	container.Env = []corev1.EnvVar{
//		{
//			Name: "S3Region",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3Region",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3Endpoint",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3Endpoint",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3AccessKey",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3AccessKey",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3SecretKey",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3SecretKey",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3Bucket",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3Bucket",
//					Optional: &boolTrue,
//				},
//			},
//		},
//	}
//
//	container.Command = append(container.Command, "sh", "-c")
//	shTar := "cd /data;tar -zcvf /tmp/" + backupName + " *.*;"
//	shS3put := "s3cmd --no-ssl --region=$(S3Region) --host=$(S3Endpoint) --host-bucket=$(S3Endpoint) --access_key=$(S3AccessKey) --secret_key=$(S3SecretKey) put /tmp/redis-bak-* " + backupInstance.Spec.BackupURL + ";"
//	shS3expire := "s3cmd --no-ssl --region=$(S3Region) --host=$(S3Endpoint) --host-bucket=$(S3Endpoint) --access_key=$(S3AccessKey) --secret_key=$(S3SecretKey) expire " + backupInstance.Spec.BackupURL + " --expiry-day=" + backupInstance.Spec.ExpireDays + " --expiry-prefix=redis-bak-" + clusterName + ";"
//	shClean := "rm -rf /tmp/redis-bak-*;"
//	container.Args = append(container.Args, shTar+shS3put+shS3expire+shClean)
//	log.Log.Info("backup file is: " + backupName)
//	//log.Log.Info("rclone command is: " + shTar + shS3put + shS3expire + shClean)
//	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{Name: "data", MountPath: "/data"})
//
//	cronJob := &batch.CronJob{}
//	cronJob.Name = backupInstance.Name + "-" + backupInstance.Spec.ClusterName + "-cronjob"
//	cronJob.Namespace = backupInstance.Namespace
//	cronJob.Spec.Schedule = backupInstance.Spec.BackupSchedule
//	cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &activedeadlineseconds
//	cronJob.Spec.JobTemplate.Spec.BackoffLimit = &backofflimit
//	cronJob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished = &ttlsecondsafterfinished
//	cronJob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = "Never"
//	cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers = append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers, container)
//	cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes, volumedir)
//	//if buildInstance.Spec.NodeSelector != nil {
//	//	job.Spec.Template.Spec.NodeSelector = buildInstance.Spec.NodeSelector
//	//}
//
//	return cronJob
//}
//
//func MakeCronJobV1beta1(backupPvc, clusterName string, backupInstance operatorv1alpha1.RedisBackup) *batchv1beta1.CronJob {
//
//	backupName := "redis-bak-" + clusterName + "-$(date '+%Y%m%d%H%M%S').tar.gz"
//	backupImg := GetEnvDefault("backup_image", "172.16.5.171/redis/s3cmd:latest")
//	backofflimit := int32(1)
//	activedeadlineseconds := int64(1800)
//	ttlsecondsafterfinished := int32(3600)
//	volumedir := corev1.Volume{
//		Name: "data",
//		VolumeSource: corev1.VolumeSource{
//			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
//				ClaimName: backupPvc,
//			},
//		},
//	}
//
//	container := corev1.Container{}
//	container.Name = "redis-backup"
//	container.Image = backupImg
//
//	//container.Env = append(initcontainer.Env, corev1.EnvVar{Name: "registrypassword", Value: buildInstance.Spec.Registry.Password})
//	boolTrue := true
//	container.Env = []corev1.EnvVar{
//		{
//			Name: "S3Region",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3Region",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3Endpoint",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3Endpoint",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3AccessKey",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3AccessKey",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3SecretKey",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3SecretKey",
//					Optional: &boolTrue,
//				},
//			},
//		},
//		{
//			Name: "S3Bucket",
//			ValueFrom: &corev1.EnvVarSource{
//				SecretKeyRef: &corev1.SecretKeySelector{
//					LocalObjectReference: corev1.LocalObjectReference{
//						Name: backupInstance.Spec.BackupSecretName,
//					},
//					Key:      "S3Bucket",
//					Optional: &boolTrue,
//				},
//			},
//		},
//	}
//
//	container.Command = append(container.Command, "sh", "-c")
//	shTar := "cd /data;tar -zcvf /tmp/" + backupName + " *.*;"
//	shS3put := "s3cmd --no-ssl --region=$(S3Region) --host=$(S3Endpoint) --host-bucket=$(S3Endpoint) --access_key=$(S3AccessKey) --secret_key=$(S3SecretKey) put /tmp/redis-bak-* " + backupInstance.Spec.BackupURL + ";"
//	shS3expire := "s3cmd --no-ssl --region=$(S3Region) --host=$(S3Endpoint) --host-bucket=$(S3Endpoint) --access_key=$(S3AccessKey) --secret_key=$(S3SecretKey) expire " + backupInstance.Spec.BackupURL + " --expiry-day=" + backupInstance.Spec.ExpireDays + " --expiry-prefix=redis-bak-" + clusterName + ";"
//	shClean := "rm -rf /tmp/redis-bak-*;"
//	container.Args = append(container.Args, shTar+shS3put+shS3expire+shClean)
//	log.Log.Info("backup file is: " + backupName)
//	//log.Log.Info("rclone command is: " + shTar + shS3put + shS3expire + shClean)
//	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{Name: "data", MountPath: "/data"})
//
//	cronJob := &batchv1beta1.CronJob{}
//	cronJob.Name = backupInstance.Name + "-" + backupInstance.Spec.ClusterName + "-cronjob"
//	cronJob.Namespace = backupInstance.Namespace
//	cronJob.Spec.Schedule = backupInstance.Spec.BackupSchedule
//	cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &activedeadlineseconds
//	cronJob.Spec.JobTemplate.Spec.BackoffLimit = &backofflimit
//	cronJob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished = &ttlsecondsafterfinished
//	cronJob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = "Never"
//	cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers = append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers, container)
//	cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes, volumedir)
//	//if buildInstance.Spec.NodeSelector != nil {
//	//	job.Spec.Template.Spec.NodeSelector = buildInstance.Spec.NodeSelector
//	//}
//
//	return cronJob
//}
