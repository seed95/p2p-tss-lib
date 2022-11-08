package tss

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/seed95/p2p-tss-lib/tss/config"
	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"

	"github.com/binance-chain/tss-lib/common"
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/ecdsa/resharing"
	"github.com/binance-chain/tss-lib/ecdsa/signing"
	tssBinance "github.com/binance-chain/tss-lib/tss"
	"go.uber.org/atomic"
)

const (
	WaitingTimeForGeneratingSafePrimes = 75 // Second
)

var (
	ErrAnotherTssProcessIsRunning = fmt.Errorf("another distributed tss process is already running")
	ErrPartyNotFound              = fmt.Errorf("party not found in sorted parties")
	ErrNotHaveSharedKey           = fmt.Errorf("party don't have a shared key")
)

type (
	Signer interface {
		GenerateKey(deadline time.Time, p GenerateKeyParams)
		UpdateKeyGenParty(msg *tsspb.TssMessage) error

		Sign(deadline time.Time, p SignParams)
		UpdateSignParty(msg *tsspb.TssMessage) error

		ReshareKey(deadline time.Time, p ReshareParams)
		UpdateReshareParty(msg *tsspb.TssMessage) error
		UpdateReshareNewParty(msg *tsspb.TssMessage) error

		GetPrivateKey() keygen.LocalPartySaveData
		HasPrivateKey() bool
		GetPublicKey() string
		GetParty() tsspb.Party
	}

	tssSigner struct {
		aProcessIsRunning atomic.Bool

		// wgKeyGen used in Update methods(keygen, sign, reshare) to wait to create party.
		// TODO check comment
		wg sync.WaitGroup

		privateKey *keygen.LocalPartySaveData
		publicKey  string
		address    string

		partyId    *tssBinance.PartyID
		newPartyId *tssBinance.PartyID

		// keyGenParty used in key generate operation to start and update.
		// After finished key generate operation, assign to nil in GenerateKey
		keyGenParty *keygen.LocalParty

		// signParty used in sign operation to start and update.
		// After finished sign operation, assign to nil in Sign
		signParty *signing.LocalParty

		// reshareParty used in reshare operation to start and update.
		// After finished reshare operation, assign to nil in ReshareKey.
		reshareParty *resharing.LocalParty

		// reshareNewParty used in reshare operation to start and update.
		// After finished reshare operation, assign to nil in ReshareKey.
		reshareNewParty *resharing.LocalParty

		// threshold specifies the number of parties used for sign, keygen and reshare operations.
		threshold int

		// chainId used to sign Ethereum transactions
		// TODO must be removed
		chainId int64
	}
)

var _ Signer = (*tssSigner)(nil)

func New(id string, threshold int, opts []config.Option) (Signer, error) {
	// TODO set default options
	var cfg config.Config
	if err := cfg.Apply(opts...); err != nil {
		return nil, err
	}

	ts := &tssSigner{
		wg:        sync.WaitGroup{},
		threshold: threshold,
		chainId:   int64(cfg.ChainId), // TODO must be remove
	}

	// Check already generate private key
	var sharedKey *big.Int
	if cfg.PrivateKey != nil {
		ts.privateKey = cfg.PrivateKey
		if err := ts.makePublicKey(); err != nil {
			return nil, errors.New("invalid shared key to make public key")
		}
		sharedKey = cfg.PrivateKey.ShareID
	} else {
		sharedKey = common.MustGetRandomInt(config.SizeOfSharedKey)
	}

	// Set value for PartyId and NewPartyId
	ts.partyId = tssBinance.NewPartyID(id, "", sharedKey)
	ts.newPartyId = tssBinance.NewPartyID(id, "", common.MustGetRandomInt(config.SizeOfSharedKey))

	return ts, nil
}

func (t *tssSigner) GetPrivateKey() keygen.LocalPartySaveData {
	return *t.privateKey
}

func (t *tssSigner) HasPrivateKey() bool {
	return t.privateKey != nil
}

func (t *tssSigner) GetPublicKey() string {
	return t.publicKey
}

func (t *tssSigner) GetParty() tsspb.Party {
	return tsspb.Party{
		ID:            t.partyId.GetId(),
		PartyKey:      t.partyId.KeyInt(),
		NewPartyKey:   t.newPartyId.KeyInt(),
		HaveSharedKey: t.HasPrivateKey(),
		Index:         t.partyId.Index,
	}
}
