package portalnet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	tmplog "log"
	"net"
	"time"
)

func isUdpPortOpen(port int) bool {
	timeout := time.Second * 2
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("udp", addr, timeout)
	if err != nil {
		tmplog.Fatal(err)
	}
	tmplog.Println(conn)
	n, err := conn.Write([]byte{8, 8, 8})
	tmplog.Println(n, err)

	tmplog.Fatal("ggg")
	return true
}

func random32Byte() [32]byte {
	token := [32]byte{}
	rand.Read(token[:])
	return token
}


func createLocalNode(
	privKey *ecdsa.PrivateKey,
	ipAddr net.IP,
	udpPort int) *enode.LocalNode {
	db, err := enode.OpenDB("")
	if err != nil {
		tmplog.Fatal(err)
	}
	localNode := enode.NewLocalNode(db, privKey)

	ipEntry := enr.IP(ipAddr)
	udpEntry := enr.UDP(udpPort)
	localNode.Set(ipEntry)
	localNode.Set(udpEntry)

	return localNode
}
