package bloom

import (
	"strconv"
	"testing"
)

func TestBloom_NoFalseNegatives(t *testing.T) {
	bf := NewWithEstimates(1000, 0.01)

	// Insert a bunch of keys
	const count = 1000
	keys := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		key := []byte("key-" + strconv.Itoa(i))
		keys = append(keys, key)
		bf.Add(key)
	}

	// Check they all "might contain" (i.e., no false negatives)
	for i, key := range keys {
		if !bf.MightContain(key) {
			t.Fatalf("expected key %d to be present, but got false", i)
		}
	}
}

func TestBloom_NegativeExample(t *testing.T) {
	bf := New(1024, 3)

	bf.Add([]byte("hello"))
	bf.Add([]byte("world"))

	if !bf.MightContain([]byte("hello")) {
		t.Fatal(`expected "hello" to be present`)
	}
	if !bf.MightContain([]byte("world")) {
		t.Fatal(`expected "world" to be present`)
	}

	// This should be very likely absent (but can be a false positive).
	// The important property is: if it returns false, then it is definitely absent.
	if bf.MightContain([]byte("another-key")) {
		// Not a test failure: just log.
		t.Log(`"another-key" reported as present (false positive is allowed)`)
	}
}

func TestBloom_Reset(t *testing.T) {
	bf := New(512, 4)

	bf.Add([]byte("foo"))
	if !bf.MightContain([]byte("foo")) {
		t.Fatal(`expected "foo" to be present before reset`)
	}

	bf.Reset()

	// After reset, "foo" must no longer be reported as present.
	if bf.MightContain([]byte("foo")) {
		t.Fatal(`expected "foo" to be absent after reset`)
	}
}
