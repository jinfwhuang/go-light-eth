package sync

import (
	eth2_types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/prysm/config/params"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/time/slots"
)

func processLightClientUpdate(
	store *ethpb.LightClientStore,
	update *ethpb.LightClientUpdate,
	currentSlot eth2_types.Slot,
	genesisValidatorsRoot [32]byte,
) error {
	validateLightClientUpdate(store, update, currentSlot, genesisValidatorsRoot)

	// Updates
	//-----------
	// Update the best update in case we have to force-update to it if the timeout elapses
	if store.BestValidUpdate == nil ||
		update.SyncCommitteeBits.Count() > store.BestValidUpdate.SyncCommitteeBits.Count() {
		store.BestValidUpdate = update
	}

	// Update max number of active participants
	if store.CurrentMaxActiveParticipants < update.SyncCommitteeBits.Count() {
		store.CurrentMaxActiveParticipants = update.SyncCommitteeBits.Count()
	}

	// Update the optimistic header
	if update.SyncCommitteeBits.Count() > getSafetyThreshold(store) &&
		update.AttestedHeader.Slot > store.OptimisticHeader.Slot {
		store.OptimisticHeader = update.AttestedHeader
	}

	// Update finalized header
	if !IsEmptyHeader(update.FinalityHeader) &&
		update.SyncCommitteeBits.Count() >= update.SyncCommitteeBits.Len()*3.0/2.0 {
		// Normal update: 2/3 threshold and apply finality_header
		applyLightClientUpdate(store, update)
	} else if currentSlot > store.FinalizedHeader.Slot+UpdateTimeout &&
		store.BestValidUpdate != nil {
		// TODO: go skip-sync mode instead?
		// Last ditch effort update: use best_valid_update
		applyLightClientUpdate(store, store.BestValidUpdate)
		store.BestValidUpdate = nil
	}

	// TODO: What is going on in this method is hard to read. Think of a way to refactor...

	return nil
}

func applyLightClientUpdate(store *ethpb.LightClientStore, update *ethpb.LightClientUpdate) {
	activeHeader := getActiveHeader(update)
	finalizedPeriod := slots.ToEpoch(store.FinalizedHeader.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
	updatePeriod := slots.ToEpoch(activeHeader.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
	if updatePeriod == finalizedPeriod+1 {
		store.CurrentSyncCommittee = store.NextSyncCommittee
	}
	store.FinalizedHeader = activeHeader
}
