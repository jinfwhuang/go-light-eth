package rpc

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	lightsync "github.com/jinfwhuang/go-light-eth/internal/consensus/sync"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	tmplog "log"
)

type Server struct {
	syncService *lightsync.Service
}

func (s *Server) debug() {
	store := s.syncService.Store

	tmplog.Println("--------------------------")
	tmplog.Println(lightsync.StoreToString(store))
	tmplog.Println("--------------------------")
}

func (s *Server) Head(ctx context.Context, _ *empty.Empty) (*ethpb.BeaconBlockHeader, error) {
	//s.debug()
	return s.syncService.Store.OptimisticHeader, nil
}

func (s *Server) FinalizedHead(ctx context.Context, _ *empty.Empty) (*ethpb.BeaconBlockHeader, error) {
	return s.syncService.Store.FinalizedHeader, nil
}
