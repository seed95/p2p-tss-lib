package option

import (
	"errors"
	"time"

	"github.com/seed95/p2p-tss-lib/config"
	"github.com/seed95/p2p-tss-lib/log"
	"github.com/seed95/p2p-tss-lib/storage"
	tssOption "github.com/seed95/p2p-tss-lib/tss/option"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
)

// SetHostId used to generate host(peer) ID by host(p2p library) private key.
// (Append to P2pOptions)
func SetHostId(key crypto.PrivKey) config.Option {
	return func(cfg *config.Config) error {
		cfg.P2pOptions = append(cfg.P2pOptions, libp2p.Identity(key))
		return nil
	}
}

// SetPrivateKey is used when the private key already has been generated
// and its value has been saved by storage.
// (Append to TssOptions)
func SetPrivateKey(key []byte) config.Option {
	return func(cfg *config.Config) error {
		cfg.TssOptions = append(cfg.TssOptions, tssOption.WithPrivateKey(key))
		return nil
	}
}

// SetProtocolId when used we have a specific communication protocol for peers,
// otherwise the default protocol is used.
func SetProtocolId(id string) config.Option {
	return func(cfg *config.Config) error {
		if id == "" {
			return errors.New("invalid protocol id")
		}
		cfg.ProtocolId = id
		return nil
	}
}

// SetThreshold Specifies the minimum number of parties to perform sign, reshare, keygen operations.
func SetThreshold(n int) config.Option {
	return func(cfg *config.Config) error {
		if n < 1 {
			return errors.New("invalid threshold")
		}
		cfg.ThresholdParty = n
		return nil
	}
}

// SetTimeout set ContextTimeout value for use in requests that do not have a timeout.
func SetTimeout(timeout time.Duration) config.Option {
	return func(cfg *config.Config) error {
		cfg.ContextTimeout = timeout
		return nil
	}
}

// WithStorage used storage.Storage interface
// for store some values in long-term memory like database.
func WithStorage(storage storage.Storage) config.Option {
	return func(cfg *config.Config) error {
		cfg.Storage = storage
		return nil
	}
}

func SetLogLevel(level log.Level) config.Option {
	return func(cfg *config.Config) error {
		cfg.LogLevel = level
		return nil
	}
}
