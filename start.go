package p2ptsslib

import (
	"context"
	"fmt"

	"github.com/seed95/p2p-tss-lib/message"
	"github.com/seed95/p2p-tss-lib/network"

	"github.com/logrusorgru/aurora"
)

func (n *node) start(ctx context.Context) {
	var (
		msg      Msg1
		err      error
		readErr  network.Error
		writeErr network.Error
	)

running:
	for {
		select {

		case msg = <-n.receivedMsg:
			if err = n.handleMessage(ctx, &msg); err != nil {
				fmt.Println(aurora.Sprintf(aurora.Red("error in handling received message from peer: %v, code: %v, error: %v"),
					msg.From, msg.Code, err))
			}

		case newConn := <-n.newConnectionCh:
			if err = n.handleNewConnection(newConn); err != nil {
				fmt.Println(aurora.Sprintf(aurora.Red("error in handling new connection from peer: %v, error: %v"),
					newConn.Conn().RemotePeer().String(), err))
			}

		case readErr = <-n.peerRErr:
			fmt.Println(aurora.Sprintf(aurora.Red("error read in peer: %v, error: %v"), readErr.ID, readErr))
			if err = n.deletePeer(readErr.ID, false); err != nil {
				fmt.Println(aurora.Sprintf(aurora.Red("error when to delete peer: %v, error: %v"), readErr.ID, readErr))
			}

		case writeErr = <-n.peerWErr:
			fmt.Println(aurora.Sprintf(aurora.Red("error write in peer: %v, error: %v"), writeErr.ID, writeErr))
			if err = n.deletePeer(writeErr.ID, false); err != nil {
				fmt.Println(aurora.Sprintf(aurora.Red("error when to delete peer: %v, error: %v"), writeErr.ID, writeErr))
			}

		case <-n.quit:
			// node was stopped. Run the cleanup logic.
			break running

		}
	}
}

func (n *node) handleMessage(ctx context.Context, msg *Msg1) error {
	if n.processIsRunning(msg.Code) {
		return ErrAnotherProcessIsRunning
	}

	switch msg.Code {

	case message.NewPeerJoinedCode:
		return n.handleNewPeerJoinedMsg(ctx, msg)

	case message.PartyIdCode: // TODO check another process is running or not??
		return n.handlePartyId(msg)

	case message.StartKeyGenCode:
		go n.handleKeyGen(ctx, msg)

	case message.UpdateKeyGenCode:
		return n.handleUpdateKeyGenParty(msg)

	case message.StartSignCode:
		go n.handleSignMessage(ctx, msg)

	case message.UpdateSignCode:
		return n.handleUpdateSignParty(msg)

	case message.StartReshareCode:
		go n.handleReshare(ctx, msg)

	case message.UpdateReshareOldPartyCode:
		go n.handleUpdateReshareParty(msg)

	case message.UpdateReshareNewPartyCode:
		go n.handleUpdateReshareNewParty(msg)

	//case peer_pb.DeletePeerCode:
	//	return n.handleDeletePeerMsg(msg)

	default:
		return fmt.Errorf("unknown message code: %v", msg.Code)

	}

	return nil
}

// processIsRunning check another process is running
// If the code is for updating parties, the check will not be performed
func (n *node) processIsRunning(code uint) bool {
	if code == message.UpdateKeyGenCode ||
		code == message.UpdateSignCode ||
		code == message.UpdateReshareOldPartyCode ||
		code == message.UpdateReshareNewPartyCode {
		return false
	}

	if n.aProcessIsRunning.Load() {
		return true
	}

	return false
}
