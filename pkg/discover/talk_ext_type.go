package discover

import (
	"crypto/md5"
	"encoding/json"
	tmplog "log"
	"math/rand"
	"sync"
	//"encoding/binary"
	"encoding/hex"
	//"hash/fnv"
	//"crypto"
	"strings"
)



func combinePackets(packets []TalkExtPacket) []byte {
	out := make([]byte, 0)
	for _, packet := range packets {
		out = append(out, packet.Packet...)
	}
	return out
}


// TODO: Use ssz serialization later on
type TalkExtPacket struct {
	Id     ConnectionId
	SeqNum int
	LastSeqNum int
	Packet []byte
}

func (p *TalkExtPacket) marshal() []byte {
	ser, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		tmplog.Fatal(err)
	}
	return ser
}

func (p *TalkExtPacket) unmarshal(data []byte) {
	err := json.Unmarshal(data, p)
	if err != nil {
		tmplog.Fatal(err)
	}
}


type TalkExtConnection struct {
	Id              ConnectionId
	LastSeqNum      int                   // 0, 1, 2, ... LastSeqNum - 1
	IncomingPackets map[int]TalkExtPacket // keyed by SeqNum

	mutexLock   sync.Mutex
	completedCh chan interface{} // whenever completed send a message here
}

func NewTalkExtConnection() *TalkExtConnection {
	connId := rand.Uint64()

	return &TalkExtConnection{
		Id: ConnectionId(connId),
		LastSeqNum: -1,
		IncomingPackets: make(map[int]TalkExtPacket),
		completedCh: make(chan interface{}),
	}
}

func (c *TalkExtConnection) hasFirstPacket() bool {
	return c.LastSeqNum != -1
}

func (c *TalkExtConnection) completed() bool {
	if !c.hasFirstPacket() {
		tmplog.Println("TalkExtConnection not initialized")
		return false
	}
	packetCounts := 0
	for range c.IncomingPackets {
		packetCounts++
	}

	tmplog.Println(c)
	tmplog.Println(c.IncomingPackets)

	// The connection is completed iff all the packets have arrived
	return packetCounts == c.LastSeqNum
}

func (c *TalkExtConnection) getMessageFromPackets() []byte {
	packets := make([]TalkExtPacket, c.LastSeqNum)
	for i := 0; i < c.LastSeqNum; i++ {
		packets[i] = c.IncomingPackets[i]
	}
	return combinePackets(packets)
}

func (c *TalkExtConnection) generatePackets(msg []byte) []TalkExtPacket {
	n := len(msg)
	step := 500
	size := n / step
	if size * step < n {
		size += 1 // adjusting for remainder
	}
	packets := make([]TalkExtPacket, size)
	for i := 0; i < size; i++ {
		left := i * step
		right := (i+1) * step
		if n < right {
			right = n
		}
		chunk := msg[left:right]
		packet := TalkExtPacket{
			Id: c.Id,
			SeqNum: i,
			LastSeqNum: size,
			Packet: chunk,
		}
		packets[i] = packet
	}
	return packets
}


func AsTalkExtProtocol(proc string) string {
	if strings.ContainsAny(proc, "-") {
		return proc
	} else {
		b := md5.Sum([]byte(proc))
		return proc + "-" + hex.EncodeToString(b[:5])
	}
}

func isTalkExtProtocol(proc string) bool {
	parts := strings.Split(proc, "-")
	inferredProc := AsTalkExtProtocol(parts[0])

	return inferredProc == proc
}

//
//func TalkExtPacketToMessage(packet TalkExtPacket) []byte {
//	ser, err := json.MarshalIndent(packet, "", "  ")
//	if err != nil {
//		tmplog.Fatal(err)
//	}
//	return ser
//}

