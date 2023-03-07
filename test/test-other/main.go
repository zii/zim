package main

import (
	"fmt"
	"math"
	"time"
)

func test1() {
	d := "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
	fmt.Printf("%x", d)
}

func test2() {
	fmt.Println("now ms:", time.Now().UnixMilli())
	fmt.Println("now nano:", time.Now().UnixNano())
	fmt.Println("max int64:", math.MaxInt64)
}

func main() {
	test2()
}
