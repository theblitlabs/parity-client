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

// Adapter wraps the keystore package to provide a clean interface
type Adapter struct {
	store *keystore.Store
}

// NewAdapter creates a new keystore adapter with default or custom configuration
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

// SaveToken saves an authentication token
func (a *Adapter) SaveToken(token string) error {
	return a.store.SaveToken(token)
}

// LoadToken loads the authentication token
func (a *Adapter) LoadToken() (string, error) {
	return a.store.LoadToken()
}

// SavePrivateKey saves a private key
func (a *Adapter) SavePrivateKey(privateKeyHex string) error {
	return a.store.SavePrivateKey(privateKeyHex)
}

// LoadPrivateKey loads the private key as an ECDSA key
func (a *Adapter) LoadPrivateKey() (*ecdsa.PrivateKey, error) {
	return a.store.LoadPrivateKey()
}

// GetPrivateKeyHex gets the private key in hex format
func (a *Adapter) GetPrivateKeyHex() (string, error) {
	return a.store.GetPrivateKeyHex()
}

// GetStore returns the underlying keystore.Store instance
// This should be used sparingly and only when direct access to the store is necessary
func (a *Adapter) GetStore() *keystore.Store {
	return a.store
}