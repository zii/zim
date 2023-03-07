package main

import (
	"fmt"
	"hash/fnv"
	"time"

	"github.com/cespare/xxhash"
)

func testFnv() {
	h := fnv.New64()
	st := time.Now()
	var out uint64
	for i := 0; i < 100000000; i++ {
		h.Write([]byte("sdfasf"))
	}
	out = h.Sum64()
	fmt.Println("out:", out, "took:", time.Since(st))
}

func testXXHash() {
	h := xxhash.New()
	st := time.Now()
	var out uint64
	for i := 0; i < 100000000; i++ {
		h.Write([]byte("sdfasf"))
	}
	out = h.Sum64()
	fmt.Println("out:", out, "took:", time.Since(st))
}

func main() {
	//testFnv()
	testXXHash()
}
