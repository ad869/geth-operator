package crypto

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

type BIP44 struct {
	mnemonic string
}

func NewBIP44(mnemonic string) *BIP44 {
	return &BIP44{mnemonic: mnemonic}
}

func (b *BIP44) genWallet(accountIndex int) (wallet *hdwallet.Wallet, account accounts.Account, err error) {
	wallet, err = hdwallet.NewFromMnemonic(b.mnemonic)
	if err != nil {
		return
	}

	path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/%d'/0/0", accountIndex))
	account, err = wallet.Derive(path, false)
	if err != nil {
		return
	}
	return
}

func (b *BIP44) DerivePrivateKey(accountIndex int) (privateKey string, err error) {

	wallet, account, err := b.genWallet(accountIndex)
	if err != nil {
		return
	}

	return wallet.PrivateKeyHex(account)
}

func (b *BIP44) DerivePublicKey(accountIndex int) (publicKey string, err error) {

	wallet, account, err := b.genWallet(accountIndex)
	if err != nil {
		return
	}

	return wallet.PublicKeyHex(account)
}

func (b *BIP44) DeriveAddress(accountIndex int) (address string, err error) {

	_, account, err := b.genWallet(accountIndex)
	if err != nil {
		return
	}

	return account.Address.Hex(), nil
}
