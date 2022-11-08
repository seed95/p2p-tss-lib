package tss

import (
	"context"
	"fmt"
	"time"

	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	tssBinance "github.com/binance-chain/tss-lib/tss"
)

type GenerateKeyParams struct {
	Parties   []tsspb.Party
	MessageCh chan<- tsspb.TssMessage
	Done      chan<- struct{}
	ErrCh     chan<- error
}

func (t *tssSigner) GenerateKey(deadline time.Time, p GenerateKeyParams) {
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

	// Sort parties
	sortedParties := sortParties(p.Parties, false)

	// Update index for party
	t.updateIndex(sortedParties, false)

	// Create params
	p2pCtx := tssBinance.NewPeerContext(sortedParties)
	params := tssBinance.NewParameters(tssBinance.S256(), p2pCtx, t.partyId, len(p.Parties), t.threshold)

	outCh := make(chan tssBinance.Message)
	endCh := make(chan keygen.LocalPartySaveData)
	errCh := make(chan error)
	t.keyGenParty = keygen.NewLocalParty(params, outCh, endCh).(*keygen.LocalParty)

	// Start key generation
	go func(P *keygen.LocalParty) {
		if err := P.Start(); err != nil {
			errCh <- err
		}
	}(t.keyGenParty)

loop:
	for {
		select {

		case <-ctx.Done():
			p.ErrCh <- fmt.Errorf("context done %v", ctx.Err())
			return

		case err := <-errCh:
			p.ErrCh <- err
			break loop

		case out := <-outCh:
			msg, err := t.convertMessage(out)
			if err != nil {
				p.ErrCh <- err
				break loop
			}
			p.MessageCh <- *msg //TODO check used pointer of channel or not

		case key := <-endCh:
			t.privateKey = &key
			if err := t.makePublicKey(); err != nil {
				p.ErrCh <- err
				break loop
			}
			p.Done <- struct{}{}
			break loop

		}
	}

	// Wait to finish updating keyGenParty in UpdateKeyGenParty
	t.wg.Wait()
	// Key generation is finished and there is no need for keyGenParty.
	// So, we set its value equal to nil to set it again for use
	t.keyGenParty = nil
}

func (t *tssSigner) UpdateKeyGenParty(msg *tsspb.TssMessage) error {
	// Wait to create keyGenParty in GenerateKey method
	if err := t.waitToCreateKeyGenParty(); err != nil {
		return err
	}

	t.wg.Add(1)
	defer t.wg.Done()

	fromPartyId := tssBinance.NewPartyID(msg.From.ID, "", msg.From.PartyKey)
	fromPartyId.Index = msg.From.Index

	if _, err := t.keyGenParty.UpdateFromBytes(msg.Payload, fromPartyId, msg.IsBroadcast); err != nil {
		return fmt.Errorf("error updating from bytes: %v", err)
	}

	return nil
}

func (t *tssSigner) waitToCreateKeyGenParty() error {
	if t.keyGenParty != nil {
		return nil
	}

	duration := time.Second
	timeWaiting := time.Duration(0)
	for {
		if t.keyGenParty != nil {
			return nil
		}

		if timeWaiting.Seconds() > WaitingTimeForGeneratingSafePrimes {
			return fmt.Errorf("key gen party is not ready")
		}

		time.Sleep(duration)
		timeWaiting += duration
	}
}
