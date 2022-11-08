package option

import (
	"time"

	"github.com/seed95/p2p-tss-lib/config"
	"github.com/seed95/p2p-tss-lib/log"
	"github.com/seed95/p2p-tss-lib/storage"
)

var DefaultProtocolId config.Option = func(cfg *config.Config) error {
	cfg.ProtocolId = "tss-p2p"
	return nil
}

var DefaultThresholdParty config.Option = func(cfg *config.Config) error {
	cfg.ThresholdParty = 2
	return nil
}

var DefaultContextTimeout config.Option = func(cfg *config.Config) error {
	cfg.ContextTimeout = 10 * time.Minute
	return nil
}

var DefaultStorage config.Option = func(cfg *config.Config) error {
	cfg.Storage = storage.NewDefaultStorage()
	return nil
}

var DefaultLogLevel config.Option = func(cfg *config.Config) error {
	cfg.LogLevel = log.NoLog
	return nil
}

// TODO check comment
// Complete list of default options and when to fallback on them.
//
// Please *DON'T* specify default options any other way. Putting this all here
// makes tracking defaults *much* easier.
var defaults = []struct {
	fallback func(cfg *config.Config) bool
	opt      config.Option
}{
	{
		fallback: func(cfg *config.Config) bool { return cfg.ContextTimeout == 0 },
		opt:      DefaultContextTimeout,
	},
	{
		fallback: func(cfg *config.Config) bool { return cfg.ProtocolId == "" },
		opt:      DefaultProtocolId,
	},
	{
		fallback: func(cfg *config.Config) bool { return cfg.ThresholdParty == 0 },
		opt:      DefaultThresholdParty,
	},
	{
		fallback: func(cfg *config.Config) bool { return cfg.Storage == nil },
		opt:      DefaultStorage,
	},
	{
		fallback: func(cfg *config.Config) bool { return cfg.LogLevel == 0 },
		opt:      DefaultLogLevel,
	},
}

// TODO check comment
// FallbackDefaults applies default options to the libp2p node if and only if no
// other relevant options have been applied. will be appended to the options
// passed into New.
var FallbackDefaults config.Option = func(cfg *config.Config) error {
	for _, def := range defaults {
		if !def.fallback(cfg) {
			continue
		}
		if err := cfg.Apply(def.opt); err != nil {
			return err
		}
	}
	return nil
}
