package p2ptsslib

import "fmt"

var (
	ErrAnotherProcessIsRunning = fmt.Errorf("another distributed process is already running")

	ErrKeyIsNotGenerated = fmt.Errorf("node doesn't shared key. please generate key or reshare")

	ErrNotEnoughPeer = fmt.Errorf("do not have enough peer")
	ErrInvalidPeer   = fmt.Errorf("new peer ID is equal to node")
	ErrDuplicatePeer = fmt.Errorf("new peer already exists")
)
