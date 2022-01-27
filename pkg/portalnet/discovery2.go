package portalnet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"os"

	//"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"

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
// Listener defines the discovery V5 network interface that is used
// to communicate with other peers.
type Listener interface {
	Self() *enode.Node
	Close()
	Lookup(enode.ID) []*enode.Node
	Resolve(*enode.Node) *enode.Node
	RandomNodes() enode.Iterator
	Ping(*enode.Node) error
	RequestENR(*enode.Node) (*enode.Node, error)
	LocalNode() *enode.LocalNode
}

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
	tmplog.Println("node2")


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
	socket, err := net.ListenUDP("udp4", addr)
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
	tmplog.Printf("upd on %s:%d", ln.Node().IP(), ln.Node().UDP())
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

	//tmplog.Fatal("fff")

	// register a TalkRequest handler

	udp.RegisterTalkHandler("bbb", func (node enode.ID, addr *net.UDPAddr, input []byte) []byte {
		tmplog.Println("request coming from", addr)
		tmplog.Println("responding in", udp.LocalNode().Node().IP(), udp.LocalNode().Node().UDP())

		return append(input, []byte("output from " + nodename)...)
	})


	return udp
	//return nil
}

func bbbHandler(node enode.ID, addr *net.UDPAddr, input []byte) []byte {
	tmplog.Println(node, addr)
	tmplog.Println(string(input))

	return append(input, []byte("output")...)
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

func (s *Disv5Service) Start(nodename string) {

	udpv5 := startLocalhostV5(nodename)

	//target := random32Byte()
	//tmplog.Println(target)

	// event loop
	for {
		time.Sleep(time.Second * 5)
		tmplog.Printf("listening with %v:%v", udpv5.Self().IP(), udpv5.Self().UDP())
		tmplog.Printf("self %s", udpv5.LocalNode().Node())

		//udpv5.Lookup(target)
		nodes := udpv5.AllNodes()
		tmplog.Println("peers", len(nodes))


		for _, n := range nodes {
			tmplog.Println(time.Now())
			//tmplog.Println("pinging", n)
			//err := udpv5.Ping(n)
			//if err != nil {
			//	tmplog.Println(err)
			//}

			response, err := udpv5.TalkRequest(n, "bbb", []byte(nodename + " is making a request. "))
			tmplog.Println("talking talkrequest", err)
			tmplog.Println(string(response))

			//udpv5.LocalNode().SetStaticIP(realaddr.IP)
			//udpv5.LocalNode().Set

			tmplog.Printf("listening with %v:%v", udpv5.Self().IP(), udpv5.Self().UDP())
		}

		// Pull the disv5 codebase, expose the routing table
	}
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