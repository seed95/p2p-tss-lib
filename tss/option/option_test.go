package option_test

import (
	"encoding/json"
	"github.com/seed95/p2p-tss-lib/test"
	"github.com/seed95/p2p-tss-lib/tss"
	"github.com/seed95/p2p-tss-lib/tss/config"
	"github.com/seed95/p2p-tss-lib/tss/option"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestWithPrivateKey(t *testing.T) {
	NumParty := 4 // The maximum can be 20
	Threshold := 1
	keys, err := test.LoadTestKeys(NumParty)
	require.NoError(t, err)
	require.Equal(t, NumParty, len(keys))

	var publicKey string

	for i := 0; i < NumParty; i++ {
		// Unmarshal key
		key, err := json.Marshal(keys[i])
		require.NoError(t, err)

		// Options
		opts := []config.Option{
			option.WithPrivateKey(key),
		}

		Id := "id" + strconv.Itoa(i)
		signer, err := tss.New(Id, Threshold, opts)
		require.NoError(t, err)

		if publicKey == "" {
			publicKey = signer.GetPublicKey()
		}
		require.Equal(t, publicKey, signer.GetPublicKey())
	}

}
