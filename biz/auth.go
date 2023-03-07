package biz

import (
	"encoding/base64"
	"strconv"
	"time"

	"zim.cn/base/uuid"

	"zim.cn/base/dh"

	"zim.cn/biz/cookie"
	"zim.cn/biz/proto"

	"zim.cn/base/db"

	"zim.cn/biz/def"

	"zim.cn/biz/cache"
)

// 用户ID生成算法: 每次用5个随机128进制数相乘, 得到一个分布范围:10000000 ~ 34359738368的值, 满足随机且唯一
// 10000000 + [0,128)**5
// 注册用户最多1亿, 让1亿个ID分布在百亿空间内, 应该不容易被猜到.
var userIdDigests = [5][128]int64{
	{118, 87, 17, 120, 61, 34, 0, 46, 36, 12, 112, 107, 62, 42, 16, 9, 86, 35, 54, 70, 64, 104, 38, 63, 103, 52, 47, 43, 126, 101, 95, 30, 68, 51, 74, 81, 77, 49, 45, 31, 18, 32, 78, 89, 109, 79, 10, 82, 57, 84, 102, 50, 115, 124, 88, 59, 108, 121, 67, 65, 8, 72, 66, 80, 76, 106, 92, 21, 127, 27, 33, 123, 99, 37, 93, 55, 2, 14, 114, 40, 15, 11, 105, 4, 7, 29, 113, 53, 96, 26, 19, 58, 20, 1, 69, 100, 41, 28, 91, 23, 39, 111, 119, 71, 73, 60, 48, 24, 116, 25, 98, 110, 97, 44, 56, 90, 94, 22, 83, 5, 3, 85, 122, 75, 6, 117, 125, 13},
	{75, 93, 96, 32, 37, 18, 31, 23, 11, 109, 61, 102, 106, 6, 43, 33, 36, 63, 29, 125, 114, 80, 98, 115, 21, 72, 70, 17, 105, 66, 78, 59, 10, 97, 113, 54, 19, 124, 52, 49, 108, 95, 99, 79, 1, 3, 67, 26, 24, 120, 87, 62, 119, 89, 5, 13, 117, 122, 53, 51, 0, 34, 64, 104, 83, 92, 58, 88, 123, 103, 86, 35, 2, 77, 14, 45, 38, 82, 27, 46, 112, 20, 39, 101, 110, 84, 69, 81, 85, 55, 4, 47, 40, 111, 57, 25, 9, 44, 22, 100, 76, 107, 12, 121, 65, 73, 126, 116, 50, 41, 71, 48, 74, 127, 91, 16, 56, 90, 42, 118, 7, 68, 28, 8, 94, 15, 30, 60},
	{88, 39, 23, 57, 48, 100, 56, 46, 70, 61, 107, 121, 82, 73, 63, 86, 3, 92, 9, 47, 105, 74, 35, 103, 41, 112, 24, 108, 93, 95, 85, 119, 7, 60, 80, 12, 77, 52, 50, 67, 14, 110, 97, 6, 13, 10, 17, 21, 102, 32, 120, 127, 36, 68, 19, 62, 11, 65, 53, 37, 109, 123, 113, 55, 30, 25, 78, 22, 79, 58, 98, 116, 34, 104, 20, 83, 31, 27, 125, 38, 8, 49, 18, 99, 69, 115, 94, 59, 114, 66, 72, 75, 0, 44, 81, 45, 106, 64, 15, 124, 16, 5, 43, 84, 29, 71, 118, 91, 126, 90, 122, 76, 89, 111, 26, 54, 87, 2, 96, 40, 1, 33, 4, 51, 101, 28, 42, 117},
	{62, 120, 50, 23, 65, 104, 77, 123, 34, 73, 108, 122, 118, 80, 88, 29, 119, 55, 89, 91, 116, 114, 5, 2, 92, 14, 21, 8, 99, 31, 126, 39, 27, 79, 7, 102, 97, 45, 110, 111, 11, 115, 33, 38, 83, 13, 10, 17, 100, 53, 125, 68, 82, 20, 109, 112, 32, 61, 1, 113, 96, 86, 64, 42, 74, 66, 16, 43, 95, 54, 81, 58, 37, 40, 26, 60, 69, 36, 124, 19, 63, 35, 41, 18, 25, 30, 87, 127, 107, 46, 22, 71, 15, 49, 70, 57, 103, 101, 44, 48, 94, 93, 72, 78, 59, 67, 47, 51, 98, 56, 85, 105, 0, 3, 117, 12, 4, 121, 84, 75, 28, 9, 106, 6, 90, 76, 52, 24},
	{41, 121, 102, 42, 103, 7, 17, 123, 15, 73, 50, 39, 29, 11, 59, 57, 84, 3, 98, 62, 10, 36, 35, 54, 9, 47, 25, 99, 95, 38, 109, 31, 64, 112, 40, 49, 26, 108, 101, 120, 60, 72, 77, 94, 37, 6, 100, 48, 81, 75, 69, 5, 89, 43, 66, 46, 34, 67, 14, 76, 44, 127, 106, 83, 80, 0, 55, 118, 125, 90, 13, 53, 65, 33, 2, 56, 1, 24, 74, 61, 82, 4, 86, 32, 16, 97, 88, 22, 30, 116, 12, 92, 85, 93, 51, 113, 18, 110, 70, 104, 19, 23, 124, 122, 21, 27, 45, 20, 63, 126, 96, 87, 91, 52, 117, 115, 111, 105, 58, 28, 8, 107, 78, 79, 71, 114, 68, 119},
}

func InitUserId() {
	const initID = 0

	cache.UserIdCounter().SetNX(initID)
	if !cache.UserIdCounter().Exists() {
		panic("Init UserIdCounter fail")
	}
}

func NextUserId() int64 {
	r := cache.UserIdCounter().LuaIncrBy(1)
	if r.Reply == nil {
		panic("cache.UserIdCounter not init!")
	}
	return r.Int64()
}

//func NextUserId() int64 {
//	const offset = 10000000
//
//	r := cache.UserIdCounter().LuaIncrBy(1)
//	if r.Reply == nil {
//		panic("cache.UserIdCounter not init!")
//	}
//	counter := r.Int64()
//	var user_id int64
//	for i := 0; i < len(userIdDigests); i++ {
//		d := counter % 128
//		user_id = user_id*128 + userIdDigests[i][d]
//		counter = counter / 128
//	}
//	return offset + user_id
//}

// 生成用户accid
func GenerateAccid(idType def.IdType) string {
	if def.UseMultiDC {
		return string(idType) + uuid.NextIDString("user")
	}
	id := NextUserId()
	return string(idType) + strconv.FormatInt(id, 10)
}

func CreateUser(name, photo, ex string) string {
	user_id := GenerateAccid(def.IdUser)
	now := int(time.Now().Unix())
	db.Primary.Exec(`insert into user set user_id=?, name=?, photo=?, ex=?, status=0, created_at=?`,
		user_id, name, photo, ex, now)
	// 在注册前调用authToken, 会有残留的用户缓存
	cache.User(user_id).Del()
	return user_id
}

// 交换密钥, 返回(我方公钥, 共享密钥, error)
func ExchangeKey(pubkey string) (string, []byte, error) {
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

func Authorize(user_id string, platform int, device_id string) (*proto.Authorization, error) {
	token, err := cookie.Sign(user_id, platform, device_id)
	if err != nil {
		return nil, err
	}
	out := &proto.Authorization{
		Token: token,
	}
	return out, nil
}
