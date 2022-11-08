package message

import (
	"encoding/json"
	"math/big"
)

type PartyId struct {
	PartyKey    *big.Int `json:"partyKey"`
	NewPartyKey *big.Int `json:"newPartyKey"`
}

func (pi *PartyId) Marshal() ([]byte, error) {
	return json.Marshal(pi)
}

func (pi *PartyId) Unmarshal(data []byte) error {
	return json.Unmarshal(data, pi)
}
