package ztype

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/io/file"
	"github.com/prysmaticlabs/prysm/testing/require"
	tmplog "log"
	"math/rand"
	"testing"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func hex0x(b []byte) string {
	return "0x" + hex.EncodeToString(b[:])
}

func hashBytes(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func TestAA(t *testing.T) {
	tmplog.Println("fff")
}

func Test_HashTreeRoot(t *testing.T) {
	ctx := context.Background()
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	// Calculate root through the ztyp library
	root1 := state.HashTreeRoot()

	// Calculate root through the algorithm in: beacon-chain/state/v2/state_trie.go
	root2, err := state.State.HashTreeRoot(ctx)
	require.NoError(t, err)

	// Validation
	require.Equal(t, hex.EncodeToString(root1[:]), hex.EncodeToString(root2[:]))

	tmplog.Println(hex.EncodeToString(root1[:]))
	tmplog.Println(hex.EncodeToString(root2[:]))
	tmplog.Println("----")
	tmplog.Println(root1[:])
	tmplog.Println(root2[:])
	tmplog.Println("----")

}

func Test_SszSerialize(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	// Through ztyp library
	sszbytes1 := state.SszSerialize()

	// Through fastssz library
	sszbytes2, err := state.State.MarshalSSZ()
	require.NoError(t, err)

	n := rand.Intn(len(sszbytes1)) // Get some random bytes
	tmplog.Println(n)
	tmplog.Println(sszbytes1[n : n+17])
	tmplog.Println(sszbytes2[n : n+17])

	tmplog.Println(hashBytes(sszbytes1))
	tmplog.Println(hashBytes(sszbytes2))
}

func Test_Verify(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	// {"next_sync_committee", SyncCommitteeType}, // 23
	root := state.HashTreeRoot()
	gIndex := state.GetGIndex(23)
	leaf := state.GetLeaf(gIndex)
	branch := state.GetBranch(gIndex)

	okay := Verify(root, gIndex, leaf, branch)

	tmplog.Println(root)
	tmplog.Println(gIndex)
	tmplog.Println(leaf)
	tmplog.Println(branch)
	tmplog.Println(okay)
}

func Test_Proof(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	// {"next_sync_committee", SyncCommitteeType}, // 23
	root := state.HashTreeRoot()
	gIndex := state.GetGIndex(23)
	leaf := state.GetLeaf(gIndex)
	branch := state.GetBranch(gIndex)

	require.Equal(t, Verify(root, gIndex, leaf, branch), true)
}

func Test_FinalityCheckpointBranch(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	root := state.HashTreeRoot()
	gIndex := state.GetGIndex(20) // finalized_checkpoint 20
	leaf := state.GetLeaf(gIndex)
	branch := state.GetBranch(gIndex)

	require.Equal(t, Verify(root, gIndex, leaf, branch), true)
}

func Test_FinalityHeaderBranch(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	tmplog.Println(base64.StdEncoding.EncodeToString(state.State.GenesisValidatorRoot()))

	root := state.HashTreeRoot()
	gIndex := state.GetGIndex(20, 1) // finalized_checkpoint (20), root (1)
	leaf := state.GetLeaf(gIndex)
	branch := state.GetBranch(gIndex)

	require.Equal(t, Verify(root, gIndex, leaf, branch), true)
	require.Equal(t, leaf.String(), hex0x(state.State.FinalizedCheckpoint().Root)) // The merkle root of the leaf is the leaf itself.

	tmplog.Println("leaf                ", leaf)
	tmplog.Println("finalized checkpoint", hex0x(state.State.FinalizedCheckpoint().Root))
}
