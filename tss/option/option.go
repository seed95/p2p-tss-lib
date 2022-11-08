package option

import (
	"encoding/json"
	"errors"
	
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/seed95/p2p-tss-lib/tss/config"
)

// WithPrivateKey is used when the private key already has been generated.
func WithPrivateKey(key []byte) config.Option {
	return func(cfg *config.Config) error {
		var privateKey keygen.LocalPartySaveData
		if err := json.Unmarshal(key, &privateKey); err != nil {
			return err
		}
		if !privateKey.ValidateWithProof() {
			return errors.New("invalid private key")
		}
		cfg.PrivateKey = &privateKey
		return nil
	}
}

// TODO must be removed
func WithChainId(chain uint) config.Option {
	return func(cfg *config.Config) error {
		cfg.ChainId = chain
		return nil
	}
}
