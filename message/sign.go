package message

import (
	"encoding/json"
	"fmt"
	tss_pb "github.com/seed95/p2p-tss-lib/tss/pb"
)

type Sign struct {
	Parties []tss_pb.Party `json:"parties"`
	Data    []byte         `json:"data"`
	Type    string         `json:"type"`
}

func (s *Sign) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Sign) Unmarshal(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *Sign) String() string {
	return fmt.Sprintf("{parties: %v, size: %v, type: %v}", s.Parties, len(s.Data), s.Type)
}
