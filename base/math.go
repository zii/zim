package base

import (
	"crypto/md5"
	crand "crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"math/big"
	"math/rand"
	"strconv"
)

func RandDigitCode(l int) string {
	var out string
	for i := 0; i < l; i++ {
		a := rand.Intn(10)
		out += strconv.Itoa(a)
	}
	return out
}

func GenerateNonce(size int) []byte {
	b := make([]byte, size)
	_, _ = crand.Read(b)
	return b
}

// 生成的字符串长度是size的2倍
func GenerateStringNonce(size int) string {
	b := make([]byte, size)
	_, _ = crand.Read(b)
	return hex.EncodeToString(b)
}

func Sha1Digest(data []byte) []byte {
	r := sha1.Sum(data)
	return r[:]
}

func Sha1String(data []byte) string {
	r := sha1.Sum(data)
	return hex.EncodeToString(r[:])
}

func Md5String(data []byte) string {
	r := md5.Sum(data)
	return hex.EncodeToString(r[:])
}

func StringToGwei(numstr string) int64 {
	z, _ := new(big.Int).SetString(numstr, 10)
	if z == nil {
		return 0
	}
	z.Div(z, big.NewInt(1000000000))
	return z.Int64()
}

// wei to Gwei
func ToGwei(z *big.Int) int64 {
	if z == nil {
		return 0
	}
	x := new(big.Int).Set(z)
	x.Div(x, big.NewInt(1000000000))
	return x.Int64()
}

func ToEthPrice(z *big.Int) float64 {
	if z == nil {
		return 0
	}
	x := new(big.Int).Set(z)
	x.Div(x, big.NewInt(1000000000))
	d := x.Int64()
	return float64(d) / 1000000000
}

func ToBzzPrice(z *big.Int) float64 {
	if z == nil {
		return 0
	}
	x := new(big.Int).Set(z)
	x.Div(x, big.NewInt(1000000000))
	d := x.Int64()
	return float64(d) / 10000000
}

// 两个int64相乘, 返回bigint
func BigMul(a, b int64) *big.Int {
	ai := big.NewInt(a)
	bi := big.NewInt(b)
	return ai.Mul(ai, bi)
}

func FileMd5String(f io.Reader) string {
	h := md5.New()
	_, err := io.Copy(h, f)
	Raise(err)
	return hex.EncodeToString(h.Sum(nil))
}

func AbsInt64(d int64) int64 {
	if d < 0 {
		return -d
	}
	return d
}

// 集合: 求差集 A-B
func Subtract[T comparable](a, b []T) []T {
	var out []T
	if len(a) == 0 {
		return out
	}
	if len(b) == 0 {
		out = make([]T, len(a))
		copy(out, a)
		return out
	}
	bm := make(map[T]struct{}, len(a))
	for _, v := range b {
		bm[v] = struct{}{}
	}
	for _, v := range a {
		if _, ok := bm[v]; !ok {
			out = append(out, v)
		}
	}
	return out
}
