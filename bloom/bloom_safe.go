package bloom

import "sync"

// SafeBloom wraps BloomFilter with a mutex to allow safe concurrent use.
type SafeBloom struct {
	mu sync.RWMutex
	bf *BloomFilter
}

// NewSafe creates a concurrency-safe Bloom filter using explicit m and k.
func NewSafe(m, k uint64) *SafeBloom {
	return &SafeBloom{bf: New(m, k)}
}

// NewSafeWithEstimates creates a concurrency-safe Bloom filter using n and fpRate.
func NewSafeWithEstimates(n uint64, fpRate float64) *SafeBloom {
	return &SafeBloom{bf: NewWithEstimates(n, fpRate)}
}

// Add inserts data safely.
func (s *SafeBloom) Add(data []byte) {
	s.mu.Lock()
	s.bf.Add(data)
	s.mu.Unlock()
}

// MightContain checks membership safely.
func (s *SafeBloom) MightContain(data []byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bf.MightContain(data)
}

// Reset clears the filter safely.
func (s *SafeBloom) Reset() {
	s.mu.Lock()
	s.bf.Reset()
	s.mu.Unlock()
}

// Info returns metadata safely.
func (s *SafeBloom) Info() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bf.Info()
}
