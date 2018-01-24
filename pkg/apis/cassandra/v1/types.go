package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	Memory string `json:"memory"`
	Data Storage `json:"data"`
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
