package ztype

import (
	"github.com/protolambda/ztyp/tree"
	"github.com/prysmaticlabs/prysm/io/file"
	"github.com/prysmaticlabs/prysm/testing/require"
	tmplog "log"
	"testing"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func Test_Gindex(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	require.Equal(t, state.GetGIndex(20, 0), tree.Gindex64(104))

	tmplog.Println("20, 1", state.GetGIndex(20, 1))
	tmplog.Println("20, 1", state.GetGIndex(20, 2))
}

func Test_GetPowerOfTwoCeil(t *testing.T) {
	inputs := []uint64{
		0, 1, 2, 4, 5, 8, 9, 16, 17,
	}
	outputs := []uint64{
		1, 1, 2, 4, 5, 8, 16, 16, 32,
	}

	for i, x := range inputs {
		input := tree.Gindex64(x)
		output := GetPowerOfTwoCeil(input)
		tmplog.Println(input, output)
		require.Equal(t, output, tree.Gindex64(outputs[i]))
	}
}

func Test_GetPowerOfTwoFloor(t *testing.T) {
	inputs := []uint64{
		0, 1, 2, 4, 5, 8, 9, 16, 17,
	}
	outputs := []uint64{
		1, 1, 2, 4, 4, 8, 8, 16, 16,
	}

	for i, x := range inputs {
		input := tree.Gindex64(x)
		output := GetPowerOfTwoFloor(input)
		tmplog.Println(input, output)
		require.Equal(t, output, tree.Gindex64(outputs[i]))
	}
}

func Test_ConcatGeneralizedIndices(t *testing.T) {
	require.Equal(t, ConcatGeneralizedIndices(52, 2), tree.Gindex64(104))
	require.Equal(t, ConcatGeneralizedIndices(52, 4), tree.Gindex64(208))
	require.Equal(t, ConcatGeneralizedIndices(52, 7), tree.Gindex64(211))
}
