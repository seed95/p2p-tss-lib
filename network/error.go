package network

import "fmt"

type Error struct {
	Err error
	ID  string
}

func (e Error) Error() string {
	return fmt.Sprintf("ID: %s, error: %v", e.ID, e.Err)
}
