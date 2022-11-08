package pb

import (
	"encoding/json"
	"fmt"
)

const (
	OldReshareType = 1
	NewReshareType = 2
)

type TssMessage struct {
	From        Party    `json:"from"`
	To          []string `json:"to"`
	IsBroadcast bool     `json:"isBroadcast"`
	Payload     []byte   `json:"payload"`
	ReshareType uint     `json:"reshareType"` // ReshareType used in reshare operation, for sending message to update old reshare party or new reshare party
}

func (t *TssMessage) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

func (t *TssMessage) Unmarshal(data []byte) error {
	return json.Unmarshal(data, t)
}

func (t *TssMessage) String() string {
	toStr := "all"
	if t.To != nil {
		toStr = fmt.Sprintf("%v", t.To)
	}
	return fmt.Sprintf("to: %s, is broadcast: %v, size: %v", toStr, t.IsBroadcast, len(t.Payload))
}
