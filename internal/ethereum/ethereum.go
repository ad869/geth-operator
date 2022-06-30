package ethereum

import (
	"github.com/ad869/geth-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Ethereum compatible client
type Client interface {
	// Generate Genesis file content
	// https://consensys.net/docs/goquorum//en/latest/configure-and-manage/configure/genesis-file/genesis-options/
	Genesis() (string, error)

	// https://consensys.net/docs/goquorum//en/latest/configure-and-manage/configure/static-nodes/
	EncodeStaticNodes() string

	ConfAndExec
}

// Configure and execute client docker image
type ConfAndExec interface {
	// Refer container args
	Args() []string

	PathData() string

	PathSecrets() string

	PathConfig() string

	// Refer docker image
	Image() string

	ENV() []corev1.EnvVar
}

func NewClient(node v1alpha1.Node) Client {

	return &GoQuorum{node: &node}
}
