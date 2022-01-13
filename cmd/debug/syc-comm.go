package debug

import (
	"context"
	"encoding/base64"
	"github.com/golang/protobuf/ptypes/empty"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"google.golang.org/grpc"
	tmplog "log"
)

func GetTrustedCurrentCommitteeRoot() string {
	ctx := context.Background()
	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	addr := "localhost:4000"
	tmplog.Printf("connecting to grpc server: %s", addr)
	conn, err := grpc.DialContext(ctx, "localhost:4000", dialOpts...)
	if err != nil {
		panic(err)
	}
	updateClient := ethpb.NewLightUpdateClient(conn)

	resp, err := updateClient.DebugGetTrustedCurrentCommitteeRoot(ctx, &empty.Empty{})
	if err != nil {
		panic(err)
	}
	key := base64.StdEncoding.EncodeToString(resp.Key)
	tmplog.Printf("Here is a recent started point from addrss: %s", addr)
	tmplog.Printf("--trusted-current-committee-root='%s'", key)
	return key
}
