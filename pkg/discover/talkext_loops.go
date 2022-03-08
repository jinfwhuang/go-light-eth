package discover

import (
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/jinfwhuang/go-light-eth/pkg/discover/v5wire"
	log "github.com/sirupsen/logrus"
	tmplog "log"
	"net"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func (t *UDPv5) talkextReadLoop() {
	tmplog.Println("readloop is started")
	for item := range t.TalkExtReadCh {
		//req := item.talkExt
		tmplog.Println("readloop action")
		t.handleTalkExt(item.talkExt, item.fromID, item.fromAddr)
	}
	tmplog.Println("readloop is done")
}

func (t *UDPv5) talkextWriteLoop() {
	tmplog.Println("writeloop is started")
	for c := range t.TalkExtWriteCh {
		tmplog.Println("writeloop action", c.reqid)
		addr := &net.UDPAddr{IP: c.node.IP(), Port: c.node.UDP()}
		_, err := t.rawsend(c.node.ID(), addr, c.packet, c.challenge)
		if err != nil {
			tmplog.Fatal(err)
		}
	}
	tmplog.Println("writeloop is done")
}

// send sends a packet to the given node.
func (t *UDPv5) rawsend(toID enode.ID, toAddr *net.UDPAddr, packet v5wire.Packet, c *v5wire.Whoareyou) (v5wire.Nonce, error) {
	addr := toAddr.String()
	enc, nonce, err := t.codec.Encode(toID, addr, packet, c)
	if err != nil {
		log.Fatal(err)
	}

	//fromID, fromNode, packet, err := t.codec.Decode(enc, addr)
	//tmplog.Println("enc", len(enc))
	//tmplog.Println("dec", packet, "fromID", fromID, "fromNode", fromNode)
	//if err != nil {
	//	tmplog.Fatal(err)
	//}

	tmplog.Println("udp write", packet.Name(), len(enc), "to", toAddr, "sender-whoareyou", c)

	_, err = t.conn.WriteToUDP(enc, toAddr)
	t.log.Info(">> "+packet.Name(), "Id", toID, "addr", addr)
	return nonce, err
}

/**
So to speak: Server action

Responding to a TalkExt and TalkRequest
*/
func (t *UDPv5) handleTalkExt(p *v5wire.TalkExt, fromID enode.ID, fromAddr *net.UDPAddr) {
	packet := TalkExtPacket{}
	packet.unmarshal(p.Message)
	tmplog.Println("server action |", packet.Id)
	talkConn := &TalkExtConnection{
		Id:         packet.Id,
		LastSeqNum: packet.LastSeqNum,
		IncomingPackets: map[int]TalkExtPacket{
			packet.SeqNum: packet, // This packet as the first packet
		},
	}
	if _talkConn, ok := t.TalkExtConnections[packet.Id]; ok {
		// TODO: turn these into a proper data structure and use instance methods
		talkConn = _talkConn
		talkConn.IncomingPackets[packet.SeqNum] = packet // Record the incoming packet
	} else {
		t.TalkExtConnections[talkConn.Id] = talkConn // Create a new TalkExtConnection entry in the DB
	}

	// If packet stream is completed, send the response as another stream
	talkConn.mutexLock.Lock()
	if talkConn.completed() {
		// Construct the response as []byte
		talkMsg := talkConn.getMessageFromPackets()
		tmplog.Println("got a complete stream message", string(talkMsg))
		handler := t.TalkExtHandlers[p.Protocol]
		resp := make([]byte, 0)
		if handler != nil {
			resp = handler(fromID, fromAddr, talkMsg)
		}

		tmplog.Println("constructed response:", string(resp))

		// Send the TalkExtResp
		t.sendTalkExtResp(p, fromID, fromAddr, talkConn, resp)
	}
	talkConn.mutexLock.Unlock()
}