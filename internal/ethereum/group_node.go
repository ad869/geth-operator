package ethereum

import (
	"fmt"

	"github.com/ad869/geth-operator/api/v1alpha1"
	"github.com/ad869/geth-operator/internal/crypto"
)

type NodeRoleType int

const (
	NodeRoleValidator NodeRoleType = 1 << (iota * 10)
	NodeRoleMember
	NodeRoleBootNode
)

type GroupNode interface {
	UniqueIndex() int
	String() string
	Address() (v1alpha1.EthereumAddress, error)
	EnodeURL() (v1alpha1.Enode, error)
	IsValidator() bool
	PrivateKey() (string, error)
}

type groupNode struct {

	// unique index of the same node role in the cluster
	index int

	nodeType NodeRoleType

	bip44 *crypto.BIP44

	cluster *v1alpha1.Cluster
}

func NewGroupNode(cluster *v1alpha1.Cluster, nodeType NodeRoleType, index int) GroupNode {

	return &groupNode{
		index:    index,
		nodeType: nodeType,
		bip44:    crypto.NewBIP44(cluster.Spec.Mnemonic),
		cluster:  cluster,
	}
}

func (n *groupNode) IsValidator() bool {
	return n.nodeType == NodeRoleValidator
}

func (n *groupNode) EnodeURL() (enodeURL v1alpha1.Enode, err error) {

	var publicKey string
	if publicKey, err = n.bip44.DerivePublicKey(n.UniqueIndex()); err != nil {
		return
	}

	enodeURL = v1alpha1.Enode(fmt.Sprintf("enode://%s@%s.%s.svc.cluster.local:%d?discport=0",
		publicKey,
		n.String(),
		n.cluster.Namespace,
		n.cluster.Spec.P2PPort,
	))

	return
}

func (n *groupNode) PrivateKey() (string, error) {
	return n.bip44.DerivePrivateKey(n.UniqueIndex())
}

func (n *groupNode) Address() (address v1alpha1.EthereumAddress, err error) {

	add, err := n.bip44.DeriveAddress(n.UniqueIndex())
	if err != nil {
		return
	}
	address = v1alpha1.EthereumAddress(add)

	return
}

func (n *groupNode) UniqueIndex() int {
	return int(n.nodeType) + n.index
}

func (n *groupNode) String() string {

	var identifier string

	switch n.nodeType {
	case NodeRoleValidator:
		identifier = "validator"
	case NodeRoleMember:
		identifier = "member"
	default:
		identifier = "node"
	}

	return fmt.Sprintf("%s-%s-%d", n.cluster.Name, identifier, n.index)
}
