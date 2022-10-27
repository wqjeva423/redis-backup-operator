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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisBackupSpec defines the desired state of RedisBackup
type RedisBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterName string `json:"clusterName,omitempty"`

	BackupSecretName string `json:"backupSecretName,omitempty"`

	ExpireDays string `json:"expireDays,omitempty"`

	BackupURL string `json:"backupURL,omitempty"`

	BackupSchedule string `json:"backupSchedule,omitempty"`
}

// BackupCondition defines condition struct for backup resource
type BackupCondition struct {
	//// type of cluster condition, values in (\"Ready\")
	//Type BackupConditionType `json:"type"`
	//// Status of the condition, one of (\"True\", \"False\", \"Unknown\")
	//Status corev1.ConditionStatus `json:"status"`
	//
	//// LastTransitionTime
	//LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	//// Reason
	//Reason string `json:"reason"`
	//// Message
	//Message string `json:"message"`
	S3BackupURL     string `json:"s3backupURL"`
	FileName        string `json:"fileName"`
	FileSize        string `json:"fileSize"`
	JobName         string `json:"jobName"`
	PodName         string `json:"podName"`
	ObjectName      string `json:"objectName"`
	BackupStartTime string `json:"backupStartTime"`
	BackupEndTime   string `json:"backupEndTime"`
	Status          string `json:"status"`
	Message         string `json:"message"`
}

// BackupConditionType defines condition types of a backup resources
type BackupConditionType string

const (
	// BackupComplete means the backup has finished his execution
	BackupComplete BackupConditionType = "Complete"
	// BackupFailed means backup has failed
	BackupFailed BackupConditionType = "Failed"
)

// RedisBackupStatus defines the observed state of RedisBackup
type RedisBackupStatus struct {
	JobStatus []JobStatus `json:"jobStatus,omitempty"`
	// Conditions represents the backup resource conditions list.
	Conditions []BackupCondition `json:"conditions,omitempty"`
}

type JobStatus struct {
	// +kubebuilder:default:=false
	Completed bool `json:"Completed,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisBackup is the Schema for the redisbackups API
type RedisBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisBackupSpec   `json:"spec,omitempty"`
	Status RedisBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisBackupList contains a list of RedisBackup
type RedisBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisBackup{}, &RedisBackupList{})
}
