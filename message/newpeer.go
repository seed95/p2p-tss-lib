package message

import (
	"encoding/json"
	"fmt"
	peer_lib "github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type NewPeer struct {
	ID      string   `json:"id"`
	Address []string `json:"address"`
}

func (np *NewPeer) Marshal() ([]byte, error) {
	return json.Marshal(np)
}

func (np *NewPeer) Unmarshal(data []byte) error {
	return json.Unmarshal(data, np)
}

func (np *NewPeer) String() string {
	return fmt.Sprintf("{%v: %v}", np.ID, np.Address)
}

func (np *NewPeer) AddrInfo() (addrsInfo peer_lib.AddrInfo, err error) {
	// Convert ID to P2PLib ID
	hostID, err := peer_lib.Decode(np.ID)
	if err != nil {
		return addrsInfo, err
	}

	// Convert addrs to Multiaddr
	var mas []multiaddr.Multiaddr
	for _, addr := range np.Address {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return addrsInfo, err
		}
		mas = append(mas, ma)
	}

	addrsInfo = peer_lib.AddrInfo{ID: hostID, Addrs: mas}
	return addrsInfo, nil
}
