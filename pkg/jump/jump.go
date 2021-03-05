package jump

import (
	"hash/crc64"
	"io"
)

// Hash consistently chooses a hash bucket number in the range [0, numBuckets) for the given key. numBuckets must be >= 1.
func Hash(key uint64, numBuckets int) int32 {

	var b int64 = -1
	var j int64

	for j < int64(numBuckets) {
		b = j
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
	}

	return int32(b)
}

var (
	// hashCRC64 uses the 64-bit Cyclic Redundancy Check (CRC-64) with the ECMA polynomial.
	hashCRC64 = crc64.New(crc64.MakeTable(crc64.ECMA))
)

// HashString takes string as key instead of integer and uses CRC-64 to generate key.
func HashString(key string, buckets int) int32 {
	hashCRC64.Reset()
	_, err := io.WriteString(hashCRC64, key)
	if err != nil {
		panic(err)
	}
	return Hash(hashCRC64.Sum64(), buckets)
}

// Hasher represents a jump consistent Hasher using a string as key.
type Hasher struct {
	n int32
}

// New returns a new instance of of Hasher.
func New(n int) *Hasher {
	if n <= 0 {
		panic("the number of buckets must be positive int")
	}
	return &Hasher{int32(n)}
}

// N returns the number of buckets the hasher can assign to.
func (h *Hasher) N() int {
	return int(h.n)
}

// Hash returns the integer hash for the given key.
func (h *Hasher) Hash(key string) int32 {
	return HashString(key, int(h.n))
}
