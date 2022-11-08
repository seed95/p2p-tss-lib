package p2ptsslib

import (
	"context"
	"fmt"
	"time"

	"github.com/seed95/p2p-tss-lib/log"
	"github.com/seed95/p2p-tss-lib/message"
	"github.com/seed95/p2p-tss-lib/network"
	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"

	p2pNetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/logrusorgru/aurora"
)

func (n *node) AddPeer(ctx context.Context, id string, addrs ...string) error {
	// Check another process is running
	if n.aProcessIsRunning.Load() {
		return ErrAnotherProcessIsRunning
	}
	n.aProcessIsRunning.Store(true)
	defer func() {
		n.aProcessIsRunning.Store(false)
	}()

	// Check addrs is not empty
	if len(addrs) == 0 {
		return fmt.Errorf("new peer address is empty")
	}

	// Add timeout to context
	ctx, cancel, deadline := n.checkContextDeadline(ctx)
	defer cancel()

	// Create new peer
	newPeer := message.NewPeer{ID: id, Address: addrs}

	// Check new peer is valid
	if err := n.isValidNewPeer(newPeer); err != nil {
		return err
	}

	// Send new peer message to all network
	if err := n.sendNewPeerMessageToPeers(deadline, newPeer); err != nil {
		return err
	}

	return n.addPeerToNetwork(ctx, deadline, newPeer)
}

func (n *node) handleNewPeerJoinedMsg(ctx context.Context, msg *Msg1) error {
	n.aProcessIsRunning.Store(true)
	defer func() {
		n.aProcessIsRunning.Store(false)
	}()

	// Add timeout to context
	ctx, cancel := context.WithDeadline(ctx, msg.Deadline)
	defer cancel()

	// Unmarshal payload
	newPeer := message.NewPeer{}
	if err := newPeer.Unmarshal(msg.Payload); err != nil {
		return err
	}

	// Check new peer ID is equal to this node
	if newPeer.ID == n.HostID() {
		return nil
	}

	// Check new peer already exists.
	if _, ok := n.peers[newPeer.ID]; ok {
		return nil
	}

	return n.addPeerToNetwork(ctx, msg.Deadline, newPeer)
}

// handleNewConnection handle new stream connection when a peer
// initiates a connection and starts a stream with this peer.
func (n *node) handleNewConnection(stream p2pNetwork.Stream) error {
	// Check another process is running
	if n.aProcessIsRunning.Load() {
		return ErrAnotherProcessIsRunning
	}
	n.aProcessIsRunning.Store(true)
	defer func() {
		n.aProcessIsRunning.Store(false)
	}()

	peer, err := network.NewPeer(stream)
	if err != nil {
		return err
	}

	// Add timeout to context
	_, cancel, deadline := n.checkContextDeadline(context.Background())
	defer cancel()

	// Create new peer
	newPeer := message.NewPeer{
		ID:      stream.Conn().RemotePeer().String(),
		Address: []string{stream.Conn().RemoteMultiaddr().String()},
	}

	// Send new peer message to all network
	if err = n.sendNewPeerMessageToPeers(deadline, newPeer); err != nil {
		return err
	}

	return n.launchNewPeer(deadline, peer)
}

func (n *node) handlePartyId(msg *Msg1) error {
	p, ok := n.peers[msg.From]
	if !ok {
		return fmt.Errorf("cannot find peer: %v", msg.From)
	}

	// Unmarshal payload
	party := tsspb.Party{}
	if err := party.Unmarshal(msg.Payload); err != nil {
		return err
	}

	p.UpdateParty(party)
	return nil
}

func (n *node) addPeerToNetwork(ctx context.Context, deadline time.Time, np message.NewPeer) error {
	peer, err := n.connectToNewPeer(ctx, np)
	if err != nil {
		return err
	}

	if err = n.launchNewPeer(deadline, peer); err != nil {
		return err
	}

	return nil
}

// connectToNewPeer connect to peer and return network.Peer
func (n *node) connectToNewPeer(ctx context.Context, np message.NewPeer) (network.Peer, error) {
	log.Info("CONNECT TO NEW PEER", aurora.Yellow(fmt.Sprint("try to connect to ", np.String())))

	addrsInfo, err := np.AddrInfo()
	if err != nil {
		return nil, err
	}

	// Connect to new peer
	if err := n.host.Connect(ctx, addrsInfo); err != nil {
		return nil, fmt.Errorf("error connecting to peer: %v, error: %v", np.ID, err)
	}

	// Create stream for new peer
	stream, err := n.host.NewStream(ctx, addrsInfo.ID, n.protocolId)
	if err != nil {
		return nil, fmt.Errorf("error creating stream: %v, error: %v", np.ID, err)
	}

	return network.NewPeer(stream)
}

// launchNewPeer Read from new peer in goroutine and add to peers of node and send partyId of node to new peer
func (n *node) launchNewPeer(deadline time.Time, peer network.Peer) error {
	go n.readFromPeer(peer)

	// Add to peers
	n.peers[peer.ID()] = peer

	if err := n.sendMyPartyToPeer(deadline, peer, n.peerWErr); err != nil {
		return err
	}

	log.Info("LAUNCH NEW PEER", aurora.Green(fmt.Sprintf("connected to %v. peer count: %v", peer.ID(), len(n.peers))))
	return nil
}

// readFromPeer read from peer forever
func (n *node) readFromPeer(p network.Peer) {
	for {
		msg, err := p.ReadData()
		if err != nil {
			n.peerRErr <- network.Error{ID: p.ID(), Err: err}
			return
		}
		log.Info("RECEIVED", aurora.Blue(aurora.Sprintf("from: %v, to: %v, code: %v, size: %v", p.ID(), n.HostID(), msg.Code, len(msg.Payload))))
		n.receivedMsg <- *msg // TODO check pointer or not
	}
}

func (n *node) deletePeer(ID string, sendToNetwork bool) error {
	peer, ok := n.peers[ID]
	if !ok {
		return fmt.Errorf("cannot find peer to delete: %v", ID)
	}

	// Close peer
	if err := peer.Close(); err != nil {
		fmt.Println(aurora.Sprintf(aurora.Red("error in closing peer: %v, error: %v"), ID, err))
	}

	// delete peer
	delete(n.peers, ID)
	if sendToNetwork {
		// TODO must be handle error when send delete peer
		return n.deletePeerFromNetwork(ID)
	}

	return nil
}

func (n *node) deletePeerFromNetwork(ID string) error {
	//// Create payload
	//data := peer_pb.DeletePeer{ID: ID}
	//payload, err := data.Marshal()
	//if err != nil {
	//	return err
	//}
	//
	//// Create delete peer message
	//msg := peer_pb.Message{
	//	From:     n.HostID(),
	//	Code:     peer_pb.DeletePeerCode,
	//	Payload:  payload,
	//	Deadline: time.Time{}, // TODO check deadline not used in new peer joined message
	//}
	//
	//n.sendMessageToNetwork(msg, ID)

	return nil
}

// isValidNewPeer check:
// 1. peer ID not equal to this node
// 2. Don't exist peer ID already
func (n *node) isValidNewPeer(peer message.NewPeer) error {
	// Check new peer ID is equal to this node
	if peer.ID == n.HostID() {
		return ErrInvalidPeer
	}

	// Check new peer already exists.
	if _, ok := n.peers[peer.ID]; ok {
		return ErrDuplicatePeer
	}

	return nil
}
