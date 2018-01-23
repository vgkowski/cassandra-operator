/*
Copyright 2017 The Rook Authors. All rights reserved.

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

// Package v1alpha1 for a sample crd
package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import _ "k8s.io/code-generator"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CassandraCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec CassandraClusterSpec `json:"spec"`
}

type CassandraClusterSpec struct {
	Cpu string `json:"cpu"`
	Ram string `json:"ram"`
	Data Storage `json:"data"`
	Commit Storage `json:"commit"`
	NbNodes int32 `json:"nbNodes"`
	RackLabel string `json:"rackLabel"`
	DCLabel string `json:"dcLabel"`
	Spec CassandraSpec `json:"spec"`
}

type Storage struct {
	StorageVolume string `json:"storageVolume"`
	StorageClass string `json:"storageClass"`
}

type CassandraSpec struct {
	NbToken int `json:"nbToken"`
	MaxHeapSize string `json:"maxHeapSize"`
	HeapNewSize string `json:"heapNewSize"`
	InterNodeTLS bool `json:"interNodeTLS"`
	ClientTLS bool `json:"clientTLS"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CassandraClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CassandraCluster `json:"items"`
}
