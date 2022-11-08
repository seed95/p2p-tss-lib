package pb

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type Party struct {
	ID            string   `json:"id"`
	PartyKey      *big.Int `json:"partyKey"`
	NewPartyKey   *big.Int `json:"newPartyKey"`
	HaveSharedKey bool     `json:"haveSharedKey"`
	Index         int      `json:"index"` // Index used when update data party for sorted parties (keygen, sign, reshare)
}

func (p *Party) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Party) Unmarshal(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *Party) String() string {
	return fmt.Sprintf("{ID: %v, partyKey: %v, newPartyKey: %v, haveSharedKey: %v, index: %v}",
		p.ID, p.PartyKey, p.NewPartyKey, p.HaveSharedKey, p.Index)
}
