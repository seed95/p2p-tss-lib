package p2ptsslib_test

import (
	"context"
	"testing"
	"time"

	tssp2pLib "github.com/seed95/p2p-tss-lib"
	"github.com/seed95/p2p-tss-lib/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_Close(t *testing.T) {
	ctx := context.Background()
	NumNodes := 2

	// New nodes
	var opts []config.Option
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		n, err := tssp2pLib.New(ctx, opts...)
		require.NoError(t, err)
		nodes = append(nodes, n)
	}

	// Close nodes
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	time.Sleep(2 * time.Second)
}

func TestNode_WithoutClose(t *testing.T) {
	ctx := context.Background()
	NumNodes := 2

	// New nodes
	var opts []config.Option
	var nodes []tssp2pLib.Node
	for i := 0; i < NumNodes; i++ {
		n, err := tssp2pLib.New(ctx, opts...)
		assert.NoError(t, err)
		nodes = append(nodes, n)
	}

	time.Sleep(2 * time.Second)
}
