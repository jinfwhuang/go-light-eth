package utils

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"

	///Users/jin/code/repos/prysm/beacon-chain/state/v2

	//"github.com/prysmaticlabs/prysm/testing/util"
	tmplog "log"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func GetReadableHash(hash [32]byte) string {
	//return strconv.Itoa(int(binary.BigEndian.Uint32(hash[:])))
	return HashBytes(hash[:])
}

func ToHex(bytearray []byte) string {
	//return strconv.Itoa(int(binary.BigEndian.Uint32(hash[:])))
	return hex.EncodeToString(bytearray)
}

func HashBytes(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func Base64Str(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
