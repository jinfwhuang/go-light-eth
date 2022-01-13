package main

import (
	"github.com/prysmaticlabs/go-bitfield"
	lightnode "github.com/prysmaticlabs/prysm/cmd/light-client/node"

	v1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	v2 "github.com/prysmaticlabs/prysm/proto/eth/v2"
	v1alpha1 "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	tmplog "log"
	"os"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

// Precomputed values for generalized indices.
const (
	FinalizedRootIndex              = 105
	FinalizedRootIndexFloorLog2     = 6
	NextSyncCommitteeIndex          = 55
	NextSyncCommitteeIndexFloorLog2 = 5
)

var log = logrus.WithField("prefix", "light")

type LightClientSnapshot struct {
	Header               *v1.BeaconBlockHeader
	CurrentSyncCommittee *v2.SyncCommittee
	NextSyncCommittee    *v2.SyncCommittee
}

type LightClientUpdate struct {
	Header                  *v1.BeaconBlockHeader
	NextSyncCommittee       *v2.SyncCommittee
	NextSyncCommitteeBranch [NextSyncCommitteeIndexFloorLog2][32]byte
	FinalityHeader          *v1.BeaconBlockHeader
	FinalityBranch          [FinalizedRootIndexFloorLog2][32]byte
	SyncCommitteeBits       bitfield.Bitvector512
	SyncCommitteeSignature  [96]byte
	ForkVersion             *v1alpha1.Version
}

func main() {
	app := cli.App{}
	app.Name = "beacon-chain-light-client"
	app.Usage = "Beacon Chain Light Client"
	app.Action = start

	app.Flags = lightnode.AppFlags

	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
	}
}

func start(ctx *cli.Context) error {
	// Fix data dir for Windows users.
	//outdatedDataDir := filepath.Join(file.HomeDir(), "AppData", "Roaming", "Eth2")
	//serverEndpoint := ctx.String(lightnode..Name)
	//
	//tmplog.Println(serverEndpoint)
	//
	//tmplog.Println("staring light client")
	//count := 0
	//for {
	//	tmplog.Println("counting", count)
	//	count += 1
	//	time.Sleep(time.Second * 10)
	//}
	lightNode, err := lightnode.New(ctx)
	if err != nil {
		panic(err)
	}
	lightNode.Start()
	return nil
}
