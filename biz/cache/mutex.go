package cache

import (
	"fmt"

	"zim.cn/base/redis"

	"github.com/go-redsync/redsync"
)

// 限速锁, 主要目的是防止用户连击造成并发问题, 返回false表示锁已经被占用
// seconds: 秒
func LimitRate(key string, seconds int) bool {
	key = "rate:" + key
	r := redis.Do("speed", "SET", key, "1", "EX", seconds, "NX").String()
	return r == "OK"
}

func ThirdSignInMutex(plat int, openid string) *redsync.Mutex {
	return redis.NewMutex(
		"speed",
		fmt.Sprintf("mutex:third_signin:%d:%s", plat, openid),
	)
}
