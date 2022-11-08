package option_test

import (
	"context"
	"testing"
	"time"

	tssp2p "github.com/seed95/p2p-tss-lib"
	"github.com/seed95/p2p-tss-lib/config"
	"github.com/seed95/p2p-tss-lib/log"
	"github.com/seed95/p2p-tss-lib/option"

	"github.com/stretchr/testify/require"
)

func TestWithHostKey(t *testing.T) {
	ctx := context.Background()

	var opts []config.Option
	n1, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, n1)
	time.Sleep(time.Second * 1)
	n1.Close()

	opts = append(opts, option.SetHostId(n1.HostIdKey()))
	n11, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.Equal(t, n1.HostID(), n11.HostID())
	n11.Close()
}

func TestWithProtocolId_Same(t *testing.T) {
	ctx := context.Background()

	opts := []config.Option{
		option.SetProtocolId("pId1"),
	}
	n1, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, n1)
	time.Sleep(time.Second * 1)

	opts = []config.Option{
		option.SetProtocolId("pId1"),
	}
	n2, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, n2)

	err = n1.AddPeer(ctx, n2.HostID(), n2.HostAddrs()...)
	require.NoError(t, err)

	time.Sleep(time.Second * 1)
	n1.Close()
	time.Sleep(time.Second * 1)
	n2.Close()
}

func TestWithProtocolId_Different(t *testing.T) {
	ctx := context.Background()

	opts := []config.Option{
		option.SetProtocolId("pId1"),
	}
	n1, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, n1)
	time.Sleep(time.Second * 1)

	opts = []config.Option{
		option.SetProtocolId("pId2"),
	}
	n2, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, n2)

	err = n1.AddPeer(ctx, n2.HostID(), n2.HostAddrs()...)
	require.NotNil(t, err)

	time.Sleep(time.Second * 1)
	n1.Close()
	time.Sleep(time.Second * 1)
	n2.Close()
}

func TestWithProtocolId_SameAndDifferent(t *testing.T) {
	ctx := context.Background()

	opts := []config.Option{
		option.SetProtocolId("pId1"),
	}
	n1, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, n1)
	time.Sleep(time.Second * 1)

	// Same protocol id
	n2, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, n2)
	time.Sleep(time.Second * 1)

	err = n1.AddPeer(ctx, n2.HostID(), n2.HostAddrs()...)
	require.NoError(t, err)
	time.Sleep(time.Second * 1)

	// Different protocol id
	opts = []config.Option{
		option.SetProtocolId("pId2"),
	}
	n3, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, n3)

	err = n1.AddPeer(ctx, n3.HostID(), n3.HostAddrs()...)
	require.NotNil(t, err)

	time.Sleep(time.Second * 1)
	n1.Close()
	time.Sleep(time.Second * 1)
	n2.Close()
	time.Sleep(time.Second * 1)
	n3.Close()

}

func TestInfoLog(t *testing.T) {
	ctx := context.Background()

	opts := []config.Option{
		option.SetLogLevel(log.InfoLevel),
	}
	n1, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, n1)
	time.Sleep(time.Second * 1)
	n1.Close()
}

func TestErrorLog(t *testing.T) {
	ctx := context.Background()

	opts := []config.Option{
		option.SetLogLevel(log.ErrorLevel),
	}
	n1, err := tssp2p.New(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, n1)
	time.Sleep(time.Second * 1)
	n1.Close()
}
