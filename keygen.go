package p2ptsslib

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/seed95/p2p-tss-lib/log"
	"github.com/seed95/p2p-tss-lib/message"
	peerPb "github.com/seed95/p2p-tss-lib/peer/pb"
	"github.com/seed95/p2p-tss-lib/tss"
	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"
)

func (n *node) KeyGen(ctx context.Context) (string, error) {
	// Check another process is running
	if n.aProcessIsRunning.Load() {
		return "", ErrAnotherProcessIsRunning
	}
	n.aProcessIsRunning.Store(true)
	defer func() {
		n.aProcessIsRunning.Store(false)
	}()

	// Add timeout to context
	_, cancel, deadline := n.checkContextDeadline(ctx)
	defer cancel()

	// Send start key generation message to all network
	n.sendKeyGenMessageToPeers(deadline, n.peerWErr)

	return n.doKeyGen(deadline)
}

func (n *node) handleKeyGen(ctx context.Context, msg *Msg1) {
	n.aProcessIsRunning.Store(true)
	defer func() {
		n.aProcessIsRunning.Store(false)
	}()

	// Add timeout to context
	_, cancel := context.WithDeadline(ctx, msg.Deadline)
	defer cancel()

	publicKey, err := n.doKeyGen(msg.Deadline)
	if err != nil {
		log.Error("HANDLE KEY GEN", aurora.Green(aurora.Sprintf("error: %v", err)))
		return
	}

	log.Info("HANDLE KEY GEN", aurora.Green(aurora.Sprintf("publicKey: %s", publicKey)))
}

func (n *node) handleUpdateKeyGenParty(msg *Msg1) error {
	// Unmarshal payload
	tssMessage := tsspb.TssMessage{}
	if err := tssMessage.Unmarshal(msg.Payload); err != nil {
		return err
	}

	if err := n.UpdateKeyGenParty(&tssMessage); err != nil {
		return err
	}

	return nil
}

func (n *node) handleKeyGenUpdateMessage(deadline time.Time, msg tsspb.TssMessage) error {
	// Create peer message payload
	payload, err := msg.Marshal()
	if err != nil {
		return err
	}

	// Create peer message
	peerMsg := peerPb.Message{
		From:     n.HostID(),
		Code:     message.UpdateKeyGenCode,
		Payload:  payload,
		Deadline: deadline,
	}

	// Broadcast message to network
	if msg.IsBroadcast {
		n.sendMessageToPeers(peerMsg, n.peers, n.peerWErr)
		return nil
	}

	return n.handleUpdateMessage(msg.To, peerMsg, n.peers)
}

func (n *node) doKeyGen(deadline time.Time) (publicKey string, err error) {
	if !n.haveEnoughPeer() {
		//errKeyGenCh <- ErrNotEnoughPeer
		return "", ErrNotEnoughPeer
	}

	// Convert peers to party
	parties := n.getPeersAsParties(false)

	// Append this node as party
	parties = append(parties, n.GetParty())

	// Create generate key params
	messageCh := make(chan tsspb.TssMessage)
	doneCh := make(chan struct{})
	errCh := make(chan error)
	params := tss.GenerateKeyParams{
		Parties:   parties,
		MessageCh: messageCh,
		Done:      doneCh,
		ErrCh:     errCh,
	}

	// Start generate key
	go n.GenerateKey(deadline, params)

	// Wait to generate key
	for {
		select {

		case err = <-errCh:
			return "", err

		case msg := <-messageCh:
			log.Info("DO KEY GEN RECEIVED", aurora.Yellow(aurora.Sprintf("from: %s, %s", n.HostID(), msg.String())))
			if err = n.handleKeyGenUpdateMessage(deadline, msg); err != nil {
				return "", err
			}

		case <-doneCh:
			n.doneKeyGen()
			return n.GetPublicKey(), nil

		}
	}
}

func (n *node) doneKeyGen() {
	// After doing the key gen, all peers have a private key
	for _, p := range n.peers {
		p.SetHasPrivateKey(true)
	}

	// Marshal private key
	key, err := json.Marshal(n.GetPrivateKey())
	if err != nil {
		log.Error("DONE KEY GEN", fmt.Sprintf("couldn't marshal shared key. err: %v", err))
		return
	}

	// Store private key in storage
	if err = n.StorePrivateKey(key); err != nil {
		log.Error("DONE KEY GEN", fmt.Sprintf("couldn't store private key in storage. err: %v", err))
	}

}
