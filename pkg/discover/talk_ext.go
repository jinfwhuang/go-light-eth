package discover

import (
	crand "crypto/rand"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/jinfwhuang/go-light-eth/pkg/discover/v5wire"
	tmplog "log"
	"net"
)

func (t *UDPv5) RegisterTalkExtHandler(protocol string, handler TalkRequestHandler) {
	t.TalkExtLock.Lock()
	defer t.TalkExtLock.Unlock()


	tmplog.Println(protocol)
	protocol = asTalkExtProtocol(protocol)
	tmplog.Println(protocol)
	t.TalkExtHandlers[protocol] = handler

	// Register an empty TalkHandler
	t.RegisterTalkHandler(protocol, func (node enode.ID, addr *net.UDPAddr, input []byte) []byte {
		tmplog.Println("empty handler invoked")
		return make([]byte, 0)
	})
}

/**
So to speak: Client action
 */
func (t *UDPv5) TalkRequestExt(n *enode.Node, protocol string, request []byte) ([]byte, error) {
	protocol = asTalkExtProtocol(protocol)

	// Setup TalkConn
	talkConn := NewTalkExtConnection()
	t.TalkExtConnections[talkConn.Id] = talkConn

	// Send all outgoing packets
	packets := talkConn.generatePackets(request)
	tmplog.Println("start sending packets")
	for i, packet := range packets {
		toAddr := &net.UDPAddr{IP: n.IP(), Port: n.UDP()}

		tmplog.Println(i, packet)
		t.rawSendTalkRequest2(n, n.ID(), toAddr, protocol, packet.marshal())
		tmplog.Println("finish", i)

		//t.TalkRequest(n, protocol, packet.marshal())
		//return nil, nil
	}
	tmplog.Println("finished sending packets")
	return nil, nil

	//// Wait for a response stream to complete, i.e. all the response packets to arrive
	//// Construct response
	//deadline := time.Now().Add(5 * time.Second)
	//for {
	//	//tmplog.Println("completed?", talkConn.completed())
	//	if talkConn.completed() {
	//		return talkConn.getMessageFromPackets(), nil
	//	}
	//	if time.Now().After(deadline) {
	//		return nil, fmt.Errorf("timeout")
	//	}
	//	time.Sleep(time.Millisecond * 2000) // TODO: use a signaling channel instead
	//}
}

/**
So to speak: Server action

Responding to a TalkExt and TalkRequest
*/
func (t *UDPv5) handleTalkExt(p *v5wire.TalkRequest, fromID enode.ID, fromAddr *net.UDPAddr) {
	tmplog.Println("ffff getting a TalkRequest wired msg")

	// TalkRequest
	//----------------
	if !isTalkExtProtocol(p.Protocol) {
		t.handleTalkRequest(p, fromID, fromAddr)
		return
	}

	// TalkExt
	//----------------

	//tmplog.Println(p.Protocol, p.Message)
	packet := TalkExtPacket{}
	packet.unmarshal(p.Message)

	tmplog.Println(packet)

	//tmplog.Println("getting a proper TalkExt packet", packet)
	//
	//tmplog.Println(fromID, fromAddr)
	//return
	//// TODO: remove

	// Put this packet into the connections DB
	//talkConn := NewTalkExtConnection()

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
		tmplog.Println("gotten a complete stream")
		// Construct the response as []byte
		talkMsg := talkConn.getMessageFromPackets()
		tmplog.Println("got a msg:", string(talkMsg))
		handler := t.TalkExtHandlers[p.Protocol]
		resp := make([]byte, 0)
		if handler != nil {
			resp = handler(fromID, fromAddr, talkMsg)
		}

		tmplog.Println("constructed response:", string(resp))
		// Send the TalkExt response
		t.sendTalkExtResp(p, fromID, fromAddr, talkConn, resp)
	}
	talkConn.mutexLock.Unlock()
}

func (t *UDPv5) lookupWithCache(id enode.ID) *enode.Node {
	n := t.tab.getNode(id)
	if n == nil {
		tmplog.Fatal("cannot find node")
	}
	//if n != nil {
	//	ns := t.Lookup(id)
	//
	//
	//
	//} else {
	//	tmplog.Println("found cache")
	//}
	return n
}

/**
1. Split []byte into packets
2. Send all packets individually as TalkRequests

Note:
- The ConnectionID is the same as the incoming
 */
func (t *UDPv5) sendTalkExtResp(p *v5wire.TalkRequest, fromID enode.ID, fromAddr *net.UDPAddr, talkConn *TalkExtConnection, data []byte) {
	packets := talkConn.generatePackets(data)
	protocol := p.Protocol  // Must be protocol-ext

	tmplog.Println(protocol)
	tmplog.Println(packets)

	//nn := &enode.Node{}
	nn := t.lookupWithCache(fromID)
	tmplog.Println("responding to node", nn.IP(), nn.UDP())
	// Send all packets
	for i, packet := range packets {
		tmplog.Println(i,packet)
		t.rawSendTalkRequest2(nn, fromID, fromAddr, protocol, packet.marshal())
		tmplog.Println("finished", i)
	}
	tmplog.Println("finished responding to", nn.IP(), nn.UDP())

}

//// Send bytes as TalkRequest without setting up any response type handling
//func (t *UDPv5) rawSendTalkRequest(toId enode.ID, toAddr *net.UDPAddr, protocol string, msg []byte) {
//	req := &v5wire.TalkRequest{Protocol: protocol, Message: msg}
//	crand.Read(req.ReqID)  // set a random ReqID
//
//	tmplog.Println(toId, toAddr, req)
//	once, err := t.send(toId, toAddr, req, nil)  // raw send; not handling TalkRespnoses
//	if err != nil {
//		tmplog.Fatal(err)
//	}
//	tmplog.Println(once)
//
//
//}

// Send bytes as TalkRequest without setting up any response type handling
func (t *UDPv5) rawSendTalkRequest2(n *enode.Node, toId enode.ID, toAddr *net.UDPAddr, protocol string, msg []byte) {
	t.TalkRequestWithoutWaiting(n, protocol, msg)  // TODO: we don't have to wait for the response
	//if err != nil {
	//	tmplog.Println(err)
	//}
	tmplog.Println("finished rawsend")
}

func (t *UDPv5) TalkRequestWithoutWaiting(n *enode.Node, protocol string, request []byte) {
	req := &v5wire.TalkRequest{Protocol: protocol, Message: request}
	tmplog.Println("starting calling an action")
	resp := t.call(n, v5wire.TalkResponseMsg, req)
	tmplog.Println("finish calling an action")

	defer t.callDone(resp)  // TODO: this need to be completed

	//go func() {
	//	select {
	//	case respMsg := <-resp.ch:
	//		tmplog.Println("only consuming, but don't really care", respMsg)
	//	case err := <-resp.err:
	//		tmplog.Println("only consuming, but don't really care", err)
	//	}
	//}()
	tmplog.Println("herer")
}







// TalkRequest sends a talk request to n and waits for a response.
func (t *UDPv5) ___TalkRequest(n *enode.Node, protocol string, request []byte) ([]byte, error) {
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
func (t *UDPv5) ___call(node *enode.Node, responseType byte, packet v5wire.Packet) *callV5 {
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


// handleTalkRequest runs the talk request handler of the requested protocol.
func (t *UDPv5) ___handleTalkRequest(p *v5wire.TalkRequest, fromID enode.ID, fromAddr *net.UDPAddr) {
	t.trlock.Lock()
	handler := t.trhandlers[p.Protocol]
	t.trlock.Unlock()

	var response []byte
	if handler != nil {
		response = handler(fromID, fromAddr, p.Message)
	}
	resp := &v5wire.TalkResponse{ReqID: p.ReqID, Message: response}
	t.sendResponse(fromID, fromAddr, resp)
}

