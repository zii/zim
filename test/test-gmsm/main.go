package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/tjfoc/gmsm/sm4"
	"zim.cn/base"
)

// 测试国密算法

// 100w次/s
func testGmsm() {
	key := []byte("1234567890abcdef")
	raw := []byte("哈哈哈哈和恶搞苹果")
	var encdata []byte
	var err error
	const n = 1000000
	st := time.Now()
	for i := 0; i < n; i++ {
		encdata, err = sm4.Sm4Ecb(key, raw, true)
		base.Raise(err)
	}
	fmt.Println("took:", time.Since(st))
	fmt.Println("encryped:", len(encdata), encdata)

	var dec []byte
	st = time.Now()
	for i := 0; i < n; i++ {
		dec, err = sm4.Sm4Ecb(key, encdata, false)
		base.Raise(err)
	}
	fmt.Println("took:", time.Since(st))
	fmt.Println("decode:", string(dec))
}

// 1000w次/s
func testAESGCM_enc(n int) {
	key := []byte("AES256Key-32Characters1234567890")
	plaintext := []byte("exampleplaintext")

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	fmt.Printf("nonce: %x\n", nonce)
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	st := time.Now()
	var ciphertext []byte
	for i := 0; i < n; i++ {
		ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)
		fmt.Println("ciphertext:", ciphertext)
	}
	fmt.Println("took:", time.Since(st))
	fmt.Printf("%x\n", ciphertext)
}

// 1000w次/s
func testAESGCM_dec(n int) {
	key := []byte("AES256Key-32Characters1234567890")
	ciphertext, _ := hex.DecodeString("27d86eef105f8baf55a853d43b7c6dab1142a627d2ece07f4d175bc8d0c45caf")

	nonce, _ := hex.DecodeString("22a2a7ae2d8673a988c5d3ef")

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	var plaintext []byte
	st := time.Now()
	for i := 0; i < n; i++ {
		plaintext, err = aesgcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			panic(err.Error())
		}
	}
	fmt.Println("took:", time.Since(st))
	fmt.Printf("%s\n", string(plaintext))
}

func main() {
	//testGmsm()
	testAESGCM_enc(10)
	//testAESGCM_dec(1000000)
}
