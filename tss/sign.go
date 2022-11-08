package tss

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"

	"github.com/binance-chain/tss-lib/common"
	"github.com/binance-chain/tss-lib/ecdsa/signing"
	tssBinance "github.com/binance-chain/tss-lib/tss"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/logrusorgru/aurora"
)

type SignParams struct {
	Parties   []tsspb.Party
	RawData   []byte
	SignType  string
	MessageCh chan<- tsspb.TssMessage
	Signature chan<- string
	ErrCh     chan<- error
}

func (t *tssSigner) Sign(deadline time.Time, p SignParams) {
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

	if !t.HasPrivateKey() {
		p.ErrCh <- ErrNotHaveSharedKey
		return
	}

	// Sort parties
	sortedParties := sortParties(p.Parties, false)

	// Update index for party
	t.updateIndex(sortedParties, false)

	// Create params
	p2pCtx := tssBinance.NewPeerContext(sortedParties)
	params := tssBinance.NewParameters(tssBinance.S256(), p2pCtx, t.partyId, len(sortedParties), t.threshold)

	// Convert data to *big.Int format
	bigData, err := t.convertRawDataToBigIntFormat(p.RawData, p.SignType)
	if err != nil {
		p.ErrCh <- fmt.Errorf("could not convert to big int format: %v", err)
		return
	}

	outCh := make(chan tssBinance.Message)
	endCh := make(chan common.SignatureData)
	errCh := make(chan error)
	t.signParty = signing.NewLocalParty(bigData, params, t.GetPrivateKey(), outCh, endCh).(*signing.LocalParty)

	// Start sign message
	go func(P *signing.LocalParty) {
		if err := P.Start(); err != nil {
			errCh <- err
		}
	}(t.signParty)

loop:
	for {
		select {

		case <-ctx.Done():
			err = fmt.Errorf("context done %v", ctx.Err())
			p.ErrCh <- err
			return

		case err = <-errCh:
			p.ErrCh <- err
			break loop

		case out := <-outCh:
			msg, err := t.convertMessage(out)
			if err != nil {
				p.ErrCh <- err
				break loop
			}
			p.MessageCh <- *msg //TODO check used pointer of channel or not

		case signedData := <-endCh:
			signature, err := t.convertSignedDataToSignature(&signedData, p.RawData, p.SignType)
			if err != nil {
				p.ErrCh <- err
				break loop
			}
			p.Signature <- signature
			break loop

		}
	}

	// Wait to finish updating signParty in UpdateSignParty
	t.wg.Wait()
	// Sign operation is finished and there is no need for signParty.
	// So, we set its value equal to nil to set it again for use
	t.signParty = nil
}

func (t *tssSigner) UpdateSignParty(msg *tsspb.TssMessage) error {
	// Wait to create signParty in Sign method
	if err := t.waitToCreateSignParty(); err != nil {
		return err
	}

	t.wg.Add(1)
	defer t.wg.Done()

	fromPartyId := tssBinance.NewPartyID(msg.From.ID, "", msg.From.PartyKey)
	fromPartyId.Index = msg.From.Index

	if _, err := t.signParty.UpdateFromBytes(msg.Payload, fromPartyId, msg.IsBroadcast); err != nil {
		return fmt.Errorf("error updating from bytes: %v", err)
	}

	return nil
}

func (t *tssSigner) waitToCreateSignParty() error {
	if t.signParty != nil {
		return nil
	}

	duration := time.Second
	timeWaiting := time.Duration(0)
	for {
		if t.signParty != nil {
			return nil
		}

		if timeWaiting.Seconds() > WaitingTimeForGeneratingSafePrimes {
			return fmt.Errorf("signing party is not ready")
		}

		time.Sleep(duration)
		timeWaiting += duration
	}
}

// convertRawDataToBigIntFormat convert data in *big.Int format
// based on signingType
func (t *tssSigner) convertRawDataToBigIntFormat(rawData []byte, signType string) (*big.Int, error) {
	switch signType {

	case tsspb.SignEthTransactionType:
		return t.getTxBigInt(rawData)

	case tsspb.SignMessageType:
		return t.getMsgBigInt(rawData), nil

	default:
		return nil, fmt.Errorf("unknown sign data type")

	}
}

// getTxBigInt convert raw data to *big.Int format
// Assumes that the raw data is the data of a transaction in the Ethereum network
// and converts it to *big.Int format.
func (t *tssSigner) getTxBigInt(rawData []byte) (*big.Int, error) {
	tx := new(types.Transaction)
	if err := tx.UnmarshalJSON(rawData); err != nil {
		return nil, err
	}

	msg := types.NewEIP155Signer(big.NewInt(t.chainId)).Hash(tx).Bytes()

	return new(big.Int).SetBytes(msg), nil
}

func (t *tssSigner) getMsgBigInt(rawData []byte) *big.Int {
	return new(big.Int).SetBytes(rawData)
}

// convertSignedDataToSignature convert signed data to string (signature) format
// based on signing type
func (t *tssSigner) convertSignedDataToSignature(signedData *common.SignatureData, rawData []byte, signType string) (signature string, err error) {
	switch signType {

	case tsspb.SignEthTransactionType:
		return t.getTxSignature(signedData, rawData)

	case tsspb.SignMessageType:
		return t.getMsgSignature(signedData), nil

	default:
		return "", fmt.Errorf("unknown signing data type: %s", signType)

	}
}

func (t *tssSigner) getTxSignature(signedData *common.SignatureData, rawData []byte) (string, error) {
	tx := new(types.Transaction)
	if err := tx.UnmarshalJSON(rawData); err != nil {
		return "", err
	}

	eip155Signer := types.NewEIP155Signer(big.NewInt(t.chainId))
	data := append(signedData.GetSignature(), signedData.GetSignatureRecovery()...)
	tx, err := tx.WithSignature(eip155Signer, data)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	err = tx.EncodeRLP(buf)
	if err != nil {
		return "", err
	}

	fmt.Println("tx rlp (final signature): ", aurora.BgBlue(hexutil.Encode(buf.Bytes())))

	return hexutil.Encode(buf.Bytes()), nil
}

func (t *tssSigner) getMsgSignature(data *common.SignatureData) string {
	return hexutil.Encode(data.GetSignature())
}
