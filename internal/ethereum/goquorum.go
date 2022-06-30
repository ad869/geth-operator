package ethereum

import (
	"encoding/json"
	"fmt"

	"github.com/ad869/geth-operator/api/v1alpha1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	corev1 "k8s.io/api/core/v1"
)

const (
	GoQuorumHomeDir      = "/data/quorum"
	GoQuorumDefaultImage = "quorumengineering/quorum:22.4.4"
)

type GoQuorum struct {
	node *v1alpha1.Node
}

// Generate genesis file content.
func (g *GoQuorum) Genesis() (content string, err error) {

	genesis := g.node.Spec.Genesis

	// create some addresses on network, the balance is always in wei.
	alloc := map[string]interface{}{}
	for _, account := range genesis.Accounts {
		alloc[string(account.Address)] = map[string]string{
			"balance": string(account.Balance),
		}
	}

	// extraData is a RLP encoded string containing the validators address.
	extraData, err := createExtraDataFromValidators(genesis.QBFT.Validators)
	if err != nil {
		return
	}

	result := map[string]interface{}{
		"nonce":      "0x0",
		"timestamp":  "0x58ee40ba",
		"gasLimit":   "0xE0000000",
		"gasUsed":    "0x0",
		"number":     "0x0",
		"difficulty": "0x1",
		"coinbase":   "0x0000000000000000000000000000000000000000",
		"mixHash":    "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
		"parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
		"extraData":  extraData,
		"config": map[string]interface{}{
			"isQuorum":            true,
			"chainId":             genesis.ChainID,
			"homesteadBlock":      0,
			"eip150Block":         0,
			"eip150Hash":          "0x0000000000000000000000000000000000000000000000000000000000000000",
			"eip155Block":         0,
			"eip158Block":         0,
			"byzantiumBlock":      0,
			"constantinopleBlock": 0,
			"transitions": []map[string]interface{}{
				{
					"contractsizelimit": 64,
					"block":             0,
				},
			},
			"txnSizeLimit": 64,
			"qbft": map[string]interface{}{
				"blockperiodseconds":    genesis.QBFT.BlockPeriodSeconds,
				"epochlength":           genesis.QBFT.EpochLength,
				"requesttimeoutseconds": genesis.QBFT.RequestTimeoutSeconds,
				"policy":                0,
			},
		},
		"alloc": alloc,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return
	}

	content = string(data)

	return
}

func (g *GoQuorum) EncodeStaticNodes() string {

	if len(g.node.Spec.StaticNodes) == 0 {
		return "[]"
	}

	encoded, _ := json.Marshal(g.node.Spec.StaticNodes)

	return string(encoded)
}

func (g *GoQuorum) Args() (args []string) {

	args = append(args, "--datadir", g.PathData())
	args = append(args, "--nodiscover")
	args = append(args, "--ipcdisable")
	args = append(args, "--verbosity", fmt.Sprintf("%d", g.node.Spec.Verbosity))
	args = append(args, "--syncmode", "full")
	args = append(args, "--metrics")

	args = append(args, "--emitcheckpoints")
	args = append(args, "--networkid", fmt.Sprintf("%d", g.node.Spec.Genesis.NetworkID))
	args = append(args, "--port", fmt.Sprintf("%d", g.node.Spec.P2PPort))

	// rpc
	args = append(args, "--http")
	args = append(args, "--http.addr", "0.0.0.0")
	args = append(args, "--http.port", fmt.Sprintf("%d", g.node.Spec.RPCPort))
	args = append(args, "--http.corsdomain", "*")
	args = append(args, "--http.vhosts", "*")
	args = append(args, "--http.api", "admin,eth,debug,miner,net,txpool,personal,web3,istanbul")

	if g.node.Spec.Miner {
		args = append(args, "--mine")
		args = append(args, "--miner.threads", "1")
		args = append(args, "--miner.gasprice", "0")
		args = append(args, "--unlock", string(g.node.Spec.Coinbase))
		args = append(args, "--allow-insecure-unlock")
		args = append(args, "--password", fmt.Sprintf("%s/account.password", g.PathSecrets()))
	}

	return args
}

func (g GoQuorum) Image() string {
	return GoQuorumDefaultImage
}

// PathData returns blockchain data directory
func (g GoQuorum) PathData() string {
	return fmt.Sprintf("%s/%s", GoQuorumHomeDir, "data")
}

// PathSecrets returns secrets directory
func (g GoQuorum) PathSecrets() string {
	return fmt.Sprintf("%s/%s", GoQuorumHomeDir, ".secret")
}

// PathConfig returns configuration directory
func (g GoQuorum) PathConfig() string {
	return fmt.Sprintf("%s/%s", GoQuorumHomeDir, "config")
}

func (g GoQuorum) ENV() (envs []corev1.EnvVar) {

	envs = []corev1.EnvVar{
		{
			Name:  "DATA_PATH",
			Value: g.PathData(),
		},
		{
			Name:  "CONFIG_PATH",
			Value: g.PathConfig(),
		},
		{
			Name:  "SECRETS_PATH",
			Value: g.PathSecrets(),
		},
		{
			Name:  "PRIVATE_CONFIG",
			Value: "ignore",
		},
	}
	return
}

func createExtraDataFromValidators(validators []v1alpha1.EthereumAddress) (string, error) {

	extraData := "0x"

	var data = &struct {
		Vanity     []byte
		Validators []common.Address
		Vote       *struct {
			RecipientAddress common.Address
			VoteType         byte
		}
		RoundNumber uint32
		Seals       [][]byte
	}{
		Vanity:     make([]byte, 32),
		Validators: make([]common.Address, 0),
	}

	for _, validator := range validators {
		data.Validators = append(data.Validators, common.HexToAddress(string(validator)[2:]))
	}

	payload, err := rlp.EncodeToBytes(data)
	if err != nil {
		return extraData, err
	}

	return extraData + common.Bytes2Hex(payload), nil
}
