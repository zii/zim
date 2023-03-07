package redis

import (
	"sort"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"zim.cn/base"
)

// 计数器分段区间
type CounterSegment struct {
	Id int64 `json:"id"` // 区间的起点data_id
	N  int   `json:"n"`  // 区间内的消息数
}

type CounterSegments []*CounterSegment

// 执行计数 返回数量
// id: 统计id之后有多少条数据(大于,不包含)
// f: 根据(start, end)区间查询行数的业务层函数, end为0表示无上限
// this: 默认为已排序
func (this CounterSegments) Count(id int64, f func(start, end int64) int) int {
	if id < 0 {
		id = 0
	}
	sz := len(this)
	// 倒序遍历
	var end int64 // 上一个区间的起始ID
	var n int
	for i := sz - 1; i >= 0; i-- {
		s := this[i].Id
		if id >= s {
			n += f(id, end)
			return n
		}
		n += this[i].N
		if id == s-1 {
			return n
		}
		end = s
	}
	return n
}

// 分段计数器
// 将一个自增整数分段存储为一个map(hash)结构, key为每个段的起始ID, value为每个段的数量
// 例如: {
//	"0": 10000, // [0~10300)包含10000条数据
//	"10300": 10000, // [10300~当前段起始点ID)包含10000条数据
//	"latest": 20, // [当前段起始点ID~最新]包含20条数据
//	"latest_id": "20300" // 最新段起点ID, 没有则默认为0
// }
// 自增步骤:
// 1. n = map["latest"]++
// 2. if n >= 10000 { map[latest_id]=n; map["latest"]=0; map["latest_id"]=id; 条数超过10条删除最小的key }
// 调用方法:
// lua_segincr xx:counter:map <ID> <LIMIT>
// ID是业务层具体数据的ID,必须为递增的数字,LIMIT是每段最大条数,返回map["latest"]自增后的值,用户如果乱传ID则行为未定义.
var lua_segincr = redis.NewScript(1, `
	local key = KEYS[1];
	local id = tonumber(ARGV[1]);
	local n = tonumber(redis.call("HINCRBY", key, "latest", 1));
	local limit = tonumber(ARGV[2]);
	if n < limit then
		return n;
	end
	local latest_id = tonumber(redis.call("HGET", key, "latest_id")) or 0;
	redis.call("HSET", key, latest_id, n);
	redis.call("HSET", key, "latest", 0);
	redis.call("HSET", key, "latest_id", id+1);
	local size = tonumber(redis.call("HLEN", key)) or 0;
	if size > 11 then
		local min_id = nil;
		local m = redis.call("HGETALL", key);
		for i=1, #m, 2 do
			local sid = tonumber(m[i]);
			if sid ~= nil and (min_id == nil or sid < min_id) then
				min_id = sid;
			end
		end
		if min_id ~= nil then
			redis.call("HDEL", key, min_id);
		end
	end
	return n;
`)

// 分段计数器自增 limit:0默认为10000
// 例: LuaSegincr(1, 10000)
func (this *Key) LuaSegincr(id int64, limit int) *Result {
	if this.Conn != nil {
		err := lua_segincr.SendHash(this.Conn, this.Key, id, limit)
		base.Raise(err)
		return &Result{Reply: nil, Err: nil}
	}
	c := GetRedisPoolClient(this.DBName)
	defer c.Close()
	reply, err := lua_segincr.Do(c, this.Key, id, limit)
	base.Raise(err)
	return &Result{Reply: reply, Err: err}
}

// 初始化分段计数器
func (this *Key) InitSegCounter(segments CounterSegments) {
	if this.Conn != nil {
		c := Begin(this.DBName)
		this.Conn = c
		defer func() {
			this.Conn = nil
			Commit(c)
		}()
	}
	if len(segments) == 0 {
		this.Del()
		this.HSet2("latest", 0)
		return
	}
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].Id < segments[j].Id
	})
	this.Del()
	n := len(segments)
	for i, seg := range segments {
		if i < n-1 {
			this.HSet2(seg.Id, seg.N)
		} else {
			this.HSet2("latest", seg.N)
			this.HSet2("latest_id", seg.Id)
		}
	}
}

// 加载分段计数器 返回排序后的分段列表
func (this *Key) LoadSegCounter() CounterSegments {
	var out CounterSegments
	m := this.HGetAll().Int64Map()
	var latest_id int64
	var latest_n int
	for k, v := range m {
		if k == "latest_id" {
			latest_id = v
		} else if k == "latest" {
			latest_n = int(v)
		} else {
			id, err := strconv.ParseInt(k, 10, 64)
			if err != nil {
				Warn(err)
				continue
			}
			seg := &CounterSegment{
				Id: id,
				N:  int(v),
			}
			out = append(out, seg)
		}
	}
	if latest_id != 0 || len(out) == 0 && latest_n != 0 {
		out = append(out, &CounterSegment{
			Id: latest_id,
			N:  latest_n,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Id < out[j].Id
	})
	return out
}
