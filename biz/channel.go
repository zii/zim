package biz

import (
	"encoding/json"
	"fmt"
	"time"

	"zim.cn/base/uuid"

	"zim.cn/base"

	"zim.cn/base/log"

	"zim.cn/biz/cache"
	"zim.cn/biz/def"

	"zim.cn/base/db"
	"zim.cn/base/redis"
	"zim.cn/biz/proto"
)

func nextChannelMessageId(chat_id string) int64 {
	now := time.Now().Unix()
	if def.UseMultiDC {
		msg_id := uuid.NextID("channel")
		c := redis.Begin("cache")
		cache.ChannelTime(chat_id).Set(now)
		cache.ChannelPts(chat_id).Set(msg_id)
		redis.Commit(c)
		return msg_id
	}
	c := redis.Begin("cache")
	cache.ChannelTime(chat_id).Tx(c).Set(now)
	cache.ChannelPts(chat_id).Tx(c).Incr().Int64()
	r := redis.Commit(c)
	base.Raise(r.Err)
	vals := r.Values()
	msg_id := vals[1].(int64)
	return msg_id
}

func GetChannelPts(chat_id string) int64 {
	return cache.ChannelPts(chat_id).Get().Int64()
}

func GetChannelPtsMap(chat_ids []string) map[string]int64 {
	var out = make(map[string]int64)
	if len(chat_ids) == 0 {
		return out
	}
	var keys []any
	for _, chat_id := range chat_ids {
		k := cache.ChannelPts(chat_id).Key
		keys = append(keys, k)
	}
	vals := redis.Do("cache", "MGET", keys...).Int64s()
	for i, pts := range vals {
		chat_id := chat_ids[i]
		out[chat_id] = pts
	}
	return out
}

// 批量获得超级群最新消息时间
func GetChannelTimeMap(chat_ids []string) map[string]int {
	var out = make(map[string]int)
	if len(chat_ids) == 0 {
		return out
	}
	var keys []any
	for _, chat_id := range chat_ids {
		k := cache.ChannelTime(chat_id).Key
		keys = append(keys, k)
	}
	vals := redis.Do("cache", "MGET", keys...).Ints()
	for i, t := range vals {
		chat_id := chat_ids[i]
		out[chat_id] = t
	}
	return out
}

func loadChannelMessage(chat_id string, msg_id int64) *proto.Message {
	var blob []byte
	db.Replica.Get(&blob, `select msg_blob from channel_msgbox where chat_id=? and msg_id=?`,
		chat_id, msg_id)
	var out *proto.Message
	json.Unmarshal(blob, &out)
	return out
}

func GetChannelMessage(chat_id string, msg_id int64) *proto.Message {
	key := cache.ChannelMsgKey(chat_id, msg_id)
	r := key.Get()
	if r.Reply != nil {
		var m *proto.Message
		err := r.Unmarshal(&m)
		if err == nil {
			return m
		}
	}
	m := loadChannelMessage(chat_id, msg_id)
	key.Set(m)
	return m
}

func loadChannelMessageMap(chat_id string, msg_ids []int64) map[int64]*proto.Message {
	var out = make(map[int64]*proto.Message)
	if len(msg_ids) == 0 {
		return out
	}
	q := fmt.Sprintf("select msg_id, msg_blob from channel_msgbox where chat_id=? and msg_id in (%s)",
		db.JoinArray(msg_ids))
	rows := db.Replica.Query(q, chat_id)
	defer rows.Close()
	for rows.Next() {
		var msg_id int64
		var blob []byte
		rows.Scan(&msg_id, &blob)
		var m *proto.Message
		json.Unmarshal(blob, &m)
		if m != nil {
			m.Id = msg_id
			out[msg_id] = m
		}
	}
	return out
}

func GetChannelMessageMap(chat_id string, msg_ids []int64) map[int64]*proto.Message {
	var out = make(map[int64]*proto.Message)
	if len(msg_ids) == 0 {
		return out
	}
	var keys = make([]any, 0, len(msg_ids))
	for _, msg_id := range msg_ids {
		k := cache.ChannelMsgKey(chat_id, msg_id).Key
		keys = append(keys, k)
	}
	var missing []int64
	vals := redis.Do("speed", "MGET", keys...).Strings()
	for i, s := range vals {
		msg_id := msg_ids[i]
		if s == "" {
			missing = append(missing, msg_id)
			continue
		}
		var m *proto.Message
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
			log.Error("Unmarshal msg:", err)
			missing = append(missing, msg_id)
			continue
		}
		out[msg_id] = m
	}
	if len(missing) > 0 {
		m := loadChannelMessageMap(chat_id, missing)
		c := redis.Begin("speed")
		for msg_id, msg := range m {
			out[msg_id] = msg
			cache.ChannelMsgKey(chat_id, msg_id).Tx(c).Set(msg)
		}
		redis.Commit(c)
	}
	return out
}

func GetChannelMessages(chat_id string, msg_ids []int64) []*proto.Message {
	var out []*proto.Message
	d := GetChannelMessageMap(chat_id, msg_ids)
	for _, id := range msg_ids {
		m := d[id]
		if m != nil {
			out = append(out, m)
		}
	}
	return out
}

type InputChanMsg struct {
	ChatId string
	MsgId  int64
}

// 根据(chat_id, msg_id)组合批量查询消息
func GetChannelsMessageMap(input []*InputChanMsg) map[InputChanMsg]*proto.Message {
	var out = make(map[InputChanMsg]*proto.Message)
	if len(input) == 0 {
		return out
	}
	var args []any
	q := `select chat_id, msg_id, msg_blob from channel_msgbox where (chat_id, msg_id) in (`
	for _, in := range input {
		q += fmt.Sprintf("(?, ?),")
		args = append(args, in.ChatId, in.MsgId)
	}
	q = q[:len(q)-1] + ")"
	rows := db.Replica.Query(q, args...)
	defer rows.Close()
	for rows.Next() {
		var in = InputChanMsg{}
		var blob []byte
		rows.Scan(&in.ChatId, &in.MsgId, &blob)
		var msg *proto.Message
		json.Unmarshal(blob, &msg)
		out[in] = msg
	}
	return out
}

func UpdateChannelMessage(msg *proto.Message) bool {
	if def.ToIdType(msg.ToId) != def.IdChannel {
		return false
	}
	blob, err := json.Marshal(msg)
	if err != nil {
		log.Println("Marshal msg:", err)
		return false
	}
	ok := db.Primary.Exec(`update channel_msgbox set msg_blob=? where chat_id=? and msg_id=?`,
		blob, msg.ToId, msg.Id).OK()
	if ok {
		cache.ChannelMsgKey(msg.ToId, msg.Id).Set(msg)
	}
	return ok
}

func delChannelMessages(chat_id string, msg_ids []int64) bool {
	q := `delete from channel_msgbox`
	p := db.Prepare()
	p.And("chat_id=?", chat_id)
	p.In("msg_id", msg_ids)
	q += p.Where()
	ok := db.Primary.Exec(q, p.Args()...).OK()
	// clear cache
	if ok {
		var keys []any
		for _, msg_id := range msg_ids {
			k := cache.ChannelMsgKey(chat_id, msg_id).Key
			keys = append(keys, k)
		}
		redis.Do("speed", "del", keys...)
	}
	return ok
}

// 倒序翻页获取超级群消息ID
func pageDownChannelMsgIds(chat_id string, max_id int64, limit int) []int64 {
	var out []int64
	q := `select msg_id from channel_msgbox`
	p := db.Prepare()
	p.And("chat_id=?", chat_id)
	if max_id > 0 {
		p.And("msg_id<?", max_id)
	}
	p.Sort("msg_id desc")
	p.Slice(0, limit)
	q += p.Clause()
	rows := db.Replica.Query(q, p.Args()...)
	defer rows.Close()
	for rows.Next() {
		var msg_id int64
		rows.Scan(&msg_id)
		out = append(out, msg_id)
	}
	return out
}

// 初始化超级群计数器
func InitChannelCounter(chat_id string) {
	var segs redis.CounterSegments
	var max_id int64
	for i := 0; i < 10; i++ {
		msg_ids := pageDownChannelMsgIds(chat_id, max_id, def.SegCounterLimit)
		n := len(msg_ids)
		if n == 0 {
			break
		}
		if i == 0 && n >= def.SegCounterLimit {
			segs = append(segs, &redis.CounterSegment{
				Id: msg_ids[0] + 1,
				N:  0,
			})
		}
		segs = append(segs, &redis.CounterSegment{
			Id: msg_ids[n-1],
			N:  n,
		})
		max_id = msg_ids[n-1]
		if n < def.SegCounterLimit {
			break
		}
	}
	cache.ChannelCounter(chat_id).InitSegCounter(segs)
}

// 查询用户超级群未读数
// read_id: 上次已读消息ID
func GetChannelUnread(chat_id string, read_id int64) int {
	segs := cache.ChannelCounter(chat_id).LoadSegCounter()
	n := segs.Count(read_id, func(s, e int64) int {
		var cnt int
		q := `select count(*) from channel_msgbox`
		p := db.Prepare()
		p.And("chat_id=?", chat_id)
		p.And("msg_id>?", s)
		if e > 0 {
			p.And("msg_id<?", e)
		}
		q += p.Clause()
		db.Replica.Get(&cnt, q, p.Args()...)
		return cnt
	})
	return n
}
