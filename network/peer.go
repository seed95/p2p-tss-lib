package network

import (
	"fmt"
	"math/big"

	"github.com/libp2p/go-libp2p-core/network"
	
	p2peer "github.com/seed95/p2p-tss-lib/peer"
	tsspb "github.com/seed95/p2p-tss-lib/tss/pb"
)

type (
	Peer interface {
		p2peer.Peer
		PartyKey() *big.Int
		NewPartyKey() *big.Int
		HasPrivateKey() bool
		SetHasPrivateKey(hasKey bool)
		UpdateParty(party tsspb.Party)
		GetParty() tsspb.Party
	}

	peer struct {
		p2peer.Peer

		party tsspb.Party
	}
)

var _ Peer = (*peer)(nil)

func NewPeer(stream network.Stream) (Peer, error) {
	if stream == nil {
		return nil, fmt.Errorf("invalid stream")
	}

	return &peer{
		Peer: p2peer.NewPeer(stream),
	}, nil
}

func (p *peer) PartyKey() *big.Int {
	return p.party.PartyKey
}

func (p *peer) NewPartyKey() *big.Int {
	return p.party.NewPartyKey
}

func (p *peer) HasPrivateKey() bool {
	return p.party.HaveSharedKey
}

func (p *peer) SetHasPrivateKey(hasKey bool) {
	p.party.HaveSharedKey = hasKey
}

func (p *peer) UpdateParty(party tsspb.Party) {
	p.party = party
}

func (p *peer) GetParty() tsspb.Party {
	return p.party
}
