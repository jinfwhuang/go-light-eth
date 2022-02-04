package portalnet

import (
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/jinfwhuang/go-light-eth/pkg/discover"
	"github.com/pion/logging"
	"github.com/pion/sctp"
	log "github.com/sirupsen/logrus"
	tmplog "log"
	"net"
	"sync"
	"time"
)

var (
	TalkExtProtocol = "talk-sctp"
)

func aaa(udpv5 *discover.UDPv5) {

	//udpv5.Ta


}


func setupTalkExt(udpv5 *discover.UDPv5) {
	udpv5.RegisterTalkHandler(TalkExtProtocol, func (node enode.ID, addr *net.UDPAddr, input []byte) []byte {
		//return append(input, []byte(nodename + "responded")...)

		tmplog.Println(len(input))

		return input
	})


}

func TalkRequestExt(udpv5 *discover.UDPv5, n *enode.Node, protocol string, request []byte) ([]byte, error) {

	config := sctp.Config{
		NetConn:       &FakeConn{
			udpv5: udpv5,
			n: n,
		},
		LoggerFactory: logging.NewDefaultLoggerFactory(),
	}
	tmplog.Println(config)

	sctpClient, err := sctp.Client(config)
	if err != nil {
		log.Panic(err)
	}
	tmplog.Println("created a client")

	clientStream, err := sctpClient.OpenStream(0, sctp.PayloadTypeWebRTCString)
	_, err = clientStream.Write([]byte(request))
	if err != nil {
		log.Panic(err)
	}

	for {
		buff := make([]byte, 10000)
		_, err = clientStream.Read(buff)
		if err != nil {
			log.Panic(err)
		}
		return buff, nil
	}
}



/**
Buid a Conn on top of TalkRequest
 */
type FakeConn struct { // nolint: unused
	mu    sync.RWMutex
	//rAddr net.Addr
	//pConn net.PacketConn

	udpv5 *discover.UDPv5

	n *enode.Node

	//buffer [][]byte

	buffer FifoQueue

	protocol string

}


func (c *FakeConn) Write(data []byte) (n int, err error) {
	c.mu.Lock()

	tmplog.Println(c.n.IP(), data, TalkExtProtocol)
	resp, err := c.udpv5.TalkRequest(c.n, TalkExtProtocol, data)
	if err != nil {
		tmplog.Fatal(err)
	}
	c.buffer.Enqueue(resp)

	c.mu.Unlock()

	return len(resp), nil
}

// Read
func (c *FakeConn) Read(p []byte) (int, error) {
	c.mu.Lock()
	b := c.buffer.Dequeue()

	copy(p, b)

	c.mu.Unlock()

	return len(b), nil
}


// Close closes the conn and releases any Read calls
func (c *FakeConn) Close() error {
	return nil
}

// LocalAddr is a stub
func (c *FakeConn) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr is a stub
func (c *FakeConn) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline is a stub
func (c *FakeConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline is a stub
func (c *FakeConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline is a stub
func (c *FakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}
