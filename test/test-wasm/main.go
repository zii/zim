package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"syscall/js"
	"time"

	"zim.cn/base/dh"
)

/*
 *	NEVER REUSE A NONCE WITH THE SAME KEY
 */

var id int

// from base/sdk/appkey.go WEB SRECRET
// TODO: obfuscation
var GCM_KEY = "zt5kqoocul3l1otvnjxoltbpx4vnteyn"
var APP_SECRET = GCM_KEY

var pool = map[int]cipher.AEAD{}

func raise(err error) {
	if err != nil {
		panic(err)
	}
}

// 交换密钥, 返回(我方公钥, 共享密钥, error)
func exchangeKey(pubkey string) (string, []byte, error) {
	pubdata, err := base64.StdEncoding.DecodeString(pubkey)
	if err != nil {
		return "", nil, err
	}
	p := dh.NewPeer(nil)
	p.RecvPeerPubKey(pubdata)
	key := p.GetKey()
	mypub := base64.StdEncoding.EncodeToString(p.GetPubKey())
	return mypub, key, nil
}

// slow: 400ms
func callExchangeKey(this js.Value, args []js.Value) any {
	mypub, key, err := exchangeKey(args[0].String())
	_, _, _ = mypub, key, err
	var skey string
	if key != nil {
		skey = base64.StdEncoding.EncodeToString(key)
	}
	var errs string
	if err != nil {
		errs = err.Error()
	}
	out := map[string]any{
		"mypub": mypub,
		"key":   skey,
		"err":   errs,
	}
	return js.ValueOf(out)
}

func Sha1String(data []byte) string {
	r := Sum(data)
	return hex.EncodeToString(r[:])
}

func sign(path string) string {
	s := fmt.Sprintf("%s%d%s", path, time.Now().Unix(), APP_SECRET)
	return Sha1String([]byte(s))
}

func callSign(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		panic("Incorrect number of arguments, example: zim_sign('path')")
	}
	if args[0].Type() != js.TypeString {
		panic("Invalid argument type, example: zim_sign('path')")
	}
	path := args[0].String()
	out := sign(path)
	return js.ValueOf(out)
}

func encrypt(id int, plaintext string, nonce string) string {
	var aesgcm cipher.AEAD
	if id != 0 {
		aesgcm = pool[id]
		if aesgcm == nil {
			panic("id invalid")
		}
	} else {
		block, err := aes.NewCipher([]byte(GCM_KEY))
		raise(err)

		aesgcm, err = cipher.NewGCM(block)
		raise(err)
	}
	ciphertext := aesgcm.Seal(nil, []byte(nonce), []byte(plaintext), nil)
	out := base64.StdEncoding.EncodeToString(ciphertext)
	return out
}

// fast
func callEncrypt(this js.Value, args []js.Value) any {
	if len(args) != 2 {
		panic("Incorrect number of arguments, example: zim_encrypt('plaintext', 'nonce')")
	}
	if args[0].Type() != js.TypeString {
		panic("Invalid argument type")
	}
	if args[1].Type() != js.TypeString {
		panic("Invalid argument type")
	}
	plaintext := args[0].String()
	nonce := args[1].String()
	out := encrypt(0, plaintext, nonce)
	return js.ValueOf(out)
}

func decrypt(id int, ciphertext string, nonce string) string {
	cipherdata, err := base64.StdEncoding.DecodeString(ciphertext)
	raise(err)
	var aesgcm cipher.AEAD
	if id != 0 {
		aesgcm = pool[id]
		if aesgcm == nil {
			panic("id invalid")
		}
	} else {
		block, err := aes.NewCipher([]byte(GCM_KEY))
		raise(err)
		aesgcm, err = cipher.NewGCM(block)
		raise(err)
	}
	plaintext, err := aesgcm.Open(nil, []byte(nonce), cipherdata, nil)
	raise(err)
	return string(plaintext)
}

// fast
func callDecrypt(this js.Value, args []js.Value) any {
	if args[0].Type() != js.TypeString {
		panic("Invalid argument type, example: zim_decrypt('ciphertext', 'nonce')")
	}
	if args[1].Type() != js.TypeString {
		panic("Invalid argument type")
	}
	ciphertext := args[0].String()
	nonce := args[1].String()
	out := decrypt(0, ciphertext, nonce)
	return js.ValueOf(out)
}

func newAead() int {
	id++
	block, err := aes.NewCipher([]byte(GCM_KEY))
	raise(err)
	aesgcm, err := cipher.NewGCM(block)
	raise(err)
	pool[id] = aesgcm
	return id
}

func callNewAead(this js.Value, args []js.Value) any {
	id := newAead()
	return js.ValueOf(id)
}

func main() {
	wait := make(chan struct{}, 0)
	js.Global().Set("zim_sign", js.FuncOf(callSign))
	js.Global().Set("zim_dh", js.FuncOf(callExchangeKey))
	js.Global().Set("zim_aead", js.FuncOf(callNewAead))
	js.Global().Set("zim_encrypt", js.FuncOf(callEncrypt))
	js.Global().Set("zim_decrypt", js.FuncOf(callDecrypt))
	<-wait
}
