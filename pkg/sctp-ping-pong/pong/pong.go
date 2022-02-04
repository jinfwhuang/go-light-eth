package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/pion/logging"
	"github.com/pion/sctp"
)

func init() {
	log.SetFlags(log.Llongfile)
}

func main() {
	addr := net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 5678,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
	fmt.Println("created a udp listener")

	config := sctp.Config{
		NetConn:       &disconnectedPacketConn{pConn: conn},
		LoggerFactory: logging.NewDefaultLoggerFactory(),
	}
	a, err := sctp.Server(config)
	if err != nil {
		log.Panic(err)
	}
	defer a.Close()
	fmt.Println("created a server")

	stream, err := a.AcceptStream()
	if err != nil {
		log.Panic(err)
	}
	defer stream.Close()
	fmt.Println("accepted a stream")

	// set unordered = true and 10ms treshold for dropping packets
	stream.SetReliabilityParams(true, sctp.ReliabilityTypeTimed, 10)
	var pongSeqNum int
	for {
		buff := make([]byte, 10000)
		_, err = stream.Read(buff)
		if err != nil {
			log.Panic(err)
		}
		log.Println("received", len(buff), cap(buff), buff[:13])

		pongMsg := fmt.Sprintf("pong %d", pongSeqNum)
		_, err = stream.Write([]byte(pongMsg))
		if err != nil {
			log.Panic(err)
		}
		log.Println("sent:", pongMsg)

		pongSeqNum++

		time.Sleep(time.Second)
	}
}
