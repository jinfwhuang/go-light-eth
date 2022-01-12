package sync

import (
	"encoding/base64"
	"errors"
	"fmt"
	eth2_types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/signing"
	"github.com/prysmaticlabs/prysm/config/params"
	"github.com/prysmaticlabs/prysm/crypto/bls/blst"
	"github.com/prysmaticlabs/prysm/crypto/bls/common"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/time/slots"
	tmplog "log"
)

const GenesisValidatorsRootBase64Str = "SzY9uU4oYSDXbrkFNA/dTlS/6fBr8z/2z1rSf1Eb/pU="

var GenesisValidatorsRoot [32]byte

func init() {
	b, err := base64.StdEncoding.DecodeString(GenesisValidatorsRootBase64Str)
	if err != nil {
		panic(err)
	}
	GenesisValidatorsRoot = bytesutil.ToBytes32(b)
}

func verifyMerkleFinalityHeader(update *ethpb.LightClientUpdate) bool {
	if IsEmptyHeader(update.FinalityHeader) { // non-finzlied update
		// TODO: check update.FinalityBranch integrity
		return true
	} else {
		gIndex := ztype.CalculateGIndex(ztype.BeaconStateAltairType, 20, 1)
		root := update.AttestedHeader.StateRoot
		leaf, err := update.FinalityHeader.HashTreeRoot()
		if err != nil {
			tmplog.Println(err)
			panic(err)
		}
		branch := update.FinalityBranch
		return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
	}
}

func verifyMerkleNextSyncComm(update *ethpb.LightClientUpdate) bool {
	gIndex := ztype.CalculateGIndex(ztype.BeaconStateAltairType, 23) // next_sync_committee  23
	leaf, err := update.NextSyncCommittee.HashTreeRoot()
	if err != nil {
		tmplog.Println(err)
		panic(err)
	}
	branch := update.NextSyncCommitteeBranch

	if IsEmptyHeader(update.FinalityHeader) { // non-finzlied update
		root := update.AttestedHeader.StateRoot
		return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
	} else {
		root := update.FinalityHeader.StateRoot
		return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
	}
}

func verifyMerkleCurrentSyncComm(update *ethpb.SkipSyncUpdate) bool {
	gIndex := ztype.CalculateGIndex(ztype.BeaconStateAltairType, 22) // current_sync_committee  22
	leaf, err := update.CurrentSyncCommittee.HashTreeRoot()
	if err != nil {
		tmplog.Println(err)
		panic(err)
	}
	branch := update.CurrentSyncCommitteeBranch

	if IsEmptyHeader(update.FinalityHeader) { // non-finzalied update
		root := update.AttestedHeader.StateRoot
		return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
	} else { // finalized update
		root := update.FinalityHeader.StateRoot
		return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
	}
}

func validateMerkleSkipSyncUpdate(update *ethpb.SkipSyncUpdate) error {
	verified := verifyMerkleFinalityHeader(toLightClientUpdate(update))
	verified = verified && verifyMerkleCurrentSyncComm(update)
	verified = verified && verifyMerkleNextSyncComm(toLightClientUpdate(update))
	if !verified {
		return errors.New("update has invalid merkle proof")
	}
	return nil
}

func validateMerkleLightClientUpdate(update *ethpb.LightClientUpdate) error {
	verified := verifyMerkleFinalityHeader(update)
	verified = verified && verifyMerkleNextSyncComm(update)
	if !verified {
		return errors.New("update has invalid merkle proof")
	}
	return nil
}

func validateLightClientUpdate(
	store *ethpb.LightClientStore,
	update *ethpb.LightClientUpdate,
	currentSlot eth2_types.Slot,
	genesisValidatorsRoot [32]byte,
) error {
	// Internal consistency
	err := validateMerkleLightClientUpdate(update)
	if err != nil {
		return err
	}

	// Verify update slot is larger than slot of current best finalized header
	activeHeader := getActiveHeader(update)
	if currentSlot < activeHeader.Slot || activeHeader.Slot < store.FinalizedHeader.Slot {
		return fmt.Errorf("update slot is invalid")
	}

	// Verify update does not skip a sync committee period
	finalizedPeriod := slots.ToEpoch(store.FinalizedHeader.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
	updatePeriod := slots.ToEpoch(activeHeader.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
	if updatePeriod != finalizedPeriod && updatePeriod != finalizedPeriod+1 {
		// TODO: go to skip-sync mode instead
		return fmt.Errorf("update period is invalid")
	}

	// Verify sync committee has sufficient participants
	if update.SyncCommitteeBits.Count() < params.BeaconConfig().MinSyncCommitteeParticipants {
		return errors.New("insufficient participants")
	}

	// # Verify sync committee aggregate signature
	syncCommittee := store.CurrentSyncCommittee
	if updatePeriod == finalizedPeriod+1 {
		syncCommittee = store.NextSyncCommittee
	}
	err = verifySign(syncCommittee, update, genesisValidatorsRoot)
	if err != nil {
		return err
	}

	return nil
}

// TODO: better name
func verifySign(
	syncCommittee *ethpb.SyncCommittee,
	update *ethpb.LightClientUpdate,
	genesisValidatorsRoot [32]byte,
) error {
	activeHeader := getActiveHeader(update)
	participantPubkeys := make([][]byte, 0)
	for i, pubKey := range syncCommittee.Pubkeys {
		bit := update.SyncCommitteeBits.BitAt(uint64(i))
		if bit {
			participantPubkeys = append(participantPubkeys, pubKey)
		}
	}
	domain, err := signing.ComputeDomain(
		params.BeaconConfig().DomainSyncCommittee,
		update.ForkVersion,
		genesisValidatorsRoot[:],
	)
	if err != nil {
		return err
	}
	signingRoot, err := signing.ComputeSigningRoot(activeHeader, domain)
	if err != nil {
		return err
	}
	sig, err := blst.SignatureFromBytes(update.SyncCommitteeSignature[:])
	if err != nil {
		return err
	}
	pubKeys := make([]common.PublicKey, 0)
	for _, pubkey := range participantPubkeys {
		pk, err := blst.PublicKeyFromBytes(pubkey)
		if err != nil {
			return err
		}
		pubKeys = append(pubKeys, pk)
	}
	if !sig.FastAggregateVerify(pubKeys, signingRoot) {
		return errors.New("failed to verify")
	}
	return nil

}
