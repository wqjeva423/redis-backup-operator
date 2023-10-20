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

// RedisCloneSpec defines the desired state of RedisClone
type RedisCloneSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	SourceCluster string `json:"sourceCluster,omitempty"`
	DestCluster   string `json:"destCluster,omitempty"`
	SourceIP      string `json:"sourceIP,omitempty"`
	SourcePort    string `json:"sourcePort,omitempty"`
	SourcePasswd  string `json:"sourcePasswd,omitempty"`
	ClusterType   string `json:"clusterType,omitempty"`
	Replicas      int32  `json:"replicas,omitempty"`
}

type CloneCondition struct {
	SourceCluster string `json:"sourceCluster,omitempty"`
	DestCluster   string `json:"destCluster,omitempty"`
	SourceIPPort  string `json:"sourceIPPort,omitempty"`
	ClusterType   string `json:"clusterType,omitempty"`
	StartTime     string `json:"startTime"`
	OffSet        string `json:"offSet"`
	CloneLag      string `json:"cloneLag"`
	CloneStatus   string `json:"cloneStatus"`
	Replicas      int32  `json:"replicas"`
}

// RedisCloneStatus defines the observed state of RedisClone
type RedisCloneStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions CloneCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisClone is the Schema for the redisclones API
type RedisClone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisCloneSpec   `json:"spec,omitempty"`
	Status RedisCloneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisCloneList contains a list of RedisClone
type RedisCloneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisClone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisClone{}, &RedisCloneList{})
}
