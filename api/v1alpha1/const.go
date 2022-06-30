package v1alpha1

// Node Resources
const (
	DefaultNodeCPURequest = "2"

	DefaultNodeCPULimit = "3"

	DefaultkNodeMemoryRequest = "4Gi"

	DefaultNodeMemoryLimit = "6Gi"

	DefaultNodeStorageRequest = "100Gi"
)

// Genesis
const (
	DefaultNetworkID = 1337
	DefaultChainID   = 1337

	//QBFT
	DefaultQBFTEpochLength           = 30000
	DefaultQBFTBlockPeriodSeconds    = 5
	DefaultQBFTRequestTimeoutSeconds = 10

	//Istanbul
	DefaultIstanbulCeil2Nby3Block = 0
	DefaultIstanbulEpoch          = 30000
	DefaultIstanbulPolicy         = 0
	DefaultIstanbulTestQBFTBlock  = 0
)

// Node Ports
const (
	DefaultP2PPort     = 30303
	DefaultRPCPort     = 8545
	DefaultWSPort      = 8546
	DefaultGraphQLPort = 8547
	DefaultRLPXPort    = 30303
	DefaultMetricsPort = 9545
)
