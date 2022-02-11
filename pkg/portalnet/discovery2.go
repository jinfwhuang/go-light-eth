package portalnet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"os"
	//"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"

	//"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/jinfwhuang/go-light-eth/pkg/discover"

	tmplog "log"
	"net"
	"time"
)

const (
	KeyLoc = "/Users/jin/code/repos/go-light-eth/pkg/portalnet/tmp/bootnode.privatekey"
	NodePath = "/Users/jin/code/repos/go-light-eth/pkg/portalnet/tmp/nodedb"

	//FixedEnr = "enr:-Iq4QB1Q4x83fxo_A4HKY7aLD_E9s0-DA1IKbv1tIGP2E1eOOLW4FS5xCWcLw5AcQB5cl-mVVFTRiU2-GuSFUYTTVfqGAX6XThQigmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQMqc-jmk1aq9OQJv2tquwGiLjZHLijJrAxqF6irWV8eoYN1ZHCC5dA"
	//FixedEnr =
)


type NodeEnv struct {
	name string
	enr string
	keypath string
	dbpath string
	port int
}

var (
	node1Env = NodeEnv {
		name: "node1",
		enr: "enr:-Iq4QBMYaIDDNPd-DlPx2Odcu3ihilFx30EfH8U2e3eo1xJqDoYHDJDPJ6pvDz5b74ck-9HqixoxtrejzE_ngtjnkIaGAX6ZrfyGgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQMqc-jmk1aq9OQJv2tquwGiLjZHLijJrAxqF6irWV8eoYN1ZHCC5c8",
		keypath: "/Users/jin/code/repos/go-light-eth/pkg/portalnet/tmp/nodedb/node1/privatekey",
		dbpath: "/Users/jin/code/repos/go-light-eth/pkg/portalnet/tmp/nodedb/node1/db",
		port: 58831,
	}
	node2Env = NodeEnv {
		name: "node2",
		enr: "enr:-Iq4QO_eT32DiimQzMaS0Qvm3RNfQO8G5bGK66yNsqq52XiJTjjOXC37-oLD0tLCJ_uM-jeqjaou0IS1_TIF03TCHvSGAX6ZoxUggmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQK_UhfbboKeTzmxJ-eOLP6N21cbpDFUZZZhpHbmAveB-IN1ZHCC5dA",
		keypath: "/Users/jin/code/repos/go-light-eth/pkg/portalnet/tmp/nodedb/node2/privatekey",
		dbpath: "/Users/jin/code/repos/go-light-eth/pkg/portalnet/tmp/nodedb/node2/db",
		port: 58832,
	}

	NodeName3 = "node3"
)


const (
	TalkProc = "talkprotocol"
)

func Aaa() {
	tmplog.Println("fff")
}

type Disv5Service struct {

}

func saveKey(key *ecdsa.PrivateKey, keypath string) {
	data := crypto.FromECDSA(key)
	err := os.WriteFile(keypath, data, 0666)
	if err != nil {
		tmplog.Fatal(err)
	}

	fileData, err := os.ReadFile(keypath)
	//fileKey := &ecdsa.PrivateKey{}
	fileKey, err := crypto.ToECDSA(fileData)
	if err != nil {
		tmplog.Fatal(err)
	}

	tmplog.Println(key)
	tmplog.Println(fileKey)

}

func newkey() *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("couldn't generate key: " + err.Error())
	}


	return key
}

func startLocalhostV5(nodename string) *discover.UDPv5 {
	// key
	// address
	// db path
	cfg := discover.Config{
		Bootnodes:  []*enode.Node {
			fromEnr(node1Env.enr),
			fromEnr(node2Env.enr),
		},
	}
	cfg.PrivateKey = newkey()

	//saveKey(cfg.PrivateKey, node2Env.keypath)
	//tmplog.Fatal("fff")

	addr := &net.UDPAddr{
		IP: net.IP{127, 0, 0, 1},
	}
	db, err := enode.OpenDB("")
	if err != nil {
		tmplog.Fatal(err)
	}
	tmplog.Println(addr)

	// fixed node
	if nodename == "node1" || nodename == "node2" {
		keypath := node1Env.keypath
		dbpath := node1Env.dbpath
		port := node1Env.port
		if nodename == "node2" {
			keypath = node2Env.keypath
			dbpath = node2Env.dbpath
			port = node2Env.port
		}

		// key, port, db
		fileData, err := os.ReadFile(keypath)
		fileKey, err := crypto.ToECDSA(fileData)
		if err != nil {
			tmplog.Fatal(err)
		}
		cfg.PrivateKey = fileKey
		addr = &net.UDPAddr{
			IP: net.IP{127, 0, 0, 1},
			Port: port,
		}
		db, err = enode.OpenDB(dbpath)
		if err != nil {
			tmplog.Fatal(err)
		}
	}

	ln := enode.NewLocalNode(db, cfg.PrivateKey)

	// Listen
	//conn, err := net.ListenUDP(networkVersion, udpAddr)

	//socket, err := net.ListenUDP("udp4", addr)
	socket, err := net.ListenUDP("udp", addr)
	if err != nil {
		tmplog.Fatal(err)
	}
	realaddr := socket.LocalAddr().(*net.UDPAddr)
	ln.SetStaticIP(realaddr.IP)
	ln.SetFallbackIP(realaddr.IP)
	ln.Set(enr.UDP(realaddr.Port))
	ln.SetFallbackUDP(realaddr.Port)
	udp, err := discover.ListenV5(socket, ln, cfg)
	if err != nil {
		tmplog.Fatal(err)
	}
	tmplog.Printf("udp on %s:%d", ln.Node().IP(), ln.Node().UDP())
	tmplog.Printf("current: %s", ln.Node())

	db.UpdateNode(ln.Node())

	// ENR validation
	if nodename == "node2" {
		fixedNode := fromEnr(node2Env.enr)
		tmplog.Printf("fixed: %s", fixedNode.String())
		tmplog.Println(fixedNode.Seq())

		_enr := toEnr(ln.Node())
		tmplog.Printf("got: %s", _enr)
		tmplog.Println(ln.Seq())
	}


	return udp
	//return nil
}

func (s *Disv5Service) Start(nodename string) {
	udpv5 := startLocalhostV5(nodename)

	// TALKREQUEST HANDLER
	udpv5.RegisterTalkHandler(TalkProc, func (node enode.ID, addr *net.UDPAddr, input []byte) []byte {
		tmplog.Println("invoking TALKREQ handler", TalkProc,  "data size", len(input))
		return input
	})

	// TALKEXT handler
	udpv5.RegisterTalkExtHandler(TalkProc, func (node enode.ID, addr *net.UDPAddr, input []byte) []byte {
		tmplog.Println("invoking TALKEXT handler", discover.AsTalkExtProtocol(TalkProc),  "data size", len(input))
		return input
	})


	//target := random32Byte()
	//tmplog.Println(target)

	// event loop
	for {
		nodes := udpv5.AllNodes()
		//tmplog.Println("how many peers connected:", len(nodes))
		//tmplog.Printf("listening with %v:%v", udpv5.Self().IP(), udpv5.Self().UDP())

		tmplog.Println("peers", len(nodes))
		for _, n := range nodes {
			tmplog.Println("---", n.IP(), n.UDP(),  time.Now(), "---")
		}

		for _, n := range nodes {
			if nodename == NodeName3 && n.UDP() == node1Env.port {

				//// small TALKREQUEST
				//response, err := udpv5.TalkRequest(n, TalkProc, []byte(nodename + " made a request. | "))
				//if err != nil {
				//	tmplog.Println(err)
				//}
				//tmplog.Println("Just made a talkrequest", string(response))


				// small TALKEXT
				//response, err := udpv5.TalkRequestExt(n, TalkProc, []byte(nodename + " made a request. | "))
				response, err := udpv5.TalkRequestExt(n, TalkProc, []byte("jin"))
				if err != nil {
					tmplog.Println(err)
				}
				tmplog.Println("Just made a talkext", string(response))


					// Large quest should fail
					//body := _requestBody(1281)
					//tmplog.Println("About to make a large talkrequest", len(body), cap(body), body[:13])
					////tmplog.Println(len(body), "utf validitiy", utf8.Valid(body[:50]))
					//resp2, err := udpv5.TalkRequest(n, TalkProc, body)
					//if err != nil {
					//	tmplog.Println("Completed a large talkrequest", len(body), cap(body), err, resp2)
					//} else {
					//	tmplog.Println("Completed a large talkrequest", len(body), cap(body), err, resp2[:13])
					//}

					//// Talk ext
					//body := _requestBody(100)
					//tmplog.Println("About to make talk-ext", len(body), cap(body), body[:13])
					//resp2, err := TalkRequestExt(udpv5, n, TalkProc, body)
					//if err != nil {
					//	tmplog.Println("Completed a large talkrequest", len(body), cap(body), err, resp2)
					//} else {
					//	tmplog.Println("Completed a large talkrequest", len(body), cap(body), err, resp2[:13])
					//}
			}

		}

		//if nodename == NodeName3 {
		//	tmplog.Fatal("die...")
		//}

		time.Sleep(time.Second * 3)
	}
}

/*
UDP has a maximum packet size of 1280 bytes.
 */
func _requestBody(n int) []byte {
	token := make([]byte, n)
	rand.Read(token)
	return token
}



func toEnr(n *enode.Node) string {
	return n.String()
}

func fromEnr(addr string) *enode.Node {
	node, err := enode.Parse(enode.ValidSchemes, addr)
	if err != nil {
		tmplog.Fatal(err)
	}
	return node
}


