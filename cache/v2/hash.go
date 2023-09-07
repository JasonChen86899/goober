package v2

import "github.com/cespare/xxhash/v2"

// SeparatorByte is a byte that cannot occur in valid UTF-8 sequences and is
// used to separate label names, label values, and other strings from each other
// when calculating their combined hash value (aka signature aka fingerprint).
const SeparatorByte byte = 255

var (
	separatorByteSlice = []byte{SeparatorByte}
)

//func hash(key interface{}) uint64 {
//	switch key.(type) {
//	case uint, int, uint8, int8, uint32, int32, uint64, int64:
//
//	}
//}

func strHash(key string) uint64 {
	//xxh := xxhash.New()
	//xxh.WriteString(key)
	//xxh.Write(separatorByteSlice)
	//return xxh.Sum64()

	return xxhash.Sum64String(key)
}

const (
	reHashThreshold = 0.75
	initBPower      = 3
)
