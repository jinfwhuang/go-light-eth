package sync

import (
	"github.com/golang/protobuf/proto"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
)

// TODO: use a better filename

func IsEmptyHeader(header *ethpb.BeaconBlockHeader) bool {
	emptyHeader := &ethpb.BeaconBlockHeader{}
	return proto.Equal(header, emptyHeader)
}

func toLightClientUpdate(update *ethpb.SkipSyncUpdate) *ethpb.LightClientUpdate {
	return &ethpb.LightClientUpdate{
		AttestedHeader:          update.AttestedHeader,
		NextSyncCommittee:       update.NextSyncCommittee,
		NextSyncCommitteeBranch: update.NextSyncCommitteeBranch,
		FinalityHeader:          update.FinalityHeader,
		FinalityBranch:          update.FinalityBranch,
		SyncCommitteeBits:       update.SyncCommitteeBits,
		SyncCommitteeSignature:  update.SyncCommitteeSignature,
		ForkVersion:             update.ForkVersion,
	}
}

func getActiveHeader(update *ethpb.LightClientUpdate) *ethpb.BeaconBlockHeader {
	if IsEmptyHeader(update.FinalityHeader) {
		return update.AttestedHeader
	} else {
		return update.FinalityHeader
	}
}

func getSafetyThreshold(store *ethpb.LightClientStore) uint64 {
	max := store.PreviousMaxActiveParticipants
	if store.CurrentMaxActiveParticipants > max {
		max = store.CurrentMaxActiveParticipants
	}
	return max / 2
}
