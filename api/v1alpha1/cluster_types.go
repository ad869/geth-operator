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
	"github.com/ad869/geth-operator/api/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeType uint

const (
	NodeTypeValidator NodeType = 0
	NodeTypeMember    NodeType = 1
)

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {

	// Image is Ethereum node client image
	Image string `json:"image,omitempty"`

	// used to derive ethereum account.
	Mnemonic string `json:"mnemonic"`

	// Genesis is genesis block configuration
	Genesis *Genesis `json:"genesis"`

	// validator is validators of qbft consensus.
	Validator NodeTemplate `json:"validator,omitempty"`

	// member is static nodes.
	Member NodeTemplate `json:"member,omitempty"`

	// node ports
	Ports `json:"ports,omitempty"`
}

type NodeTemplate struct {
	// number of validator
	Number int `json:"number"`

	// Resources is node compute and storage resources
	shared.Resources `json:"resources"`

	// Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail (default: 3)
	Verbosity uint `json:"verbosity,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
