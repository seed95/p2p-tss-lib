package p2ptsslib

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/seed95/p2p-tss-lib/config"
	"github.com/seed95/p2p-tss-lib/log"
	"github.com/seed95/p2p-tss-lib/network"
	"github.com/seed95/p2p-tss-lib/option"
	peerPb "github.com/seed95/p2p-tss-lib/peer/pb"
	"github.com/seed95/p2p-tss-lib/storage"
	"github.com/seed95/p2p-tss-lib/tss"
	tssPb "github.com/seed95/p2p-tss-lib/tss/pb"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	p2pNetwork "github.com/libp2p/go-libp2p-core/network"
	p2pPeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/logrusorgru/aurora"
	"go.uber.org/atomic"
)

type Msg1 = peerPb.Message
type TssMsg = tssPb.Message

type (
	Node interface {
		// AddPeer try to connect to new peer and call other peers to connect to new peer
		// Assume that the peers are connected without error and the new peers that connect to the network do not have a shared key.
		AddPeer(ctx context.Context, id string, addrs ...string) error
		KeyGen(ctx context.Context) (publicKey string, err error)
		// SignMessage signs the raw message. For the sign operation,
		// we assume that the node on which the sign request occurs participates in the operation
		// and must have a shared key otherwise return error.
		SignMessage(ctx context.Context, rawMsg []byte) (signature string, err error)
		// SignEthTransaction signs the ethereum transactions. For the sign operation,
		// we assume that the node on which the sign request occurs participates in the operation
		// and must have a shared key otherwise return error.
		// TODO add chain id to this function
		SignEthTransaction(ctx context.Context, tx *types.Transaction) (signature string, err error)
		Reshare(ctx context.Context) (publicKey string, err error)

		// Close node and peers
		// The behavior of Close after the first call is undefined.
		// Specific implementations may document their own behavior.
		Close()

		HostID() string
		// HostAddrs used this method in test functions
		HostAddrs() []string
		// HostIdKey return host private key for generated host(peer) ID
		// inside the p2p library
		HostIdKey() crypto.PrivKey

		// PublicKey return public key. Returns an empty string if KeyGen was never called for this node
		PublicKey() string

		GetStorage() storage.Storage
	}

	node struct {
		// node as a tss signer
		tss.Signer

		// node as a storage
		storage.Storage

		//TODO check interface embedding in struct

		// isClosed set to true when Close the node
		// TODO must be used in each method in node to check node is started and not closed
		isClosed bool
		// TODO check sync.Once in basic_host libp2p(closeSync)
		// ensures we shutdown ONLY once
		//closeSync sync.Once

		// TODO check this variable.
		// closed variable is used to check that the node is already closed
		// When call Close method this variable set to true
		//closed bool

		// aProcessIsRunning used to check that only one process is running
		// Check this variable in all methods(KeyGen, SignMessage, ReShare)
		aProcessIsRunning atomic.Bool

		host host.Host
		// protocolId is the protocol used to connect to peers
		protocolId protocol.ID

		kademliaDHT      *dht.IpfsDHT
		routingDiscovery *discovery.RoutingDiscovery
		bootstrapWg      sync.WaitGroup

		// timeout is used for requests that do not have a timeout for context.
		timeout time.Duration

		// thresholdParty is the number of parties we need for sign, keygen, and reshare operations.
		// Note: This node always participates in sign, keygen, and reshare operations.
		// So, when the thresholdParty value is equal to 2, the number of parties participating in any operation will be (thresholdParty + 1 = 3).
		thresholdParty int

		// peers is map of the peer ID to peer, that does not include this node.
		// Used to communicate with the other Peers.
		// note: ID of peer is equal with its partyId.
		peers map[string]network.Peer

		// signerPeers is map of the peer ID to peer that does not include this node.
		// Used to communicate with the other Peers to sign data and send update sign message.
		// note: ID of peer is equal with its partyId.
		signerPeers map[string]network.Peer

		// receivedMsg used to process messages received from other peers.
		receivedMsg chan Msg1

		// newConnectionCh used to process new connection from streamHandler
		newConnectionCh chan p2pNetwork.Stream

		// peerRErr used to delete peer when happened an error in read data from peer
		peerRErr chan network.Error // peer read error

		// peerWErr is used for errors that occur when sending messages to peers
		peerWErr chan network.Error // peer write error

		// tssErr used to handle tss operations error
		// If any tss operation, including KeyGen, SignMessage, Reshare, fails, the error will drop into this channel.
		//tssErr chan error

		// quit used to stop node in run function.
		// When the node is stops, will be sent to this channel to stop the program in the run function
		quit chan struct{}
	}
)

var _ Node = (*node)(nil)

func New(ctx context.Context, opts ...config.Option) (Node, error) {
	opts = append(opts, option.FallbackDefaults)
	var cfg config.Config
	var err error
	if err = cfg.Apply(opts...); err != nil {
		return nil, err
	}

	// Set log level
	log.SetLevel(cfg.LogLevel)

	n := new(node)
	n.protocolId = protocol.ID(cfg.ProtocolId)
	n.timeout = cfg.ContextTimeout
	n.thresholdParty = cfg.ThresholdParty
	n.Storage = cfg.Storage

	// Make channels
	n.receivedMsg = make(chan Msg1)
	n.newConnectionCh = make(chan p2pNetwork.Stream)
	n.peerRErr = make(chan network.Error)
	n.peerWErr = make(chan network.Error)
	n.quit = make(chan struct{})

	// Make wait groups
	n.bootstrapWg = sync.WaitGroup{}

	// Create p2p host
	n.host, err = libp2p.New(cfg.P2pOptions...)
	if err != nil {
		return nil, err
	}
	n.host.SetStreamHandler(n.protocolId, n.streamHandler)

	// Create tss for node
	n.Signer, err = tss.New(n.HostID(), cfg.ThresholdParty, cfg.TssOptions)
	if err != nil {
		return nil, err
	}

	// Assign peers for node
	n.peers = make(map[string]network.Peer)

	log.Info("NEW P2P", aurora.Green(fmt.Sprint("node created. ID: ", n.host.ID().String())))
	log.Info("NEW P2P", aurora.Cyan(fmt.Sprint("node addresses: ", n.host.Addrs())))

	// Start node
	go n.start(ctx)

	return n, nil
}

func (n *node) Close() {
	n.isClosed = true

	fmt.Println("close")
	// Stop run function
	n.quit <- struct{}{}
	fmt.Println("close after")

	var err error
	for key, p := range n.peers {
		if err = p.Close(); err != nil {
			fmt.Println(aurora.Sprintf(aurora.Red("error in closing peer: %v, error: %v"), p.ID(), err))
		}
		n.host.Peerstore().RemovePeer(p2pPeer.ID(p.ID()))
		delete(n.peers, key)
	}

	if n.host != nil {
		if err = n.host.Close(); err != nil {
			fmt.Println(aurora.Sprintf(aurora.Red("error in closing host: %v, error: %v"), n.HostID(), err))
		}
	}

	if n.kademliaDHT != nil {
		if err := n.kademliaDHT.Close(); err != nil {
			fmt.Println(aurora.Sprintf(aurora.Red("error in closing kademliaDHT, error: %v"), err))
		}
	}

	//TODO check needed?
	//close(n.quit)
	// TODO check others channel must be closed?

	log.Info("CLOSE NODE", aurora.BrightGreen(fmt.Sprintf("node closed. ID: %s", n.HostID())))
}

// This function is called when a peer initiates a connection and
// starts a stream with this peer.
func (n *node) streamHandler(stream p2pNetwork.Stream) {
	n.newConnectionCh <- stream
}

func (n *node) HostID() string {
	return n.host.ID().String()
}

func (n *node) HostAddrs() []string {
	var addrs []string
	for _, addr := range n.host.Addrs() {
		addrs = append(addrs, addr.String())
	}

	return addrs
}

func (n *node) HostIdKey() crypto.PrivKey {
	return n.host.Peerstore().PrivKey(n.host.ID())
}

func (n *node) PublicKey() string {
	return n.GetPublicKey()
}

func (n *node) GetStorage() storage.Storage {
	return n.Storage
}
