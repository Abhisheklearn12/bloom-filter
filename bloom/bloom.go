package bloom

import (
	"fmt"
	"math"
)

// Bloomfilter is a standard Bloom Filter implementation.
// Note: This type is not safe for concurrent use without external locking
type BloomFilter struct {
	m    uint64   // no. of bits
	k    uint64   // no. of hash functions
	bits []uint64 //bitset storage
}

// New creates a bloom filter wiht an explicit no. of bits (m) and hash functions (k).
// m and k ==> must be >0.
func New(m, k uint64) *BloomFilter {
	if m == 0 {
		panic("bloom: m (no. of bits) must be > 0")
	}
	if k == 0 {
		panic("bloom: k (no. of hash fucntions) must be > 0")
	}

	wordCount := (m + 63) / 64 // round up to whole 63-bit words
	return &BloomFilter{
		m:    m,
		k:    k,
		bits: make([]uint64, wordCount),
	}
}

// NewWithEstimates constructs a Bloom filter for an expected number of items (n)
// and desired false positive probability (fpRate).
//
// m = - (n * ln(fpRate)) / (ln 2)^2
// k = (m / n) * ln 2
//
// This panics if n == 0 or fpRate is not in (0, 1).
func NewWithEstimates(n uint64, fpRate float64) *BloomFilter {
	if n == 0 {
		panic("bloom: n (expected insertions) must be > 0")
	}
	if fpRate <= 0.0 || fpRate >= 1.0 {
		panic("bloom: fpRate must be between 0 and 1 (exclusive)")
	}

	ln2 := math.Ln2

	mFloat := -float64(n) * math.Log(fpRate) / (ln2 * ln2)
	m := uint64(math.Ceil(mFloat))
	if m == 0 {
		m = 1
	}

	kFloat := (float64(m) / float64(n)) * ln2
	k := uint64(math.Ceil(kFloat))
	if k == 0 {
		k = 1
	}

	return New(m, k)
}

// Add inserts data into the Bloom filter.
func (bf *BloomFilter) Add(data []byte) {
	if bf.m == 0 || bf.k == 0 {
		panic("bloom: filter not initialized")
	}

	h1, h2 := hash128(data)
	if h2 == 0 {
		// avoid degenerate double-hash sequence
		h2 = 0x9e3779b97f4a7c15 // some odd constant
	}

	for i := uint64(0); i < bf.k; i++ {
		// double hashing: position = (h1 + i*h2) mod m
		pos := (h1 + i*h2) % bf.m
		bf.setBit(pos)
	}
}

// MightContain checks if data might be in the filter.
// Returns false -> definitely not present.
// Returns true  -> might be present (subject to false positives).
func (bf *BloomFilter) MightContain(data []byte) bool {
	if bf.m == 0 || bf.k == 0 {
		panic("bloom: filter not initialized")
	}

	h1, h2 := hash128(data)
	if h2 == 0 {
		h2 = 0x9e3779b97f4a7c15
	}

	for i := uint64(0); i < bf.k; i++ {
		pos := (h1 + i*h2) % bf.m
		if !bf.getBit(pos) {
			return false
		}
	}
	return true
}

// Reset clears all bits in the filter.
func (bf *BloomFilter) Reset() {
	for i := range bf.bits {
		bf.bits[i] = 0
	}
}

// Info returns a small description of the filter's configuration.
func (bf *BloomFilter) Info() string {
	return fmt.Sprintf("BloomFilter{m=%d bits, k=%d}", bf.m, bf.k)
}

// setBit sets the bit at position pos (0 <= pos < m).
func (bf *BloomFilter) setBit(pos uint64) {
	wordIndex := pos / 64
	bitIndex := pos % 64
	mask := uint64(1) << bitIndex
	bf.bits[wordIndex] |= mask
}

// getBit returns true if the bit at position pos is set.
func (bf *BloomFilter) getBit(pos uint64) bool {
	wordIndex := pos / 64
	bitIndex := pos % 64
	mask := uint64(1) << bitIndex
	return (bf.bits[wordIndex] & mask) != 0
}

// --- Hashing helpers ---
//
// We implement double hashing using two independent 64-bit FNV-1a hashes.
// h1, h2 = two 64-bit hashes of the data
// position_i = (h1 + i*h2) mod m

const (
	fnv64Offset = 14695981039346656037
	fnv64Prime  = 1099511628211
)

// fnv64a returns the FNV-1a 64-bit hash of data.
func fnv64a(data []byte) uint64 {
	var hash uint64 = fnv64Offset
	for _, b := range data {
		hash ^= uint64(b)
		hash *= fnv64Prime
	}
	return hash
}

// fnv64aSalted is the same as fnv64a but uses a different offset basis
// so that we get an independent second hash.
func fnv64aSalted(data []byte, salt uint64) uint64 {
	var hash uint64 = fnv64Offset ^ salt
	for _, b := range data {
		hash ^= uint64(b)
		hash *= fnv64Prime
	}
	return hash
}

// hash128 produces two 64-bit hashes from the same input.
func hash128(data []byte) (uint64, uint64) {
	h1 := fnv64a(data)
	const salt = 0x9e3779b97f4a7c15 // arbitrary odd 64-bit constant
	h2 := fnv64aSalted(data, salt)
	return h1, h2
}
