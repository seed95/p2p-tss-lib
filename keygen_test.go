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
	"github.com/seed95/p2p-tss-lib/storage"
	"github.com/seed95/p2p-tss-lib/test"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/stretchr/testify/require"
)

func TestDoKeyGenAndSaveKeys(t *testing.T) {
	ctx := context.Background()
	NumNodes := test.Participants

	// New nodes
	fmt.Println("Create nodes")
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		// Options
		opts := []config.Option{
			option.SetThreshold(test.Threshold),
			option.WithStorage(storage.NewDefaultStorage()),
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

	// Generate key
	fmt.Println("Generate key")
	publicKey, err := nodes[0].KeyGen(ctx)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	// Check public key for all nodes is equal
	for _, n := range nodes[1:] {
		require.Equal(t, publicKey, n.PublicKey())
	}

	// Remove old keys
	fmt.Println("Remove old keys")
	test.RemoveKeysAndMakeDir(t)

	// Save keys
	fmt.Println("Save keys")
	for i, n := range nodes {
		// Get shared key
		sharedKey := n.GetStorage().PrivateKey()

		// Unmarshal key
		var key keygen.LocalPartySaveData
		err := json.Unmarshal(sharedKey, &key)
		require.NoError(t, err)

		// Save key to file
		test.SaveTestKey(t, i, key)
	}
}

func TestNode_KeyGen(t *testing.T) {
	ctx := context.Background()
	NumNodes := 2

	// New nodes
	fmt.Println("Create nodes")
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		// Options
		opts := []config.Option{
			option.SetThreshold(NumNodes - 1),
			option.WithStorage(storage.NewDefaultStorage()),
			//option.SetLogLevel(),
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

	// Generate key
	fmt.Println("Generate key")
	publicKey, err := nodes[0].KeyGen(ctx)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	// Check public key for all nodes is equal
	for _, n := range nodes[1:] {
		require.Equal(t, publicKey, n.PublicKey())
	}
}

// Test key generation with insufficient number of party
func TestNode_KeyGen_InsufficientNumParty(t *testing.T) {
	ctx := context.Background()
	NumNodes := 2
	Threshold := NumNodes // There must be equal or bigger than NumNodes

	// Options
	opts := []config.Option{
		option.SetThreshold(Threshold),
	}
	opts = append(opts)

	// New nodes
	fmt.Println("Create nodes")
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
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

	// Generate key
	fmt.Println("Generate key")
	publicKey, err := nodes[0].KeyGen(ctx)
	require.Equal(t, tssp2pLib.ErrNotEnoughPeer, err)
	require.Equal(t, "", publicKey)
	time.Sleep(2 * time.Second)
}

// Send KeyGen request simultaneously in two different nodes
func TestNode_KeyGen_FromTwoNode(t *testing.T) {
	ctx := context.Background()
	NumNodes := 2

	// Options
	opts := []config.Option{
		option.SetThreshold(NumNodes - 1),
	}

	// New nodes
	fmt.Println("Create nodes")
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
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
	for _, p := range nodes[1:] {
		err := nodes[0].AddPeer(ctx, p.HostID(), p.HostAddrs()...)
		require.NoError(t, err)
		time.Sleep(time.Duration(NumNodes) * time.Second)
	}

	pkCh := make(chan string)
	errCh := make(chan error)

	// Generate key in node 1
	fmt.Println("Generate key in node 1")
	go func() {
		publicKey, err := nodes[0].KeyGen(ctx)
		if err != nil {
			errCh <- err
			return
		}
		pkCh <- publicKey
	}()

	time.Sleep(3 * time.Second)

	// Generate key in node 2
	fmt.Println("Generate key in node 2")
	go func() {
		publicKey, err := nodes[1].KeyGen(ctx)
		if err != nil {
			errCh <- err
			return
		}
		pkCh <- publicKey
	}()

loop:
	for {
		select {

		case err := <-errCh:
			require.Equal(t, tssp2pLib.ErrAnotherProcessIsRunning, err)

		case <-pkCh:
			break loop

		}
	}
}

// Test perform KeyGen operations twice
func TestNode_KeyGen_Twice(t *testing.T) {
	ctx := context.Background()
	NumNodes := 2

	// Options
	opts := []config.Option{
		option.SetThreshold(NumNodes - 1),
	}

	// New nodes
	fmt.Println("Create nodes")
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
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

	// Generate key
	fmt.Println("Generate key first time")
	publicKey1, err := nodes[0].KeyGen(ctx)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	// Check public key for all nodes is equal
	for _, n := range nodes[1:] {
		require.Equal(t, publicKey1, n.PublicKey())
	}

	// Generate key for the second time
	fmt.Println("Generate key second time")
	publicKey2, err := nodes[0].KeyGen(ctx)
	require.NoError(t, err)
	require.NotEqual(t, publicKey1, publicKey2)
	time.Sleep(2 * time.Second)

	// Check public key for all nodes is equal
	for _, n := range nodes[1:] {
		require.Equal(t, publicKey2, n.PublicKey())
	}
}

// Test start key generation and add peer call together
func TestNode_KeyGen_AddPeer(t *testing.T) {
	ctx := context.Background()
	NumNodes := 3
	Threshold := NumNodes - 2 // There must be less than (NumNodes-1)

	// Options
	opts := []config.Option{
		option.SetThreshold(Threshold),
		//option.SetLogLevel(log.InfoLevel),
	}

	// New nodes
	fmt.Println("Create nodes")
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
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

	// Add nodes expect last node to node 1
	fmt.Println("Create network")
	for _, n := range nodes[1 : NumNodes-1] {
		err := nodes[0].AddPeer(ctx, n.HostID(), n.HostAddrs()...)
		require.NoError(t, err)
		time.Sleep(time.Duration(NumNodes) * time.Second)
	}

	pkCh := make(chan string)
	errCh := make(chan error)

	// Generate key in node 1
	fmt.Println("Generate key in node 1")
	go func() {
		publicKey, err := nodes[0].KeyGen(ctx)
		if err != nil {
			errCh <- err
			return
		}
		pkCh <- publicKey
	}()

	time.Sleep(2 * time.Second)

	// Add last node to node 1
	fmt.Println("Add last node to node 1")
	go func() {
		err := nodes[0].AddPeer(ctx, nodes[NumNodes-1].HostID(), nodes[NumNodes-1].HostAddrs()...)
		require.Equal(t, tssp2pLib.ErrAnotherProcessIsRunning, err)
	}()

loop:
	for {
		select {

		case err := <-errCh:
			require.NoError(t, err)

		case publicKey := <-pkCh:
			// Check public key for all nodes is equal
			for _, n := range nodes[1 : NumNodes-1] {
				require.Equal(t, publicKey, n.PublicKey())
			}
			break loop

		}
	}

}
