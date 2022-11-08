package p2ptsslib

import (
	"context"
	"time"

	"github.com/seed95/p2p-tss-lib/log"
	"github.com/seed95/p2p-tss-lib/message"
	"github.com/seed95/p2p-tss-lib/network"
	peerPb "github.com/seed95/p2p-tss-lib/peer/pb"

	"github.com/logrusorgru/aurora"
)

//func (n *node) sendPartyKeyToPeer(deadline time.Time, peer network.Peer) error {
//	partyId := message.PartyId{
//		PeerId:      n.HostID(),
//		PartyKey:    n.PartyKey(),
//		NewPartyKey: n.NewPartyKey(),
//	}
//
//	payload, err := partyId.Marshal()
//	if err != nil {
//		return err
//	}
//
//	// Create peer message
//	msg := peerPb.Message{
//		From:     n.HostID(),
//		Code:     message.PartyIdCode,
//		Payload:  payload,
//		Deadline: deadline,
//	}
//
//	return peer.SendMsg(msg)
//}

// sendNewPeerMessageToPeers send new peer message to all peers (network)
func (n *node) sendNewPeerMessageToPeers(deadline time.Time, newPeer message.NewPeer) error {
	// Create payload for new peer message
	data, err := newPeer.Marshal()
	if err != nil {
		return err
	}

	// Create new peer joined message
	peerMsg := peerPb.Message{
		From:     n.HostID(),
		Code:     message.NewPeerJoinedCode,
		Payload:  data,
		Deadline: deadline,
	}

	// Send new peer message to all network
	n.sendMessageToPeers(peerMsg, n.peers, n.peerWErr)
	return nil
}

// sendMyPartyToPeer send node as a party to peer
func (n *node) sendMyPartyToPeer(deadline time.Time, peer network.Peer, errCh chan<- network.Error) error {
	party := n.GetParty()

	payload, err := party.Marshal()
	if err != nil {
		return err
	}

	// Create peer message
	msg := peerPb.Message{
		From:     n.HostID(),
		Code:     message.PartyIdCode,
		Payload:  payload,
		Deadline: deadline,
	}

	n.sendMessageToPeer(msg, peer, errCh)
	return nil
}

// sendMyPartyToPeers send node party keys to peers
func (n *node) sendMyPartyToPeers(deadline time.Time, errCh chan<- network.Error) error {
	party := n.GetParty()

	payload, err := party.Marshal()
	if err != nil {
		return err
	}

	// Create peer message
	msg := peerPb.Message{
		From:     n.HostID(),
		Code:     message.PartyIdCode,
		Payload:  payload,
		Deadline: deadline,
	}

	n.sendMessageToPeers(msg, n.peers, errCh)
	return nil
}

// sendKeyGenMessageToPeers send start key gen message to all peers (network)
func (n *node) sendKeyGenMessageToPeers(deadline time.Time, errCh chan<- network.Error) {
	// Create peer message
	peerMsg := peerPb.Message{
		From:     n.HostID(),
		Code:     message.StartKeyGenCode,
		Payload:  nil,
		Deadline: deadline,
	}

	// Send start key generation message to all network
	n.sendMessageToPeers(peerMsg, n.peers, errCh)
}

// sendKeyGenMessageToPeers send start key gen message to all peers (network)
func (n *node) sendSignMessageToPeers(deadline time.Time, sign message.Sign, errCh chan<- network.Error) error {
	// Create payload for sign message
	payload, err := sign.Marshal()
	if err != nil {
		return err
	}

	// Create peer message
	peerMsg := peerPb.Message{
		From:     n.HostID(),
		Code:     message.StartSignCode,
		Payload:  payload,
		Deadline: deadline,
	}

	// Send start sign operation message to all network
	n.sendMessageToPeers(peerMsg, n.signerPeers, errCh)
	return nil
}

// sendReshareMessageToPeers send start reshare message to all peers (network)
func (n *node) sendReshareMessageToPeers(deadline time.Time, errCh chan<- network.Error) {
	// Create peer message
	peerMsg := peerPb.Message{
		From:     n.HostID(),
		Code:     message.StartReshareCode,
		Payload:  nil,
		Deadline: deadline,
	}

	// Send start reshare key message to all network
	n.sendMessageToPeers(peerMsg, n.peers, errCh)
}

// sendMessageToPeers sends the message to peers
func (n *node) sendMessageToPeers(msg Msg1, peers map[string]network.Peer, errCh chan<- network.Error) {
	for _, peer := range peers {
		n.sendMessageToPeer(msg, peer, errCh)
	}
}

// sendMessageToPeer sends the message to peer
func (n *node) sendMessageToPeer(msg Msg1, peer network.Peer, errCh chan<- network.Error) {
	go func() {
		log.Info("SEND",
			aurora.Magenta(aurora.Sprintf("from: %v, to: %v, code: %v, size: %v", n.HostID(), peer.ID(), msg.Code, len(msg.Payload))))
		if err := peer.SendMsg(msg); err != nil {
			errCh <- network.Error{Err: err, ID: peer.ID()}
		}
	}()

}

func (n *node) sendPartyIdToNetwork(ctx context.Context) error {
	////TODO create tss message in another file
	//// Create partyId payload
	//partyId := tss_pb.PartyId{
	//	Id:           n.PartyId(),
	//	NewId:        n.NewPartyId(),
	//	HasPrivateKey: n.HasPrivateKey(),
	//}
	//data, err := partyId.Marshal()
	//if err != nil {
	//	return err
	//}
	//
	//// Create tss message payload
	//tssMessage := tss_pb.Message{
	//	From:        n.PartyId(),
	//	Code:        tss_pb.PartyIdCode,
	//	IsBroadcast: false,
	//	Payload:     data,
	//}
	//payload, err := tssMessage.Marshal()
	//if err != nil {
	//	return err
	//}
	//
	//deadline, _ := ctx.Deadline()
	//
	//msg := peerPb.Message{
	//	From:     n.HostID(),
	//	Code:     peerPb.TssCode,
	//	Payload:  payload,
	//	Deadline: deadline,
	//}
	//
	//// TODO must be handle error when send message to network
	//// TODO not good implementation for expectedId
	//n.sendMessageToNetwork(msg, "")

	return nil
}

// sendMessageToNetwork sends the message to all networks except the `exceptedID`
// TODO must be handle error when send message to network
// TODO not good implementation for expectedId
func (n *node) sendMessageToNetwork(msg Msg1, exceptedID string) {
	for _, peer := range n.peers {
		if peer.ID() == exceptedID {
			continue
		}

		go func(p network.Peer) {
			//TODO handle error
			_ = p.SendMsg(msg)
		}(peer)
	}
}
