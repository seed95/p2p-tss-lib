package peer

import (
	"bufio"
	"fmt"
	"sync"

	"github.com/seed95/p2p-tss-lib/peer/pb"

	"github.com/libp2p/go-libp2p-core/network"
	peerLib "github.com/libp2p/go-libp2p-core/peer"
	"github.com/logrusorgru/aurora"
)

type AddrInfo = peerLib.AddrInfo

const delimiter = '\n' //TODO: Is valid delim??

type (
	Peer interface {
		//Run(receivedMsg chan<- pb.Message) *Error // TODO check error need
		Close() error // TODO check error need

		ID() string

		// SendMsg send msg to peer.
		// safe concurrency mode
		SendMsg(msg pb.Message) error

		ReadData() (*pb.Message, error)
	}

	peer struct {
		stream network.Stream
		rw     *bufio.ReadWriter

		// mu used to send message to peer in concurrency mode.
		mu sync.Mutex

		// quit used to stop peer in Run function.
		// When the peer is close, will be sent to this channel to stop the program in the Run function
		//quit chan struct{}
	}
)

var _ Peer = (*peer)(nil)

func NewPeer(stream network.Stream) Peer {
	p := &peer{
		//quit:     make(chan struct{}),
	}

	if stream != nil {
		p.stream = stream
		p.rw = bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	}

	return p
}

func (p *peer) Run(receiverCh chan<- pb.Message) *Error {
	// Read data forever
	for {
		data, err := p.rw.ReadBytes(delimiter)
		if err != nil {
			err = fmt.Errorf("error reading bytes: %v", err)
			return &Error{Message: err.Error(), ID: p.ID()}
		}

		// Remove delimiter
		data = data[:len(data)-1]

		// Unmarshal received message
		msg := pb.Message{}
		err = msg.Unmarshal(data)
		if err != nil {
			err = fmt.Errorf("error un-marshalling message: %v", err)
			return &Error{Message: err.Error(), ID: p.ID()}
		}
		msg.From = p.ID()

		fmt.Println(aurora.Sprintf(aurora.Blue("Received message from peer: %v, code: %v, size: %v"), p.ID(), msg.Code, len(data)))

		receiverCh <- msg
	}
}

func (p *peer) Close() error {
	// Stop read data
	//p.quit <- struct{}{}

	if err := p.stream.Close(); err != nil {
		return err
	}
	return nil
}

func (p *peer) ID() string {
	return p.stream.Conn().RemotePeer().String()
}

func (p *peer) SendMsg(msg pb.Message) error {
	data, err := msg.Marshal()
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	_, err = p.rw.Write(append(data, delimiter))
	if err != nil {
		return fmt.Errorf("error writing to stream: %v", err)
	}

	err = p.rw.Flush()
	if err != nil {
		return fmt.Errorf("error flushing stream: %v", err)
	}

	return nil
}

func (p *peer) ReadData() (*pb.Message, error) {
	data, err := p.rw.ReadBytes(delimiter)
	if err != nil {
		return nil, fmt.Errorf("error reading bytes: %v", err)
	}

	// Remove delimiter
	data = data[:len(data)-1]

	// Unmarshal received data
	msg := pb.Message{}
	err = msg.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("error un-marshalling message: %v", err)
	}
	msg.From = p.ID()

	return &msg, nil
}
