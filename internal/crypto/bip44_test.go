package crypto

import (
	"crypto/ecdsa"
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestDerivePrivateKey(t *testing.T) {
	mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	privateKey, _ := NewBIP44(mnemonic).DerivePrivateKey(1)
	t.Log(privateKey)

	prv1, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return
	}

	// public key
	publicKey := prv1.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		err = errors.New("publicKey is not of type *ecdsa.PublicKey")
		return
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKeyHex := hexutil.Encode(publicKeyBytes)[4:]
	t.Log(publicKeyHex)
	//7b22ab6b06ca8790b08829e4a2525367ad9916fa77c215933a126a2ac55b76ff15ebb765e1d0421c02c383cd2e4525a914437c166b465b28390a7fd6345650a0
	//7b22ab6b06ca8790b08829e4a2525367ad9916fa77c215933a126a2ac55b76ff15ebb765e1d0421c02c383cd2e4525a914437c166b465b28390a7fd6345650a0

}

func TestDerivePublicKey(t *testing.T) {
	mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	pubKey, _ := NewBIP44(mnemonic).DerivePublicKey(1)
	t.Log(pubKey)
}

func TestDeriveAddress(t *testing.T) {
	mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	address, _ := NewBIP44(mnemonic).DeriveAddress(1)
	t.Log(address)
}

func TestDecript(t *testing.T) {
	// mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	// prv1, _ := NewBIP44(mnemonic).DerivePrivateKey(1)
	// prv2, _ := NewBIP44(mnemonic).DerivePrivateKey(2)

	// pubKey1, _ := NewBIP44(mnemonic).DerivePublicKey(1)
	// pubKey2, _ := NewBIP44(mnemonic).DerivePublicKey(1)

	// message := []byte("Hello, world.")
	// ct, err := crypto.Encrypt(rand.Reader, &prv2.PublicKey, message, nil, nil)
	// if err != nil {
	// 	t.Fatal(err)
	// }

}
