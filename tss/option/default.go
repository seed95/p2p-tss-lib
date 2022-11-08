package option

import (
	"github.com/seed95/p2p-tss-lib/tss/config"
)

var DefaultChainId config.Option = func(cfg *config.Config) error {
	cfg.ChainId = 1 // Eth main network
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
		fallback: func(cfg *config.Config) bool { return cfg.ChainId == 0 },
		opt:      DefaultChainId,
	},
}

// TODO check comment
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
