package p2ptsslib

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/seed95/p2p-tss-lib/log"
	"github.com/seed95/p2p-tss-lib/message"
	peerPb "github.com/seed95/p2p-tss-lib/peer/pb"
	"github.com/seed95/p2p-tss-lib/tss"
	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"

	"github.com/logrusorgru/aurora"
)

func (n *node) Reshare(ctx context.Context) (string, error) {
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

	// Send start reshare message to all network
	n.sendReshareMessageToPeers(deadline, n.peerWErr)

	return n.doReshare(deadline)
}

func (n *node) handleReshare(ctx context.Context, msg *Msg1) {
	n.aProcessIsRunning.Store(true)
	defer func() {
		n.aProcessIsRunning.Store(false)
	}()

	// Add timeout to context
	_, cancel := context.WithDeadline(ctx, msg.Deadline)
	defer cancel()

	publicKey, err := n.doReshare(msg.Deadline)
	if err != nil {
		log.Error("HANDLE RESHARE", aurora.Green(aurora.Sprintf("error: %v", err)))
		return
	}

	log.Info("HANDLE RESHARE", aurora.Green(aurora.Sprintf("publicKey: %s", publicKey)))
}

func (n *node) handleUpdateReshareParty(msg *Msg1) error {
	fmt.Println("start handleUpdateReshareParty")
	// Unmarshal payload
	tssMessage := tsspb.TssMessage{}
	if err := tssMessage.Unmarshal(msg.Payload); err != nil {
		return err
	}

	if err := n.UpdateReshareParty(&tssMessage); err != nil {
		return err
	}

	fmt.Println("finish handleUpdateReshareParty")

	return nil
}

func (n *node) handleUpdateReshareNewParty(msg *Msg1) error {
	fmt.Println("start handleUpdateReshareNewParty")

	// Unmarshal payload
	tssMessage := tsspb.TssMessage{}
	if err := tssMessage.Unmarshal(msg.Payload); err != nil {
		return err
	}

	if err := n.UpdateReshareNewParty(&tssMessage); err != nil {
		return err
	}

	fmt.Println("finish handleUpdateReshareNewParty")
	return nil
}

func (n *node) handleReshareUpdateMessage(deadline time.Time, msg tsspb.TssMessage) error {
	// Create peer message payload
	payload, err := msg.Marshal()
	if err != nil {
		return err
	}

	var code uint
	if msg.ReshareType == tsspb.OldReshareType {
		code = message.UpdateReshareOldPartyCode
	} else if msg.ReshareType == tsspb.NewReshareType {
		code = message.UpdateReshareNewPartyCode
	} else {
		panic(fmt.Errorf("invalid reshare type for update reshare party"))
	}

	// Create peer message
	peerMsg := peerPb.Message{
		From:     n.HostID(),
		Code:     code,
		Payload:  payload,
		Deadline: deadline,
	}

	//for _, dest := range msg.To {
	//	if dest == n.HostID() {
	//		fmt.Println("send to my node", dest)
	//		fmt.Println("message", peerMsg.Code, msg.To)
	//		go func() {
	//			n.receivedMsg <- peerMsg
	//		}()
	//		break
	//	}
	//}

	// Broadcast message to network
	if msg.IsBroadcast {
		n.sendMessageToPeers(peerMsg, n.peers, n.peerWErr)
		return nil
	}

	return n.handleUpdateMessage(msg.To, peerMsg, n.peers)
}

func (n *node) doReshare(deadline time.Time) (publicKey string, err error) {
	if !n.haveEnoughPeer() {
		return "", ErrNotEnoughPeer
	}

	// Convert peers to party
	parties := n.getPeersAsParties(false)

	// Append this node as party
	parties = append(parties, n.GetParty())

	// Create reshare params
	messageCh := make(chan tsspb.TssMessage)
	doneCh := make(chan struct{})
	errCh := make(chan error)
	params := tss.ReshareParams{
		Parties:   parties,
		MessageCh: messageCh,
		Done:      doneCh,
		ErrCh:     errCh,
	}

	// Start reshare key
	go n.ReshareKey(deadline, params)

	// Wait to reshare key
	for {
		select {

		case err = <-errCh:
			return "", err

		case msg := <-messageCh:
			log.Info("DO RESHARE RECEIVED", aurora.Yellow(aurora.Sprintf("from: %s, %s", n.HostID(), msg.String())))
			if err = n.handleReshareUpdateMessage(deadline, msg); err != nil {
				return "", err
			}

		case <-doneCh:
			if err = n.doneReshare(deadline); err != nil {
				return "", err
			}
			return n.GetPublicKey(), nil

		}
	}
}

func (n *node) doneReshare(deadline time.Time) error {
	// After doing reshare, all peers have a private key
	for _, p := range n.peers {
		p.SetHasPrivateKey(true)
	}

	// Marshal private key
	key, err := json.Marshal(n.GetPrivateKey())
	if err != nil {
		log.Error("DONE RESHARE", fmt.Sprintf("couldn't marshal shared key. err: %v", err))
	}

	// Store private key in storage
	if err = n.StorePrivateKey(key); err != nil {
		log.Error("DONE RESHARE", fmt.Sprintf("couldn't store private key in storage. err: %v", err))
	}

	// Send my party of node to all network
	return n.sendMyPartyToPeers(deadline, n.peerWErr)
}
