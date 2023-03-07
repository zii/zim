// 进程内缓存
package cache

import (
	"sync/atomic"
	"time"

	"zim.cn/base/log"

	cache "github.com/patrickmn/go-cache"
	"zim.cn/base/redis"
)

// 可远程擦除的本地缓存
// 依赖redis
type ErasableLocal struct {
	*cache.Cache
	name    string
	started int32
}

func (this *ErasableLocal) Set(k string, v interface{}, d time.Duration) {
	this.Cache.Set(k, v, d)
	if atomic.CompareAndSwapInt32(&this.started, 0, 1) {
		// start when redis installed
		pool := redis.GetPool(LocalRevisonNo(this.name).DBName)
		if pool != nil {
			go this.loop()
		} else {
			log.Warn("ErasableLocal redis is not work!")
		}
	}
}

func (this *ErasableLocal) checkRevision(current int) (int, bool) {
	rev := LocalRevisonNo(this.name).Get().Int()
	if rev > current {
		for i := current + 1; i <= rev; i++ {
			k := LocalRevisionK(this.name, i).Get().String()
			this.Delete(k)
		}
	}
	return rev, rev > current
}

func (this *ErasableLocal) loop() {
	var rev = LocalRevisonNo(this.name).Get().Int()
	for {
		time.Sleep(time.Second * 3)
		next, _ := this.checkRevision(rev)
		rev = next
	}
}

func (this *ErasableLocal) Erase(k string) {
	this.Delete(k)
	rev := LocalRevisonNo(this.name).Incr().Int()
	LocalRevisionK(this.name, rev).Set(k)
}

// 一般性的缓存
var L0 *cache.Cache

// 专门缓存配置
var L1 *ErasableLocal

func init() {
	L0 = cache.New(time.Hour*24, time.Hour)
	L1 = &ErasableLocal{
		Cache: cache.New(time.Hour*24, time.Hour),
		name:  "L1",
	}
}
