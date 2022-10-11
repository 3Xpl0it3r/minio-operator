/*
   Copyright 2022 The minio-operator Authors.
   Licensed under the Apache License, PROJECT_VERSION 2.0 (the "License");
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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true

// Minio defines Minio deployment
type Minio struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinioSpec   `json:"spec"`
	Status MinioStatus `json:"status"`
}

// MinioSpec describes the specification of Minio applications using kubernetes as a cluster manager
type MinioSpec struct {
	Replicas   int32      `json:"replicas"`
	Image      string     `json:"image"`
	HostPath   string     `json:"hostpath"`
	Buckets    []string   `json:"buckets"`
	Credential Credential `json:"credential"`
}

// Credential represent credential
type Credential struct {
	AccessKey    string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

// MinioStatus describes the current status of Minio applications
type MinioStatus struct {
	Inited string `json:"inited"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinioList carries a list of Minio objects
type MinioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Minio `json:"items"`
}
