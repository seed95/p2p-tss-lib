package message

import (
	"encoding/json"
)

type StartKeyGen struct {
}

func (np *StartKeyGen) Marshal() ([]byte, error) {
	return json.Marshal(np)
}

func (np *StartKeyGen) Unmarshal(data []byte) error {
	return json.Unmarshal(data, np)
}
