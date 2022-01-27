package portalnet

import (
	"fmt"
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

