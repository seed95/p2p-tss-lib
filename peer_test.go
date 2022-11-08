package p2ptsslib_test

import (
	"context"
	"testing"
	"time"

	tssp2pLib "github.com/seed95/p2p-tss-lib"

	"github.com/stretchr/testify/require"
)

// Add peer to network
func TestNode_AddPeer_P2N(t *testing.T) {
	ctx := context.Background()
	NumNodes := 3 // There must be at least 3

	// New nodes
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		n, err := tssp2pLib.New(ctx, nil)
		require.NoError(t, err)
		nodes = append(nodes, n)
	}

	// Close nodes
	defer func() {
		for _, n := range nodes {
			n.Close()
			time.Sleep(2 * time.Second)
		}
	}()

	// Add nodes to peer 1
	for _, n := range nodes[1:] {
		err := nodes[0].AddPeer(ctx, n.HostID(), n.HostAddrs()...)
		require.NoError(t, err)
		time.Sleep(time.Duration(NumNodes) * time.Second)
	}

	time.Sleep(2 * time.Second)
}

// Add network to peer
func TestNode_AddPeer_N2P(t *testing.T) {
	ctx := context.Background()
	NumNodes := 4 // There must be at least 3

	// New nodes
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		n, err := tssp2pLib.New(ctx, nil)
		require.NoError(t, err)
		nodes = append(nodes, n)
	}

	// Add nodes except last node to node 1
	for _, n := range nodes[1 : NumNodes-1] {
		err := nodes[0].AddPeer(ctx, n.HostID(), n.HostAddrs()...)
		require.NoError(t, err)
		time.Sleep(time.Duration(NumNodes) * time.Second)
	}

	// Add node 1 to last node
	err := nodes[NumNodes-1].AddPeer(ctx, nodes[0].HostID(), nodes[0].HostAddrs()...)
	require.NoError(t, err)
	time.Sleep(time.Duration(NumNodes) * time.Second)

	// Close nodes
	defer func() {
		for _, n := range nodes {
			n.Close()
			time.Sleep(2 * time.Second)
		}
	}()

}

// Add network to network
func TestNode_AddPeer_N2N(t *testing.T) {
	ctx := context.Background()
	NumNodes := 5 // There must be at least 4

	// New nodes
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		n, err := tssp2pLib.New(ctx, nil)
		require.NoError(t, err)
		nodes = append(nodes, n)
	}

	// Close nodes
	defer func() {
		for _, n := range nodes {
			n.Close()
			time.Sleep(2 * time.Second)
		}
	}()

	// Add nodes except 2 last node to node 1
	for _, n := range nodes[1 : NumNodes-2] {
		err := nodes[0].AddPeer(ctx, n.HostID(), n.HostAddrs()...)
		require.NoError(t, err)
		time.Sleep(time.Duration(NumNodes) * time.Second)
	}

	// Connect the last two nodes together to form a network
	err := nodes[NumNodes-1].AddPeer(ctx, nodes[NumNodes-2].HostID(), nodes[NumNodes-2].HostAddrs()...)
	require.NoError(t, err)
	time.Sleep(time.Duration(NumNodes) * time.Second)

	// Connect two network
	err = nodes[NumNodes-1].AddPeer(ctx, nodes[0].HostID(), nodes[0].HostAddrs()...)
	require.NoError(t, err)
	time.Sleep(time.Duration(NumNodes) * time.Second)
}

func TestNode_AddPeer_Duplicate(t *testing.T) {
	ctx := context.Background()
	NumNodes := 2 // There must be at least 3

	// New nodes
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		n, err := tssp2pLib.New(ctx, nil)
		require.NoError(t, err)
		nodes = append(nodes, n)
	}

	// Close nodes
	defer func() {
		for _, n := range nodes {
			n.Close()
			time.Sleep(2 * time.Second)
		}
	}()

	// Add nodes to peer 1
	for _, p := range nodes[1:] {
		err := nodes[0].AddPeer(ctx, p.HostID(), p.HostAddrs()...)
		require.Nil(t, err)
		time.Sleep(time.Duration(NumNodes) * time.Second)
	}

	err := nodes[0].AddPeer(ctx, nodes[1].HostID(), nodes[1].HostAddrs()...)
	require.Equal(t, tssp2pLib.ErrDuplicatePeer, err)

	time.Sleep(2 * time.Second)
}

// Add node to itself
func TestNode_AddPeer_ToItself(t *testing.T) {
	ctx := context.Background()
	NumNodes := 2 // There must be at least 3

	// New nodes
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		n, err := tssp2pLib.New(ctx, nil)
		require.NoError(t, err)
		nodes = append(nodes, n)
	}

	// Close nodes
	defer func() {
		for _, n := range nodes {
			n.Close()
			time.Sleep(2 * time.Second)
		}
	}()

	// Add nodes to node 1
	for _, n := range nodes[1:] {
		err := nodes[0].AddPeer(ctx, n.HostID(), n.HostAddrs()...)
		require.Nil(t, err)
		time.Sleep(time.Duration(NumNodes) * time.Second)
	}

	err := nodes[0].AddPeer(ctx, nodes[0].HostID(), nodes[0].HostAddrs()...)
	require.Equal(t, tssp2pLib.ErrInvalidPeer, err)

	time.Sleep(2 * time.Second)
}
