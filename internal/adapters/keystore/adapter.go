package keystore

import (
	"crypto/ecdsa"
	"os"
	"path/filepath"

	"github.com/theblitlabs/keystore"
)

const (
	DefaultDirName  = ".parity"
	DefaultFileName = "keystore.json"
)

type Adapter struct {
	store *keystore.Store
}

func NewAdapter(cfg *keystore.Config) (*Adapter, error) {
	if cfg == nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		cfg = &keystore.Config{
			DirPath:  filepath.Join(homeDir, DefaultDirName),
			FileName: DefaultFileName,
		}
	}

	store, err := keystore.NewKeystore(*cfg)
	if err != nil {
		return nil, err
	}

	return &Adapter{
		store: store,
	}, nil
}

func (a *Adapter) SaveToken(token string) error {
	return a.store.SaveToken(token)
}

func (a *Adapter) LoadToken() (string, error) {
	return a.store.LoadToken()
}

func (a *Adapter) SavePrivateKey(privateKeyHex string) error {
	return a.store.SavePrivateKey(privateKeyHex)
}

func (a *Adapter) LoadPrivateKey() (*ecdsa.PrivateKey, error) {
	return a.store.LoadPrivateKey()
}

func (a *Adapter) GetPrivateKeyHex() (string, error) {
	return a.store.GetPrivateKeyHex()
}

func (a *Adapter) GetStore() *keystore.Store {
	return a.store
}
