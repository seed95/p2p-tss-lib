package storage

import (
	"sync"
)

// storage is a simple in-memory implementation of Storage
type storage struct {
	sync.RWMutex
	sharedKey []byte
}

var _ Storage = (*storage)(nil)

// NewDefaultStorage returns a new simple in-memory storage
func NewDefaultStorage() Storage {
	s := storage{}
	return &s
}

func (s *storage) StorePrivateKey(key []byte) error {
	s.Lock()
	defer s.Unlock()
	s.sharedKey = key
	return nil
}

func (s *storage) PrivateKey() []byte {
	return s.sharedKey
}
