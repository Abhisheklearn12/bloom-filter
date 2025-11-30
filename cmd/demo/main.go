package main

import (
	"fmt"

	"github.com/Abhisheklearn12/bloom-filter/bloom"
)

func main() {
	bf := bloom.NewSafeWithEstimates(10000, 0.01)
	fmt.Println(bf.Info())

	keys := [][]byte{
		[]byte("abhi"),
		[]byte("golang"),
		[]byte("bloom"),
	}

	for _, k := range keys {
		bf.Add(k)
	}

	checks := [][]byte{
		[]byte("abhi"),
		[]byte("golang"),
		[]byte("bloom"),
		[]byte("kafka"),
		[]byte("redis"),
	}

	for _, c := range checks {
		fmt.Printf("%s: %v\n", c, bf.MightContain(c))
	}
}
