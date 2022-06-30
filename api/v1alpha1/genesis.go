package v1alpha1

// Genesis is genesis block sepcficition
type Genesis struct {
	// Accounts is array of accounts to fund or associate with code and storage
	Accounts []Account `json:"accounts,omitempty"`

	// NetworkID is network id
	NetworkID uint `json:"networkId"`

	// ChainID is the the chain ID used in transaction signature to prevent reply attack
	// more details https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md
	ChainID uint `json:"chainId"`

	// QBFT QBFT `json:"qbft,omitempty"`

	Istanbul Istanbul `json:"istanbul,omitempty"`
}

type Istanbul struct {
	// +kubebuilder:validation:MinItems=1
	Validators []EthereumAddress `json:"validators,omitempty"`

	Ceil2Nby3Block uint64 `json:"ceil2Nby3Block,omitempty"`
	Epoch          uint64 `json:"epoch,omitempty"`
	Policy         uint64 `json:"policy,omitempty"`
	TestQBFTBlock  uint64 `json:"testQBFTBlock,omitempty"`
}

// https://consensys.net/docs/goquorum//en/latest/configure-and-manage/configure/consensus-protocols/qbft/
type QBFT struct {
	// Validators are initial ibft2 validators
	// +kubebuilder:validation:MinItems=1
	Validators []EthereumAddress `json:"validators,omitempty"`

	EpochLength           uint64 `json:"epochlength"`           // Number of blocks that should pass before pending validator votes are reset
	BlockPeriodSeconds    uint64 `json:"blockperiodseconds"`    // Minimum time between two consecutive IBFT or QBFT blocks’ timestamps in seconds
	RequestTimeoutSeconds uint64 `json:"requesttimeoutseconds"` // Minimum request timeout for each IBFT or QBFT round in milliseconds
}

// Account is Ethereum account
type Account struct {
	// Address is account address
	Address EthereumAddress `json:"address"`

	// Balance is account balance in wei
	Balance HexString `json:"balance,omitempty"`
}

// HexString is String in hexadecial format
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]+$"
type HexString string

// EthereumAddress is ethereum address
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]{40}$"
type EthereumAddress string

// Hash is KECCAK-256 hash
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]{64}$"
type Hash string
