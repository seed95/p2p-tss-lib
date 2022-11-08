package p2ptsslib_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	tssp2pLib "github.com/seed95/p2p-tss-lib"
	"github.com/seed95/p2p-tss-lib/config"
	"github.com/seed95/p2p-tss-lib/option"
	"github.com/seed95/p2p-tss-lib/test"

	"github.com/ipfs/go-log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_Reshare_WithPrivateKey(t *testing.T) {
	ctx := context.Background()
	// Must use the following values for NumNodes and Threshold
	NumNodes := test.Participants + 1
	Threshold := test.Threshold

	// Load keys
	keys, err := test.LoadTestKeys(NumNodes - 1)
	require.NoError(t, err)
	require.Equal(t, NumNodes-1, len(keys))

	// New nodes
	fmt.Println("Create nodes")
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		// Options
		opts := []config.Option{
			option.SetThreshold(Threshold),
			//option.SetTimeout(30 * time.Second),
		}

		// Last node has no key
		if i != NumNodes-1 {
			// Unmarshal key
			key, err := json.Marshal(keys[i])
			require.NoError(t, err)

			opts = append(opts, option.SetPrivateKey(key))
		}

		n, err := tssp2pLib.New(ctx, opts...)
		require.NoError(t, err)
		nodes = append(nodes, n)
	}

	// Close nodes
	defer func() {
		fmt.Println("Close nodes")
		for _, n := range nodes {
			n.Close()
		}
	}()

	// Add nodes to node 1
	fmt.Println("Create network")
	for _, n := range nodes[1:] {
		err := nodes[0].AddPeer(ctx, n.HostID(), n.HostAddrs()...)
		require.NoError(t, err)
		time.Sleep(time.Duration(NumNodes) * time.Second)
	}

	err = log.SetLogLevel("tss-lib", "info")
	require.NoError(t, err)

	// Reshare key
	fmt.Println("Reshare key")
	publicKeyReshare, err := nodes[0].Reshare(ctx)
	require.NoError(t, err)

	fmt.Println(publicKeyReshare)

	// Sign message
	fmt.Println("Sign Message")
	rawMessage := []byte("hello world!")
	_, err = nodes[0].SignMessage(ctx, rawMessage)
	require.NoError(t, err)
}

func TestNode_Reshare(t *testing.T) {
	ctx := context.Background()
	NumPeers := 3 // There must be at least 3
	Threshold := 2

	// Options
	var opts []config.Option
	opts = append(opts, option.SetThreshold(Threshold))

	// New peers
	var peers []tssp2pLib.Node
	for i := 0; i < NumPeers; i++ {
		p, err := tssp2pLib.New(ctx, opts...)
		assert.NoError(t, err)
		peers = append(peers, p)
	}

	// Close peers
	defer func() {
		for _, p := range peers {
			p.Close()
		}
	}()

	// Add peers to peer 1 except last peer
	for _, p := range peers[1 : NumPeers-1] {
		err := peers[0].AddPeer(ctx, p.HostID(), p.HostAddrs()...)
		assert.NoError(t, err)
		time.Sleep(time.Duration(NumPeers) * time.Second)
	}

	// Generate key
	publicKey, err := peers[0].KeyGen(ctx)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	// Check public key for all peers is equal
	for _, p := range peers[1 : NumPeers-1] {
		require.Equal(t, publicKey, p.PublicKey())
	}

	// Add last peer to peer 1
	err = peers[0].AddPeer(ctx, peers[NumPeers-1].HostID(), peers[NumPeers-1].HostAddrs()...)
	assert.NoError(t, err)
	time.Sleep(time.Duration(NumPeers) * time.Second)

	publicKeyReshare, err := peers[0].Reshare(ctx)
	require.NoError(t, err)

	require.Equal(t, publicKey, publicKeyReshare)

	// Check public key for all peers is equal
	for _, p := range peers[1:] {
		require.Equal(t, publicKey, p.PublicKey())
	}

}
