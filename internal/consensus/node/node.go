package node

import (
	"context"
	lightrpc "github.com/jinfwhuang/go-light-eth/internal/consensus/rpc"
	lightsync "github.com/jinfwhuang/go-light-eth/internal/consensus/sync"
	lightutil "github.com/jinfwhuang/go-light-eth/internal/consensus/util"
	"github.com/prysmaticlabs/prysm/runtime"
	"github.com/prysmaticlabs/prysm/runtime/debug"
	"github.com/urfave/cli/v2"
	tmplog "log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	NetworkingMode = &cli.StringFlag{
		Name:  "networking-mode",  // grpc, json-rpc, libp2p, portal
		Usage: "TODO: xxx",
		Value: "grpc",
	}
	FullNodeServerEndpoint = &cli.StringFlag{
		Name:  "full-node-server-endpoint",
		Usage: "TODO: xxx",
		Value: "localhost:4000",
	}
	SyncModeStr = &cli.StringFlag{
		Name:  "sync-mode",
		Usage: "TODO: xxx",
		Value: "latest",
	}
	DataDir = &cli.StringFlag{
		Name:  "data-dir",
		Usage: "TODO: xxx",
		Value: "../prysm-data/lightnode/",
	}
	GrpcPort = &cli.IntFlag{
		Name:  "grpc-port",
		Usage: "TODO: xxx",
		Value: 4001,
	}
	GrpcGatewayPort = &cli.IntFlag{
		Name:  "grpc-gateway-port",
		Usage: "TODO: xxx",
		Value: 3500,
	}
	GrpcMaxCallRecvMsgSize = &cli.IntFlag{
		Name:  "grpc-max-msg-size",
		Usage: "Integer to define max recieve message call size (default: 4194304 (for 4MB))",
		Value: 1 << 22, // 2^22, ~4.2MB
	}
	GrpcRetryDelay = &cli.IntFlag{
		Name:  "grpc-retry-delay",
		Usage: "In seconds",
		Value: 1, // TODO: 1 is too fast; increase this
	}
	GrpcRetries = &cli.IntFlag{
		Name:  "grpc-retries",
		Usage: "TODO: xxx",
		Value: 3,
	}
	TrustedCurrentCommitteeRoot = &cli.StringFlag{
		Name:  "trusted-current-committee-root",
		Usage: "In base64 string",
		//Value: "UeSv92gwGs+DSk34NqOaCM1DaU9zyclQE6Tc9morK0M=", // roughly 2021-12-02
		//Value: "rcWo3eE6KOLBLDQeahrXkdzxjWnE8qYHmL8HyNWv7b8=" // roughly 2021-12-03
		//Value: "Kmm8RWMsAvIvDOjnuGNqVNKLNf5opImvXyfnL5ebDoo=", // roughly 2021-12-09
		//Value: "susgL7vNce2IqAq6HWBytSZBA8MIDhLmjixQYs6CRug=", // roughly 2021-12-10
		Value: "3c7Wey0TrcWk0FmdSsZPdoQJE1Vho5Ytm71gySCVWfg=", // roughly 2021-12-12
	}
)

var AppFlags = []cli.Flag{
	NetworkingMode,
	FullNodeServerEndpoint,
	SyncModeStr,
	DataDir,
	GrpcPort,
	GrpcGatewayPort,
	GrpcMaxCallRecvMsgSize,
	GrpcRetryDelay,
	GrpcRetries,
	TrustedCurrentCommitteeRoot,
}

type LightNode struct {
	cliCtx   *cli.Context
	ctx      context.Context
	cancel   context.CancelFunc
	services *runtime.ServiceRegistry
	lock     sync.RWMutex
	stop     chan struct{} // Channel to wait for termination notifications.
}

func New(cliCtx *cli.Context) (*LightNode, error) {
	registry := runtime.NewServiceRegistry()

	ctx, cancel := context.WithCancel(cliCtx.Context)
	beacon := &LightNode{
		cliCtx:   cliCtx,
		ctx:      ctx,
		cancel:   cancel,
		services: registry,
		stop:     make(chan struct{}),
	}

	if err := beacon.registerSyncService(); err != nil {
		return nil, err
	}
	if err := beacon.registerRpcService(); err != nil {
		return nil, err
	}

	return beacon, nil
}

func (ln *LightNode) registerSyncService() error {
	// TODO: Link the configuration variables with flags
	syncModeStr := ln.cliCtx.String(SyncModeStr.Name)
	syncMode := lightsync.ModeLatest // Default
	if syncModeStr == lightsync.ModeFinality.String() {
		syncMode = lightsync.ModeFinality
	}
	tmplog.Println(syncMode)

	svc, err := lightsync.NewService(ln.ctx, &lightsync.Config{
		TrustedCurrentCommitteeRoot: ln.cliCtx.String(TrustedCurrentCommitteeRoot.Name),
		SyncMode:                    syncMode,
		DataDir:                     ln.cliCtx.String(DataDir.Name),
		FullNodeServerEndpoint:      ln.cliCtx.String(FullNodeServerEndpoint.Name),
		GrpcRetryDelay:              time.Second * time.Duration(ln.cliCtx.Int(GrpcRetryDelay.Name)),
		GrpcRetries:                 ln.cliCtx.Int(GrpcRetries.Name),
		GrpcMaxCallRecvMsgSize:      ln.cliCtx.Int(GrpcMaxCallRecvMsgSize.Name),
	})
	if err != nil {
		panic(err)
	}

	return ln.services.RegisterService(svc)
}

func (ln *LightNode) registerRpcService() error {
	var lightsyncService *lightsync.Service
	if err := ln.services.FetchService(&lightsyncService); err != nil {
		panic(err)
	}

	svc, err := lightrpc.NewService(ln.ctx, &lightrpc.Config{
		GrpcPort:        ln.cliCtx.Int(GrpcPort.Name),
		GrpcGatewayPort: ln.cliCtx.Int(GrpcGatewayPort.Name),
		SyncService:     lightsyncService,
	})
	if err != nil {
		panic(err)
	}

	return ln.services.RegisterService(svc)
}

// Start the BeaconNode and kicks off every registered service.
func (ln *LightNode) Start() {
	ln.lock.Lock()
	ln.services.StartAll()

	stop := ln.stop
	ln.lock.Unlock()

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		lightutil.Log.Info("Got interrupt, shutting down...")
		debug.Exit(ln.cliCtx) // Ensure trace and CPU profile data are flushed.
		go ln.Close()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				lightutil.Log.WithField("times", i-1).Info("Already shutting down, interrupt more to panic")
			}
		}
		panic("Panic closing the beacon node")
	}()

	// Wait for stop channel to be closed.
	<-stop
}

// Close handles graceful shutdown of the system.
func (ln *LightNode) Close() {
	ln.lock.Lock()
	defer ln.lock.Unlock()

	lightutil.Log.Info("Stopping beacon node")
	ln.services.StopAll()
	//if err := b.db.Close(); err != nil {
	//	log.Errorf("Failed to close database: %v", err)
	//}
	//b.collector.unregister()
	ln.cancel()
	close(ln.stop)
}
