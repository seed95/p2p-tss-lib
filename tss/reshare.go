package tss

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/seed95/p2p-tss-lib/tss/config"
	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"

	tssBinanceCommon "github.com/binance-chain/tss-lib/common"
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/ecdsa/resharing"
	tssBinance "github.com/binance-chain/tss-lib/tss"
)

type ReshareParams struct {
	Parties   []tsspb.Party
	MessageCh chan<- tsspb.TssMessage
	Done      chan<- struct{}
	ErrCh     chan<- error
}

func (t *tssSigner) ReshareKey(deadline time.Time, p ReshareParams) {
	// Check another process is running
	if t.aProcessIsRunning.Load() {
		p.ErrCh <- ErrAnotherTssProcessIsRunning
		return
	}
	t.aProcessIsRunning.Store(true)
	defer func() {
		t.aProcessIsRunning.Store(false)
	}()

	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Sort old parties (parties that have private key)
	partiesWithPrivateKey := partiesWithKey(p.Parties) // Old parties
	sortedOldParties := sortParties(partiesWithPrivateKey, false)

	// Sort new parties
	sortedNewParties := sortParties(p.Parties, true)

	// Update index for party
	t.updateIndex(sortedNewParties, true)
	if t.HasPrivateKey() {
		t.updateIndex(sortedOldParties, false)
	}

	oldCommitteeCount, newCommitteeCount := len(sortedOldParties), len(sortedNewParties)

	// Create P2P context
	oldP2PCtx := tssBinance.NewPeerContext(sortedOldParties)
	newP2PCtx := tssBinance.NewPeerContext(sortedNewParties)

	// Create channels
	outCh := make(chan tssBinance.Message)
	endCh := make(chan keygen.LocalPartySaveData)
	errCh := make(chan error)

	// If this party has a private key so create reshare party for old parties
	if t.HasPrivateKey() {
		fmt.Println("party has private key", t.partyId.Id)
		params := tssBinance.NewReSharingParameters(tssBinance.S256(), oldP2PCtx, newP2PCtx, t.partyId,
			oldCommitteeCount, t.threshold, newCommitteeCount, t.threshold)
		t.reshareParty = resharing.NewLocalParty(params, t.GetPrivateKey(), outCh, endCh).(*resharing.LocalParty)
	}

	// TODO check needed?
	//wait for some times to first init old parties
	time.Sleep(2 * time.Second)

	// Create reshare party for new parties
	params := tssBinance.NewReSharingParameters(tssBinance.S256(), oldP2PCtx, newP2PCtx, t.newPartyId,
		oldCommitteeCount, t.threshold, newCommitteeCount, t.threshold)
	save := keygen.NewLocalPartySaveData(newCommitteeCount)
	t.reshareNewParty = resharing.NewLocalParty(params, save, outCh, endCh).(*resharing.LocalParty)

	// Start reshare for new parties
	// They will wait for received messages
	go func(P *resharing.LocalParty) {
		if err := P.Start(); err != nil {
			errCh <- err
		}
	}(t.reshareNewParty)

	// If this party has a private key so create reshare party for old parties
	if t.HasPrivateKey() {
		//fmt.Println("party has private key", t.partyId.Id)
		//params := tssBinance.NewReSharingParameters(tssBinance.S256(), oldP2PCtx, newP2PCtx, t.partyId,
		//	oldCommitteeCount, t.threshold, newCommitteeCount, t.threshold)
		//t.reshareParty = resharing.NewLocalParty(params, t.GetPrivateKey(), outCh, endCh).(*resharing.LocalParty)

		// TODO check needed?
		//wait for some times to first start new parties
		time.Sleep(15 * time.Second)
		fmt.Println("create reshare party", t.partyId.Id)

		// Start reshare for old parties.
		// They will send messages
		go func(P *resharing.LocalParty) {
			if err := P.Start(); err != nil {
				errCh <- err
			}
		}(t.reshareParty)
	}

	var reSharingEnded int32
loop:
	for {
		select {

		case <-ctx.Done():
			p.ErrCh <- fmt.Errorf("context done %v", ctx.Err())
			break loop

		case err := <-errCh:
			p.ErrCh <- err
			break loop

		case out := <-outCh:
			if err := t.handleMessageCommittee(out, oldCommitteeCount, p.MessageCh); err != nil {
				p.ErrCh <- err
				break loop
			}

		case key := <-endCh:
			atomic.AddInt32(&reSharingEnded, 1)

			fmt.Println("received key", atomic.LoadInt32(&reSharingEnded))
			// wait till both old and new committee have received the key
			if t.HasPrivateKey() && atomic.LoadInt32(&reSharingEnded) == 1 {
				continue
			}

			t.privateKey = &key
			if err := t.makePublicKey(); err != nil {
				p.ErrCh <- err
				break loop
			}

			// as the generated key is coupled with newPartyId so update partyId with newPartyId
			t.partyId.Key = t.newPartyId.Key
			// regenerate a random key for newPartyId to be used in next reshare procedure
			t.newPartyId.Key = tssBinanceCommon.MustGetRandomInt(config.SizeOfSharedKey).Bytes()

			p.Done <- struct{}{}
			break loop

		}
	}

	// Wait to finish updating reshareParty, reshareNewParty
	// in UpdateReshareParty, UpdateReshareNewParty
	t.wg.Wait()

	// Reshare key is finished and there is no need
	// for reshareParty and reshareNewParty.
	// So, we set their values equal to nil to set them again for use
	//t.reshareParty = nil
	//t.reshareNewParty = nil
}

func (t *tssSigner) UpdateReshareParty(msg *tsspb.TssMessage) error {
	if msg.From.PartyKey == t.partyId.KeyInt() {
		fmt.Println("UpdateReshareParty party key equal")
		return nil
	}

	fmt.Println("UpdateReshareParty tss")

	// Wait to create reshareParty in ReshareKey method
	if err := t.waitToCreateReshareParty(); err != nil {
		return err
	}

	fmt.Println("UpdateReshareParty wait group")

	t.wg.Add(1)
	defer t.wg.Done()

	fromPartyId := tssBinance.NewPartyID(msg.From.ID, "", msg.From.PartyKey)
	fromPartyId.Index = msg.From.Index

	fmt.Println("UpdateReshareParty update from bytes")
	if _, err := t.reshareParty.UpdateFromBytes(msg.Payload, fromPartyId, msg.IsBroadcast); err != nil {
		return fmt.Errorf("error updating from bytes: %v", err)
	}
	fmt.Println("UpdateReshareParty update from bytes finished")
	return nil
}

func (t *tssSigner) UpdateReshareNewParty(msg *tsspb.TssMessage) error {
	if msg.From.NewPartyKey == t.newPartyId.KeyInt() {
		fmt.Println("UpdateReshareParty new party key equal")
		return nil
	}

	fmt.Println("UpdateReshareNewParty tss")
	// Wait to create reshareParty in ReshareKey method
	if err := t.waitToCreateReshareNewParty(); err != nil {
		return err
	}

	fmt.Println("UpdateReshareNewParty wait group")

	t.wg.Add(1)
	defer t.wg.Done()

	fromPartyId := tssBinance.NewPartyID(msg.From.ID, "", msg.From.PartyKey)
	fromPartyId.Index = msg.From.Index

	fmt.Println("UpdateReshareNewParty update from bytes")
	if _, err := t.reshareNewParty.UpdateFromBytes(msg.Payload, fromPartyId, msg.IsBroadcast); err != nil {
		return fmt.Errorf("error updating from bytes: %v", err)
	}
	fmt.Println("UpdateReshareNewParty update from bytes finished")
	return nil
}

func (t *tssSigner) waitToCreateReshareParty() error {
	if t.reshareParty != nil {
		return nil
	}

	duration := time.Second
	timeWaiting := time.Duration(0)
	for {
		if t.reshareParty != nil {
			return nil
		}

		if timeWaiting.Seconds() > WaitingTimeForGeneratingSafePrimes {
			return fmt.Errorf("reshare party is not ready")
		}

		fmt.Println("wait to create reshare party", t.partyId.Id)
		time.Sleep(duration)
		timeWaiting += duration
	}
}

func (t *tssSigner) waitToCreateReshareNewParty() error {
	fmt.Println("waitToCreateReshareNewParty")
	if t.reshareNewParty != nil {
		return nil
	}
	fmt.Println("waitToCreateReshareNewParty before loop")

	duration := time.Second
	timeWaiting := time.Duration(0)
	for {
		fmt.Println("waitToCreateReshareNewParty for loop")
		if t.reshareNewParty != nil {
			return nil
		}

		if timeWaiting.Seconds() > WaitingTimeForGeneratingSafePrimes {
			return fmt.Errorf("reshare new party is not ready")
		}

		fmt.Println("wait to create reshare new party")
		time.Sleep(duration)
		timeWaiting += duration
	}
}

func (t *tssSigner) handleMessageCommittee(message tssBinance.Message, oldCommitteeCount int, msgCh chan<- tsspb.TssMessage) error {
	dest := message.GetTo()
	if dest == nil {
		//TODO define error in tss
		panic(fmt.Errorf("did not expect a message to have a nil destination during resharing"))
	}

	//fmt.Println("from:", message.GetFrom().KeyInt(), "to:", dest)
	//fmt.Println("partyId:", t.partyId.GetId(), "key:", t.partyId.KeyInt())

	// Convert message from tssBinance to tsspb
	msg, err := t.convertMessage(message)
	if err != nil {
		return err
	}

	fmt.Println("handleMessageCommittee type", message.Type())

	//fmt.Println("old", message.IsToOldCommittee(), "old,new", message.IsToOldAndNewCommittees())

	// Message for old committee
	if message.IsToOldCommittee() || message.IsToOldAndNewCommittees() {
		// Update reshare type
		msg.ReshareType = tsspb.OldReshareType

		// Only send to old committee if message is for old and new committee
		msg.To = msg.To[:oldCommitteeCount]

		go func() {
			msgCh <- *msg //TODO check used pointer of channel or not
		}()
	}

	// Message for new committee
	if !message.IsToOldCommittee() || message.IsToOldAndNewCommittees() {
		// Update reshare type
		msg.ReshareType = tsspb.NewReshareType

		go func() {
			msgCh <- *msg //TODO check used pointer of channel or not
		}()
	}

	return nil
}
