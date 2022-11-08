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

	"github.com/stretchr/testify/require"
)

func TestNode_SignMessage_WithPrivateKey(t *testing.T) {
	ctx := context.Background()
	// Must use the following values for NumNodes and Threshold
	NumNodes := test.Participants
	Threshold := test.Threshold

	// Load keys
	keys, err := test.LoadTestKeys(NumNodes)
	require.NoError(t, err)
	require.Equal(t, NumNodes, len(keys))

	// New nodes
	fmt.Println("Create nodes")
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		// Unmarshal key
		key, err := json.Marshal(keys[i])
		require.NoError(t, err)

		// Options
		opts := []config.Option{
			option.SetPrivateKey(key),
			option.SetThreshold(Threshold),
			option.SetTimeout(10 * time.Second),
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

	// Sign message
	fmt.Println("Sign Message")
	rawMessage := []byte("hello world!")
	_, err = nodes[0].SignMessage(ctx, rawMessage)
	require.NoError(t, err)
}

func TestNode_SignMessage(t *testing.T) {
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
	fmt.Println("Generate key")
	publicKey, err := nodes[0].KeyGen(ctx)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	// Check public key for all nodes is equal
	for _, p := range nodes[1:] {
		require.Equal(t, publicKey, p.PublicKey())
	}

	// Sign message
	fmt.Println("Sign message")
	rawMessage := []byte("hello world")
	_, err = nodes[0].SignMessage(ctx, rawMessage)
	require.NoError(t, err)
}

func TestNode_SignEthTx(t *testing.T) {
}

// Test sign message with insufficient number of party
func TestNode_SignMessage_InsufficientNumParty(t *testing.T) {
	ctx := context.Background()
	NumNodes := 2
	Threshold := NumNodes - 1 // Exactly one less than NumNodes. Don't change

	// Options
	var opts []config.Option
	opts = append(opts, option.SetThreshold(Threshold))

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
		for _, n := range nodes[:NumNodes-1] {
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

	// Generate key
	fmt.Println("Generate key")
	publicKey, err := nodes[0].KeyGen(ctx)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	// Check public key for all nodes is equal
	for _, p := range nodes[1:] {
		require.Equal(t, publicKey, p.PublicKey())
	}

	// Close last node
	nodes[NumNodes-1].Close()

	// Sign message
	fmt.Println("Sign message")
	rawMessage := []byte("hello")
	_, err = nodes[0].SignMessage(ctx, rawMessage)
	require.NotNil(t, err)
}

// Test sign message without key generation
func TestNode_SignMessage_NotKeyGen(t *testing.T) {
	ctx := context.Background()
	NumNodes := 3 // There must be at least 3

	// Options
	var opts []config.Option
	opts = append(opts, option.SetThreshold(NumNodes))

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
		for _, n := range nodes[:NumNodes-1] {
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

	// Sign message
	fmt.Println("Sign message")
	rawMessage := []byte("hello")
	_, err := nodes[0].SignMessage(ctx, rawMessage)
	require.Equal(t, tssp2pLib.ErrKeyIsNotGenerated, err)
}

// Test sign message without key generation for one node
func TestNode_SignMessage_NodeNoKey(t *testing.T) {
	ctx := context.Background()
	NumNodes := 3             // There must be at least 3
	Threshold := NumNodes - 2 // Exactly two less than NumNodes. Don't change

	// Options
	var opts []config.Option
	opts = append(opts, option.SetThreshold(Threshold))

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
		for _, n := range nodes[:NumNodes-1] {
			n.Close()
		}
	}()

	// Add nodes to node 1 except last node
	fmt.Println("Create network")
	for _, p := range nodes[1 : NumNodes-1] {
		err := nodes[0].AddPeer(ctx, p.HostID(), p.HostAddrs()...)
		require.NoError(t, err)
		time.Sleep(time.Duration(NumNodes) * time.Second)
	}

	// Generate key
	fmt.Println("Generate key")
	publicKey, err := nodes[0].KeyGen(ctx)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	// Check public key for all nodes is equal
	for _, p := range nodes[1 : NumNodes-1] {
		require.Equal(t, publicKey, p.PublicKey())
	}

	// Add last node to node 1
	fmt.Println("Add last node to node 1")
	err = nodes[0].AddPeer(ctx, nodes[NumNodes-1].HostID(), nodes[NumNodes-1].HostAddrs()...)
	require.NoError(t, err)
	time.Sleep(time.Duration(NumNodes) * time.Second)

	// Sign message with last node that have not a private key
	fmt.Println("Sign message with last node that have not a private key")
	rawMessage := []byte("hello")
	_, err = nodes[NumNodes-1].SignMessage(ctx, rawMessage)
	require.Equal(t, tssp2pLib.ErrKeyIsNotGenerated, err)

	// Sign message with first node that have a private key
	fmt.Println("Sign message with first node that have a private key")
	_, err = nodes[0].SignMessage(ctx, rawMessage)
	require.NoError(t, err)
}

// Test sign message in network and send add peer request to peer that not participate in sign operation
func TestNode_SignMessage_AddPeer(t *testing.T) {
}
