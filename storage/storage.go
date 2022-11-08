package storage

// Storage is an interface for storing some values
// in long-term memory like that database
type Storage interface {
	// StorePrivateKey stores the shared key of party.
	StorePrivateKey(key []byte) error

	// PrivateKey return private key in storage.
	// If it does not exist return nil.
	PrivateKey() []byte
}
