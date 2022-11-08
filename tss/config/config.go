package config

import (
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
)

const (
	SizeOfSharedKey = 256 // Size of shared or party key
)

type Config struct {
	// PrivateKey is set when this party has already participated in the keygen or reshare operation
	// and its private key value has been saved by storage.
	// So, the value of shared (party) key and publicKey is obtained from the value of PrivateKey.
	PrivateKey *keygen.LocalPartySaveData

	// ChainId used to sign Ethereum transactions.
	// transaction signature process uses the chain ID.
	// TODO must be moved chain id to SignParam
	ChainId uint
}

// Option is a tss config option that can be given to the tss constructor
// (`tss.New`).
type Option func(cfg *Config) error

// Apply applies the given options to the config, returning the first error
// encountered (if any).
func (cfg *Config) Apply(opts ...Option) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}
