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

// GoQuorumHomeDir is goquorum home directory
const GoQuorumHomeDir = "/home/quorum"

type Enode string

// NodeSpec defines the desired state of Node
type NodeSpec struct {

	// Image is Ethereum node client image
	Image string `json:"image,omitempty"`

	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`

	// StaticNodes is a set of ethereum nodes to maintain connection to
	// +listType=set
	StaticNodes []Enode `json:"staticNodes,omitempty"`

	// Genesis is genesis block configuration
	Genesis *Genesis `json:"genesis,omitempty"`

	// NodePrivateKeySecretName is the secret name holding node private key
	NodePrivateKeySecretName string `json:"nodePrivateKeySecretName,omitempty"`

	// Miner is whether node is mining/validating blocks or no
	Miner bool `json:"miner,omitempty"`

	// import is account to import
	Import *ImportedAccount `json:"import,omitempty"`

	// Coinbase is the account to which mining rewards are paid
	Coinbase EthereumAddress `json:"coinbase,omitempty"`

	// Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail (default: 3)
	Verbosity uint `json:"verbosity,omitempty"`

	// P2PPort is port used for peer to peer communication
	P2PPort uint `json:"p2pPort,omitempty"`

	// RPCPort is HTTP-RPC server listening port
	RPCPort uint `json:"rpcPort,omitempty"`

	// WSPort is the web socket server listening port
	WSPort uint `json:"wsPort,omitempty"`

	// GraphQLPort is the GraphQL server listening port
	GraphQLPort uint `json:"graphqlPort,omitempty"`

	RLPXPort uint `json:"rlpxPort,omitempty"`

	MetricsPort uint `json:"metricsPort,omitempty"`
}

// ImportedAccount is account derived from private key
type ImportedAccount struct {
	// PrivateKeySecretName is the secret name holding account private key
	PrivateKeySecretName string `json:"privateKeySecretName"`
	// PasswordSecretName is the secret holding password used to encrypt account private key
	PasswordSecretName string `json:"passwordSecretName"`
}

// NodeStatus defines the observed state of Node
type NodeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Node is the Schema for the nodes API
type Node struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeSpec   `json:"spec,omitempty"`
	Status NodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeList contains a list of Node
type NodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Node `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Node{}, &NodeList{})
}
