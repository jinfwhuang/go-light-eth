package ssz_test

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	v2 "github.com/prysmaticlabs/prysm/beacon-chain/state/v2"
	"github.com/prysmaticlabs/prysm/io/file"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/testing/require"
	tmplog "log"
	"testing"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func hashBytes(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func toHex(bytearray []byte) string {
	return hex.EncodeToString(bytearray)
}

func Test_DifferentMarshallSsz(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"

	bytearray1, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	hash1 := hashBytes(bytearray1)

	// Get different state representation
	ethpbState := &ethpb.BeaconStateAltair{}
	err = ethpbState.UnmarshalSSZ(bytearray1)
	require.NoError(t, err)
	state, err := v2.InitializeFromProto(ethpbState)
	require.NoError(t, err)

	// Get different ssz bytes
	bytearray2, err := ethpbState.MarshalSSZ()
	require.NoError(t, err)
	hash2 := hashBytes(bytearray2)

	bytearray3, err := state.MarshalSSZ()
	require.NoError(t, err)
	hash3 := hashBytes(bytearray3)

	// Validation
	require.Equal(t, hash1, hash2)
	require.Equal(t, hash1, hash3)
	require.Equal(t, hash2, hash3)
}

func Test_DifferentHashRoot(t *testing.T) {
	ctx := context.Background()
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	bytearray, err := file.ReadFileAsBytes(filename)
	if err != nil {
		panic(err)
	}

	// fastssz HashRoot
	state1 := &ethpb.BeaconStateAltair{}
	err = state1.UnmarshalSSZ(bytearray)
	require.NoError(t, err)
	root1, err := state1.HashTreeRoot()
	require.NoError(t, err)

	// state_trie.go HashRoot
	state2, err := v2.InitializeFromProto(state1)
	require.NoError(t, err)
	root2, err := state2.HashTreeRoot(ctx)
	require.NoError(t, err)

	// Validation
	require.NotEqual(t, root1, root2) // Why this is not equal?

	tmplog.Println(toHex(root1[:]))
	tmplog.Println(toHex(root2[:]))

}
