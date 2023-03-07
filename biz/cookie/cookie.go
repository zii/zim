package cookie

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"zim.cn/base"
	"zim.cn/base/redis"
	"zim.cn/biz/cache"
)

const (
	CurrentVersion  = 1
	UserCookieLimit = 5 // 同一账号最多保留cookie数
)

var V1Secret = []byte("aK6n7OdxMAGR7YbV") // DO NOT MODIFY

var ErrNewTokenDuplicated = fmt.Errorf("err generate token duplicated")
var ErrCookieInvalid = fmt.Errorf("cookie invalid")
var ErrOtherDeviceOnline = fmt.Errorf("other device online")

func md5_4b(data []byte) []byte {
	hash := md5.Sum(data)
	var out [4]byte
	for i := 0; i < 4; i++ {
		out[i] = hash[i] ^ hash[i+4] ^ hash[i+8] ^ hash[i+12]
	}
	return out[:]
}

// token基本格式: {1字节:版本号}{bytes}
func NewToken(version uint8) string {
	if version != 1 {
		panic("version invalid")
	}
	return NewTokenV1()
}

// token v1: {1字节:版本号}{4字节:nonce}{4字节:md5}
func NewTokenV1() string {
	var buf [9]byte
	buf[0] = byte(1)
	nonce := base.GenerateNonce(4)
	copy(buf[1:5], nonce)
	var key = make([]byte, len(V1Secret)+5)
	copy(key[:5], buf[:5])
	copy(key[5:], V1Secret)
	hash := md5_4b(key)
	copy(buf[5:], hash)
	return hex.EncodeToString(buf[:])
}

func ParseToken(tokenString string) bool {
	buf, err := hex.DecodeString(tokenString)
	if err != nil {
		logrus.Error("parse token")
		return false
	}
	version := uint8(buf[0])
	if version != 1 {
		logrus.Error("token version invalid")
		return false
	}
	return ParseTokenV1(buf)
}

func ParseTokenV1(buf []byte) bool {
	if len(buf) != 9 {
		return false
	}
	sign := buf[5:9]
	var key = make([]byte, len(V1Secret)+5)
	copy(key[:5], buf[:5])
	copy(key[5:], V1Secret)
	hash := md5_4b(key)
	return bytes.Equal(sign, hash)
}

func CreateToken() string {
	for i := 0; i < 3; i++ {
		token := NewToken(CurrentVersion)
		key := cache.CookieKey(token)
		if key.Exists() {
			continue
		}
		return token
	}
	return ""
}

// 回收多个cookie
func FreeCookies(tokens []string) int {
	if len(tokens) <= 0 {
		return 0
	}
	dbname := cache.CookieKey("").DBName
	var args []interface{}
	for _, token := range tokens {
		args = append(args, cache.CookieKey(token).Key)
	}
	return redis.Do(dbname, "DEL", args...).Int()
}

// 强制回收多出的cookie, 腾出个地方
func PopUserCookies(user_id string) int {
	key := cache.UserCookiesKey(user_id)

	// 回收过期的
	now := time.Now().Unix()
	tokens := key.Do("ZRANGEBYSCORE", "-inf", now).Strings()

	// 回收超限的
	n := key.ZCard()
	if n >= UserCookieLimit {
		rows := key.ZPopMin(1)
		if len(rows) == 2 {
			token := rows[0]
			tokens = append(tokens, token)
		}
	}

	return FreeCookies(tokens)
}

// 删除用户的token
func DelUserToken(user_id string, token string) {
	key := cache.UserCookiesKey(user_id)
	key.ZRem(token)

	FreeCookies([]string{token})
}

// 删除用户所有token
func ClearUserToken(user_id string) {
	key := cache.UserCookiesKey(user_id)
	tokens := key.Do("ZRANGE", 0, -1).Strings()
	FreeCookies(tokens)
	key.Del()
}

// 签署token
func Sign(user_id string, platform int, device_id string) (string, error) {
	token := CreateToken()
	if token == "" {
		return "", ErrNewTokenDuplicated
	}
	PopUserCookies(user_id)

	key := cache.CookieKey(token)
	now := int(time.Now().Unix())
	expire_at := now + key.Timeout
	val := &cache.CookieValue{
		UserId:   user_id,
		ExpireAt: expire_at,
		Platform: platform,
		DeviceId: device_id,
	}
	key.Set(val)

	key2 := cache.UserCookiesKey(user_id)
	key2.ZAdd(token, expire_at)
	key2.Expire(key.Timeout)

	return token, nil
}

// 解析token, 返回cookieValue
func Parse(token string) (*cache.CookieValue, error) {
	if token == "" {
		return nil, ErrCookieInvalid
	}
	if !ParseToken(token) {
		return nil, ErrCookieInvalid
	}
	key := cache.CookieKey(token)
	v := &cache.CookieValue{}
	err := key.Get().Unmarshal(&v)
	if err == redis.ErrNil {
		return nil, ErrCookieInvalid
	}
	if err != nil {
		return nil, err
	}
	if v.UserId == "" {
		return nil, ErrCookieInvalid
	}
	now := int(time.Now().Unix())
	ttl := v.ExpireAt - now
	if ttl <= 0 {
		return nil, ErrCookieInvalid
	}
	// auto renew
	if ttl <= key.Timeout*2/3 {
		v.ExpireAt = now + key.Timeout
		key.Set(v)
	}
	return v, nil
}
