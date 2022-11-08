package pb

//TODO must be declare with protobuf

import (
	"encoding/json"
	"time"
)

type (
	Message struct {
		From     string    `json:"from"` // peer ID that sent message
		Code     uint      `json:"code"`
		Payload  []byte    `json:"payload"`
		Deadline time.Time `json:"deadline"`
	}
)

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

type DeletePeer struct {
	ID string `json:"id"`
}

func (m *DeletePeer) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *DeletePeer) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}
