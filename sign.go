package p2ptsslib

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/seed95/p2p-tss-lib/log"
	"github.com/seed95/p2p-tss-lib/message"
	"github.com/seed95/p2p-tss-lib/network"
	peerPb "github.com/seed95/p2p-tss-lib/peer/pb"
	"github.com/seed95/p2p-tss-lib/tss"
	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/logrusorgru/aurora"
)

func (n *node) SignMessage(ctx context.Context, rawMsg []byte) (signature string, err error) {
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

	// Node has private key
	if !n.HasPrivateKey() {
		return "", ErrKeyIsNotGenerated
	}

	// Select randomly peers that have shared key for signing
	parties, err := n.selectRandomlySignerPeer()
	if err != nil {
		return "", err
	}

	// Append this node as party
	parties = append(parties, n.GetParty())

	// Update signer peers
	n.updateSignerPeers(parties)

	// Create start sign payload
	signMessage := message.Sign{
		Parties: parties,
		Data:    rawMsg,
		Type:    tsspb.SignMessageType,
	}

	// Send start sign message to all signer peers in network
	if err = n.sendSignMessageToPeers(deadline, signMessage, n.peerWErr); err != nil {
		return "", err
	}

	return n.doSign(deadline, signMessage)
}

func (n *node) SignEthTransaction(ctx context.Context, tx *types.Transaction) (signature string, err error) {
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

	// Node has private key
	if !n.HasPrivateKey() {
		return "", ErrKeyIsNotGenerated
	}

	// Select randomly peers that have shared key for signing
	parties, err := n.selectRandomlySignerPeer()
	if err != nil {
		return "", err
	}

	// Append this node as party
	parties = append(parties, n.GetParty())

	// Update signer peers
	n.updateSignerPeers(parties)

	// Marshal transaction
	rawTx, err := tx.MarshalJSON()
	if err != nil {
		return "", err
	}

	// Create start sign payload
	signMessage := message.Sign{
		Parties: parties,
		Data:    rawTx,
		Type:    tsspb.SignEthTransactionType,
	}

	// Send start sign message to all signer peers in network
	errSendMessageCh := make(chan network.Error)
	if err = n.sendSignMessageToPeers(deadline, signMessage, errSendMessageCh); err != nil {
		return "", err
	}

	signatureCh := make(chan string)
	errSignCh := make(chan error)
	// TODO uncomment
	//go n.doSign(deadline, signMessage, signatureCh, errSignCh)

	// Wait to done sign message
	select {

	case signature = <-signatureCh:
		return signature, nil

	case err = <-errSignCh:
		return "", err

	case err = <-errSendMessageCh:
		return "", err

	}
}

func (n *node) handleSignMessage(ctx context.Context, msg *Msg1) {
	n.aProcessIsRunning.Store(true)
	defer func() {
		n.aProcessIsRunning.Store(false)
	}()

	// Add timeout to context
	_, cancel := context.WithDeadline(ctx, msg.Deadline)
	defer cancel()

	// Unmarshal msg to sign
	signMessage := message.Sign{}
	if err := signMessage.Unmarshal(msg.Payload); err != nil {
		fmt.Println("unmarshal sign message err", err)
		return
	}

	// Update signer peers
	n.updateSignerPeers(signMessage.Parties)

	//signatureCh := make(chan string)
	//errSignCh := make(chan error)
	signature, err := n.doSign(msg.Deadline, signMessage)
	if err != nil {
		log.Error("HANDLE SIGN MESSAGE", aurora.Green(aurora.Sprintf("error: %v", err)))
		return
	}

	log.Info("HANDLE SIGN MESSAGE", aurora.Green(aurora.Sprintf("signature: %s", signature)))
}

func (n *node) handleUpdateSignParty(msg *Msg1) error {
	// Unmarshal payload
	tssMessage := tsspb.TssMessage{}
	if err := tssMessage.Unmarshal(msg.Payload); err != nil {
		return err
	}

	if err := n.UpdateSignParty(&tssMessage); err != nil {
		return err
	}

	return nil
}

func (n *node) handleSignUpdateMessage(deadline time.Time, msg tsspb.TssMessage) error {
	// Create peer message payload
	payload, err := msg.Marshal()
	if err != nil {
		return err
	}

	// Create peer message
	peerMsg := peerPb.Message{
		From:     n.HostID(),
		Code:     message.UpdateSignCode,
		Payload:  payload,
		Deadline: deadline,
	}

	// Broadcast message to network
	if msg.IsBroadcast {
		n.sendMessageToPeers(peerMsg, n.signerPeers, n.peerWErr)
		return nil
	}

	return n.handleUpdateMessage(msg.To, peerMsg, n.signerPeers)
}

func (n *node) doSign(deadline time.Time, sign message.Sign) (signature string, err error) {
	// Create sign params
	messageCh := make(chan tsspb.TssMessage)
	sigCh := make(chan string) // used to return from for loop
	errCh := make(chan error)
	params := tss.SignParams{
		Parties:   sign.Parties,
		RawData:   sign.Data,
		SignType:  sign.Type,
		MessageCh: messageCh,
		Signature: sigCh,
		ErrCh:     errCh,
	}

	// Start sign data
	go n.Sign(deadline, params)

	// Wait to sign data
	for {
		select {

		case err = <-errCh:
			return "", err

		case msg := <-messageCh:
			log.Info("DO SIGN RECEIVED", aurora.Yellow(aurora.Sprintf("from: %s, %s", n.HostID(), msg.String())))
			if err = n.handleSignUpdateMessage(deadline, msg); err != nil {
				return "", err
			}

		case sig := <-sigCh:
			return sig, nil

		}
	}
}

// selectRandomlySignerPeer select randomly peers that have shared key for signing.
// If the number of peers that have a shared key is less than the threshold value,
// the error returns.
func (n *node) selectRandomlySignerPeer() ([]tsspb.Party, error) {
	// Return peers that have a shared key as a party
	partiesHaveSharedKey := n.getPeersAsParties(true)

	// Check have enough peers that have a shared key
	if len(partiesHaveSharedKey) < n.thresholdParty {
		return nil, fmt.Errorf("not enough available signer peers to sign. expected: %d, got: %d", n.thresholdParty, len(partiesHaveSharedKey))
	}

	// Generate random permutation to pick (threshold-1) signer peers
	rand.Seed(time.Now().UnixNano())
	perm := rand.Perm(len(partiesHaveSharedKey)) // random permutation of 0..n-1. e.g.(n=4) [2, 0, 3, 1]
	perm = perm[:n.thresholdParty]               // pick first threshold elements. e.g.(thresholdParty=3) [2, 0, 3]

	var parties []tsspb.Party
	for _, v := range perm {
		parties = append(parties, partiesHaveSharedKey[v])
	}

	return parties, nil
}

func (n *node) updateSignerPeers(parties []tsspb.Party) {
	n.signerPeers = make(map[string]network.Peer) // clear old signerPeers to assign new ones
	for _, party := range parties {
		if party.ID == n.HostID() {
			continue
		}
		if peer, ok := n.peers[party.ID]; ok {
			n.signerPeers[party.ID] = peer
		} else {
			// TODO check this condition must be panic
			fmt.Println("not found peer for signing!!!!!")
		}
	}
}
