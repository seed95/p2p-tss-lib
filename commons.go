package p2ptsslib

import (
	"context"
	"fmt"
	"time"

	"github.com/seed95/p2p-tss-lib/network"
	peerPb "github.com/seed95/p2p-tss-lib/peer/pb"
	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"
)

// checkContextDeadline check ctx has deadline or not.
// If ctx does not include the deadline, we add the timeout value to it.
func (n *node) checkContextDeadline(ctx context.Context) (context.Context, context.CancelFunc, time.Time) {
	d, ok := ctx.Deadline()
	if !ok {
		d = time.Now().Add(n.timeout)
	}
	ctx, cancel := context.WithDeadline(ctx, d)
	return ctx, cancel, d
}

// haveEnoughPeer check this node have enough peers
func (n *node) haveEnoughPeer() bool {
	return len(n.peers) >= n.thresholdParty
}

// partiesWithKey return peer IDs in peers that have shared key
func (n *node) partiesWithKey() []string {
	// Find peer IDs that have shared key
	var haveSharedKeyIDs []string
	for _, p := range n.peers {
		if p.HasPrivateKey() {
			haveSharedKeyIDs = append(haveSharedKeyIDs, p.ID())
		}
	}
	return haveSharedKeyIDs
}

// getPeersAsParties returns the party for peers
// If checkKey is true, returns party for peers that those have shared key.
func (n *node) getPeersAsParties(checkKey bool) []tsspb.Party {
	var parties []tsspb.Party

	// Append parties for peers
	for _, p := range n.peers {
		if checkKey {
			if p.HasPrivateKey() {
				parties = append(parties, p.GetParty())
			}
		} else {
			parties = append(parties, p.GetParty())
		}
	}

	return parties
}

// handleUpdateMessage sent update tss message to destinationsIDs for update their parties
func (n *node) handleUpdateMessage(destinationIDs []string, msg peerPb.Message, peers map[string]network.Peer) error {
	// Check destination is valid
	if destinationIDs == nil {
		return fmt.Errorf("message from party %v does not include the destination", n.HostID())
	}

	fmt.Println("destination size:", len(destinationIDs))

	destPeers := make(map[string]network.Peer)
	var peer network.Peer
	var ok bool
	for _, dest := range destinationIDs {
		// Check destination and source ID is equal
		if dest == n.HostID() {
			go func() {
				n.receivedMsg <- msg
			}()
			continue
		}

		// Check, destination party exist
		peer, ok = peers[dest]
		if !ok {
			return fmt.Errorf("party %v tried to send message to unknown party %v", n.HostID(), dest)
		}
		destPeers[peer.ID()] = peer
	}

	n.sendMessageToPeers(msg, destPeers, n.peerWErr)
	return nil
}
