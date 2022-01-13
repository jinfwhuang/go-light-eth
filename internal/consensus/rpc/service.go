package rpc

import (
	"context"
	"fmt"
	lightsync "github.com/jinfwhuang/go-light-eth/internal/consensus/sync"
	ethpbv1alpha1 "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
)

// headSyncMinEpochsAfterCheckpoint defines how many epochs should elapse after known finalization
// checkpoint for head sync to be triggered.
const headSyncMinEpochsAfterCheckpoint = 128

type Service struct {
	cfg *Config
	ctx context.Context

	cancel               context.CancelFunc
	listener             net.Listener
	grpcServer           *grpc.Server
	canonicalStateChan   chan *ethpbv1alpha1.BeaconState
	incomingAttestation  chan *ethpbv1alpha1.Attestation
	credentialError      error
	connectedRPCClients  map[net.Addr]bool
	clientConnectionLock sync.Mutex
}

type Config struct {
	GrpcPort        int
	GrpcGatewayPort int // TODO: add http json gateway
	SyncService     *lightsync.Service
	//GrpcMaxCallRecvMsgSize int
	//GrpcRetryDelay         time.Duration
	//GrpcRetries            int
}

func NewService(ctx context.Context, cfg *Config) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	svr := &Service{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
	}
	return svr, nil
}

func (s *Service) Stop() error {
	defer s.cancel()
	return nil
}

func (s *Service) Status() error {
	return nil
}

// Start the gRPC server.
func (s *Service) Start() {
	address := fmt.Sprintf("%s:%d", "127.0.0.1", s.cfg.GrpcPort)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Errorf("could not listen to port in Start() %s: %v", address, err))
	}
	s.listener = lis
	log.WithField("address", address).Info("gRPC server listening on port")

	// GRPC Server
	s.grpcServer = grpc.NewServer([]grpc.ServerOption{}...)

	// Register endpoints
	ethpbv1alpha1.RegisterLightNodeServer(s.grpcServer, &Server{
		syncService: s.cfg.SyncService,
	})
	reflection.Register(s.grpcServer)

	go func() {
		if s.listener != nil {
			if err := s.grpcServer.Serve(s.listener); err != nil {
				panic(fmt.Errorf("could not serve gRPC: %v", err))
			}
		}
	}()
}
