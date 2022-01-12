package main

import (
	"context"
	"encoding/base64"
	"github.com/golang/protobuf/ptypes/empty"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"google.golang.org/grpc"
	tmplog "log"
)

func main() {
	ctx := context.Background()
	dialOpts := []grpc.DialOption{
		//grpc.WithDefaultCallOptions(
		//	grpc.MaxCallRecvMsgSize(s.cfg.GrpcMaxCallRecvMsgSize),
		//	grpc_retry.WithMax(uint(s.cfg.GrpcRetries)),
		//	grpc_retry.WithBackoff(grpc_retry.BackoffLinear(s.cfg.GrpcRetryDelay)),
		//),
		//grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		//grpc.WithUnaryInterceptor(middleware.ChainUnaryClient(
		//	grpc_opentracing.UnaryClientInterceptor(),
		//	grpc_prometheus.UnaryClientInterceptor,
		//	grpc_retry.UnaryClientInterceptor(),
		//	grpcutil.LogRequests,
		//)),
		//grpc.WithChainStreamInterceptor(
		//	grpcutil.LogStream,
		//	grpc_opentracing.StreamClientInterceptor(),
		//	grpc_prometheus.StreamClientInterceptor,
		//	grpc_retry.StreamClientInterceptor(),
		//),
		grpc.WithInsecure(),
		//grpc.WithResolvers(&multipleEndpointsGrpcResolverBuilder{}),
	}
	//v.ctx = grpcutil.AppendHeaders(s.ctx, s.grpcHeaders)

	addr := "localhost:4000"
	tmplog.Printf("connecting to grpc server: %s", addr)
	conn, err := grpc.DialContext(ctx, "localhost:4000", dialOpts...)
	if err != nil {
		panic(err)
	}
	server := ethpb.NewLightClientClient(conn)

	resp, err := server.DebugGetTrustedCurrentCommitteeRoot(ctx, &empty.Empty{})
	if err != nil {
		panic(err)
	}
	key := base64.StdEncoding.EncodeToString(resp.Key)
	tmplog.Printf("Here is a recent started point from addrss: %s", addr)
	tmplog.Printf("--trusted-current-committee-root='%s'", key)
}
