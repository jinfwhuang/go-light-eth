package discover

import (
	"crypto/md5"
	crand "crypto/rand"
	"encoding/json"
	tmplog "log"
	"math/rand"
	"time"

	//"encoding/binary"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/jinfwhuang/go-light-eth/pkg/discover/v5wire"
	"net"
	//"hash/fnv"
	//"crypto"
	"strings"
)

const (
	//lookupRequestLimit      = 3  // max requests against a single node during lookup
	//findnodeResultLimit     = 16 // applies in FINDNODE handler
	//totalNodesResponseLimit = 5  // applies in waitForNodes
	//nodesResponseItemLimit  = 3  // applies in sendNodes
	//
	//respTimeoutV5 = 3000 * time.Millisecond
)

var (
	TalkExtProtocol = "talk-sctp"
)


func (t *UDPv5) RegisterTalkExtHandler(protocol string, handler TalkRequestHandler) {
	//t.trlock.Lock()
	//defer t.trlock.Unlock()
	//t.trhandlers[protocol] = handler
}

func (t *UDPv5) TalkRequestExt(n *enode.Node, protocol string, request []byte) ([]byte, error) {
	protocolExt := toTalkExtProtocol(protocol)

	// Split up the request into packets
	connId := rand.Uint64()
	packets := splitMessage(ConnectionId(connId), request)
	talkConn := TalkExtConnection {
		Id:         packets[0].Id,
		LastSeqNum: packets[0].LastSeqNum,
		Packets: make(map[int]TalkExtPacket),
	}
	t.TalkExtConnections[talkConn.Id] = talkConn

	// Send all packets
	for _, packet := range packets {
		// TODO: the TalkRequest protocol could be modified in such a way that it does not respond to any TalkRequestExt with TalkResponse
		_r, err := t.TalkRequest2(n, protocolExt, TalkExtPacketToMessage(packet))
		if err != nil {
			tmplog.Fatal(string(_r)) // ignore these responses; the real response is streamed
		}
	}

	// Wait for a response stream to complete
	// Construct response
	deadline := time.Now().Add(30 * time.Second)
	for {
		if talkConn.completed() {
			// construct response
			response := constructResponse(talkConn)
			return response, nil
		}
		if time.Now().After(deadline) {
			//return nil, fmt.Errorf("timeout")
			tmplog.Fatal("timeout")
		}
	}
}



// TalkRequest sends a talk request to n and waits for a response.
func (t *UDPv5) TalkRequest2(n *enode.Node, protocol string, request []byte) ([]byte, error) {
	req := &v5wire.TalkRequest{Protocol: protocol, Message: request}
	resp := t.call(n, v5wire.TalkResponseMsg, req)
	defer t.callDone(resp)
	select {
	case respMsg := <-resp.ch:
		return respMsg.(*v5wire.TalkResponse).Message, nil
	case err := <-resp.err:
		return nil, err
	}
}

// call sends the given call and sets up a handler for response packets (of message type
// responseType). Responses are dispatched to the call's response channel.
func (t *UDPv5) callTalkExt(node *enode.Node, responseType byte, packet v5wire.Packet) *callV5 {
	c := &callV5{
		node:         node,
		packet:       packet,
		responseType: responseType,
		reqid:        make([]byte, 8),
		ch:           make(chan v5wire.Packet, 1),
		err:          make(chan error, 1),
	}
	// Assign request ID.
	crand.Read(c.reqid)
	packet.SetRequestID(c.reqid)
	// Send call to dispatch.
	select {
	case t.callCh <- c:
	case <-t.closeCtx.Done():  // There is data in the t.closeCtx.Done channel
		c.err <- errClosed
	}
	return c
}

func splitMessage(id ConnectionId, msg []byte) []TalkExtPacket {
	n := len(msg)
	step := 500
	size := n / step
	if size * step < n {
		size += 1 // adjusting for remainder
	}
	packets := make([]TalkExtPacket, size)
	for i := 0; i < size; i++ {
		tmplog.Println(i, i *step, n)
		left := i * step
		right := (i+1) * step
		if n < right {
			right = n
		}
		chunk := msg[left:right]
		packet := TalkExtPacket{
			Id: id,
			SeqNum: i,
			LastSeqNum: size,
			Packet: chunk,
		}
		packets[i] = packet
	}
	return packets
}

func combinePackets(packets []TalkExtPacket) []byte {
	out := make([]byte, 0)
	for _, packet := range packets {
		out = append(out, packet.Packet...)
	}
	return out
}

func constructResponse(talkConn TalkExtConnection) []byte {
	packets := make([]TalkExtPacket, talkConn.LastSeqNum)
	for i := 0; i < talkConn.LastSeqNum; i++ {
		packets[i] = talkConn.Packets[i]
	}
	return combinePackets(packets)
}




// TODO: Use ssz serialization later on
type TalkExtPacket struct {
	Id     ConnectionId
	SeqNum int
	LastSeqNum int
	Packet []byte
}

func MessageToTalkExtPacket(b []byte) TalkExtPacket {
	packet := TalkExtPacket{}
	err := json.Unmarshal(b, &packet)
	if err != nil {
		tmplog.Fatal(err)
	}
	return packet
}

func TalkExtPacketToMessage(packet TalkExtPacket) []byte {
	ser, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		tmplog.Fatal(err)
	}
	return ser
}

type TalkExtConnection struct {
	Id         ConnectionId
	LastSeqNum int // 0, 1, 2, ... LastSeqNum - 1
	Packets map[int]TalkExtPacket // keyed by SeqNum
}

func (c *TalkExtConnection) completed() bool {
	packetCounts := 0
	for range c.Packets {
		packetCounts++
	}
	// The connection is completed iff all the packets have arrived
	return packetCounts == c.LastSeqNum
}

func toTalkExtProtocol(proc string) string {
	b := md5.Sum([]byte(proc))
	return proc + "-" + hex.EncodeToString(b[:5])
}

func isTalkExtProtocol(proc string) bool {
	parts := strings.Split(proc, "-")
	inferredProc := toTalkExtProtocol(parts[0])

	return inferredProc == proc
}

/**
1. If TalkRequestExt, extract connection_id and put data on a map

2.

TalkExtConnections := map[int]TalkExtConnection

TODO: SERVER action???

*/
func (t *UDPv5) handleTalkExt(p *v5wire.TalkRequest, fromID enode.ID, fromAddr *net.UDPAddr) {
	if isTalkExtProtocol(p.Protocol) {
		// If it is completed, construct response???
		// The response is a bunch of "TalkExtPacket"
		packet := MessageToTalkExtPacket(p.Message)

		// Put the packets into the connections DB
		//var talkConn TalkExtConnection
		talkConn := TalkExtConnection {
			Id:         packet.Id,
			LastSeqNum: packet.LastSeqNum,
			Packets: map[int]TalkExtPacket{
				packet.SeqNum: packet,
			},
		}
		if _talkConn, ok := t.TalkExtConnections[packet.Id]; ok {
			talkConn = _talkConn
			talkConn.Packets[packet.SeqNum] = packet // Update packet
			// TODO: turn these into a proper data structure and use instance methods
		} else {
			t.TalkExtConnections[talkConn.Id] = talkConn // Create a new TalkExtConnection entry in the DB
		}
		if talkConn.completed() {
			// All the packets have been received, we should send the response
			// TODO: xxx; somehow use the handler

			//t.trlock.Lock()
			//handler := t.TalkExtHandlers[p.Protocol]
			//t.trlock.Unlock()


			//var response []byte
			//if handler != nil {
			//	response = handler(fromID, fromAddr, p.Message)
			//}


			//resp := &v5wire.TalkResponse{ReqID: p.ReqID, Message: response}
			//t.sendResponse(fromID, fromAddr, resp)



		}
	} else {  // Normal TalkRequest handling
		t.handleTalkRequest(p, fromID, fromAddr)
	}
}

