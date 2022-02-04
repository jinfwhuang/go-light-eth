package discover

import (
	crand "crypto/rand"
	"encoding/json"
	"gotest.tools/assert"
	tmplog "log"
	"math/rand"
	"testing"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func randomBytes(n int) []byte {
	token := make([]byte, n)
	crand.Read(token)
	return token
}



func Test_TalkRequestExtDerser(t *testing.T) {
	packet := TalkExtPacket{
		Id:         ConnectionId(rand.Uint64()),
		SeqNum:     0,
		LastSeqNum: 10,
		Packet:     []byte("jinfwhuang"),
	}
	tmplog.Println(packet)
	tmplog.Println(string(packet.Packet))

	ser, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		tmplog.Fatal(err)
	}
	tmplog.Println(string(ser))

	// Deser
	packet2 := TalkExtPacket{}
	err = json.Unmarshal(ser, &packet2)

	tmplog.Println(packet2)
	tmplog.Println(string(packet2.Packet))

}

func Test_toTalkExtProtocol(t *testing.T) {
	procStr := "jinfwhuang"
	talkExtProcStr := toTalkExtProtocol(procStr)

	tmplog.Println(talkExtProcStr)
	tmplog.Println(isTalkExtProtocol(talkExtProcStr))

}

func Test_splitMessage(t *testing.T) {
	data := randomBytes(1090)

	packets := splitMessage(0, data)

	tmplog.Println(len(packets))

	combinedData := combinePackets(packets)

	tmplog.Println(data)
	tmplog.Println(combinedData)

	//assert.Equal(t, data, combinedData)
	assert.DeepEqual(t, data, combinedData)

	//tmplog.Println(packets[0])
	//tmplog.Println(packets)

}
//splitMessage