package tss

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/seed95/p2p-tss-lib/log"
	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"

	tssBinance "github.com/binance-chain/tss-lib/tss"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/logrusorgru/aurora"
)

func (t *tssSigner) convertMessage(libMessage tssBinance.Message) (*tsspb.TssMessage, error) {
	// Create payload
	payload, _, err := libMessage.WireBytes()
	if err != nil {
		return nil, err
	}

	var destinations []string
	for _, dest := range libMessage.GetTo() {
		destinations = append(destinations, dest.Id)
	}

	from := tsspb.Party{
		ID:       libMessage.GetFrom().GetId(),
		PartyKey: libMessage.GetFrom().KeyInt(),
		Index:    libMessage.GetFrom().Index,
	}

	// Create tss message
	msg := tsspb.TssMessage{
		From:        from,
		To:          destinations,
		IsBroadcast: libMessage.IsBroadcast(),
		Payload:     payload,
	}

	return &msg, nil
}

// sortParties returns sorted parties.
// Sort parties with party key if newKey set to false,
// otherwise, sort parties with new party key.
func sortParties(parties []tsspb.Party, newKey bool) tssBinance.SortedPartyIDs {
	var pIDs []*tssBinance.PartyID

	// Append partyIDs for parties
	for _, p := range parties {
		pIDs = append(pIDs, convertParty2PartyID(p, newKey))
	}

	// Sort PartyIDs
	return tssBinance.SortPartyIDs(pIDs)
}

// updateIndex update index for this party in sorted parties
// If newKey is true, update index for new party in sorted parties
func (t *tssSigner) updateIndex(sortedParties tssBinance.SortedPartyIDs, newKey bool) {
	//fmt.Println("updateIndex", newKey)
	//fmt.Println("updateIndex", sortedParties.Keys())
	//fmt.Println("updateIndex", t.newPartyId.KeyInt())
	//fmt.Println("updateIndex", t.partyId.KeyInt())

	if newKey {
		pid := sortedParties.FindByKey(t.newPartyId.KeyInt())
		if pid == nil {
			panic(ErrPartyNotFound)
		}
		t.newPartyId.Index = pid.Index
	} else {
		pid := sortedParties.FindByKey(t.partyId.KeyInt())
		if pid == nil {
			panic(ErrPartyNotFound)
		}
		t.partyId.Index = pid.Index
	}
}

// classifyPartiesByKey classify parties by shared key
func classifyPartiesByKey(parties []tsspb.Party) (withSharedKey []tsspb.Party, noSharedKey []tsspb.Party) {
	for _, p := range parties {
		if p.HaveSharedKey {
			withSharedKey = append(withSharedKey, p)
		} else {
			noSharedKey = append(noSharedKey, p)
		}
	}

	return withSharedKey, noSharedKey
}

// partiesWithKey return peer IDs in peers that have private key
func partiesWithKey(parties []tsspb.Party) []tsspb.Party {
	// Find peer IDs that have private key
	var withPrivateKey []tsspb.Party
	for _, p := range parties {
		if p.HaveSharedKey {
			withPrivateKey = append(withPrivateKey, p)
		}
	}
	return withPrivateKey
}

// convertParty2PartyId convert my party model to tss partyID model
// If newKey is true, used NewPartyKey for convert to partyID
func convertParty2PartyID(party tsspb.Party, newKey bool) *tssBinance.PartyID {
	if newKey {
		return tssBinance.NewPartyID(party.ID, "", party.NewPartyKey)
	} else {
		return tssBinance.NewPartyID(party.ID, "", party.PartyKey)
	}
}

func (t *tssSigner) makePublicKey() error {
	if t.privateKey == nil || t.privateKey.ECDSAPub == nil {
		return fmt.Errorf("ECDSAPub is nil in shared key")
	}

	pk := ecdsa.PublicKey{
		Curve: t.privateKey.ECDSAPub.Curve(),
		X:     t.privateKey.ECDSAPub.X(),
		Y:     t.privateKey.ECDSAPub.Y(),
	}

	t.publicKey = hexutil.Encode(crypto.FromECDSAPub(&pk))

	// Create address from public key
	t.address = crypto.PubkeyToAddress(pk).Hex()
	log.Info("MAKE PUBLIC KEY", aurora.Green(fmt.Sprint("address:", t.address)))

	return nil
}
