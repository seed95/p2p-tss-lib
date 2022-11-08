package config

import (
	"github.com/libp2p/go-libp2p"
	"github.com/seed95/p2p-tss-lib/log"
	"github.com/seed95/p2p-tss-lib/storage"
	"github.com/seed95/p2p-tss-lib/tss/config"
	"time"
)

type Config struct {
	// P2pOptions is used to create a host from the p2p library.
	P2pOptions []libp2p.Option

	// TODO make comment
	TssOptions []config.Option

	// ProtocolId is an identifier used for the communication protocol with other peers.
	// It must have the same value for all peers who want to connect to each other.
	ProtocolId string

	// ThresholdParty is the number of parties we need for sign, keygen, and reshare operations.
	// Note: The main node always participates in sign, keygen, and reshare operations.
	// So, when the ThresholdParty value is equal to 2, the number of parties participating in any operation will be (ThresholdParty + 1 = 3).
	ThresholdParty int

	// ContextTimeout is used for requests that do not have a timeout.
	ContextTimeout time.Duration

	// TODO make comment
	Storage storage.Storage

	// LogLevel is used for logging in stdout.
	// The default value is no log is shown in stdout.
	LogLevel log.Level
}

// Option is a node config option that can be given to the node constructor
// (`tssp2p.New`).
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
