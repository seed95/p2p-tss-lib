package pb

//TODO generate with protobuf

// TODO this file to pb in root package

import (
	"encoding/json"
	"fmt"
	tssBinance "github.com/binance-chain/tss-lib/tss"
)

type MessageCode uint

//TODO check values
const (
	PartyIdCode          MessageCode = 0x01
	StartKeyGenCode      MessageCode = 0x02
	UpdateKeyGenCode     MessageCode = 0x03
	StartSigningCode     MessageCode = 0x04
	UpdateSignCode       MessageCode = 0x05
	StartReshareCode     MessageCode = 0x06
	UpdateReshareCode    MessageCode = 0x07
	UpdateNewReshareCode MessageCode = 0x08
	NewReShareCode       MessageCode = 0x09
)

type Message struct {
	From        *tssBinance.PartyID   `json:"from"`
	To          []*tssBinance.PartyID `json:"to"`
	Code        MessageCode           `json:"code"`
	IsBroadcast bool                  `json:"isBroadcast"` //TODO is needed?
	Payload     []byte                `json:"payload"`
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *Message) String() string {
	toStr := "all"
	if m.To != nil {
		toStr = fmt.Sprintf("%v", m.To)
	}
	return fmt.Sprintf("Code: %s, From: %s, To: %s", m.CodeString(), m.From.String(), toStr)
}

func (m *Message) CodeString() string {
	switch m.Code {

	case PartyIdCode:
		return "PartyIdCode"

	case StartKeyGenCode:
		return "StartKeyGenCode"

	case StartReshareCode:
		return "StartReshareCode"

	case UpdateKeyGenCode:
		return "UpdateKeyGenCode"

	case UpdateSignCode:
		return "UpdateSignCode"

	case UpdateReshareCode:
		return "UpdateReshareCode"

	case UpdateNewReshareCode:
		return "UpdateNewReshareCode"

	default:
		return "unknown"

	}
}

type SigningDataType string

const (
	SignEthTransactionType = "ethTransaction"
	SignMessageType        = "message"
)

type Sign struct {
	IDs  tssBinance.SortedPartyIDs `json:"ids"`
	Data []byte                    `json:"data"`
	Type SigningDataType           `json:"type"`
}

func (s *Sign) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Sign) Unmarshal(data []byte) error {
	return json.Unmarshal(data, s)
}
