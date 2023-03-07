package biz

import (
	"math"
	"sync"
	"time"

	"zim.cn/base/log"

	redigo "github.com/gomodule/redigo/redis"

	"zim.cn/base/redis"
	"zim.cn/biz/cache"
	"zim.cn/biz/def"
	"zim.cn/biz/proto"
)

// 持久化单条消息 redis
func flushRedisTx(c redigo.Conn, cmd *proto.Command) {
	msg := cmd.Message
	if msg == nil {
		return
	}
	// 修改对话状态, 保存消息
	tot := def.ToIdType(msg.ToId)
	if tot == def.IdUser || tot == def.IdGroup {
		now := int(time.Now().Unix())
		// 全体刷新对话时间, 其他人增加未读数
		for _, user_id := range cmd.UserIds {
			peer_id := msg.GetPeerId(user_id)
			score := now
			// 将系统账号提升到对话列表首页
			if def.ToIdType(peer_id) == def.IdSys {
				score = math.MaxInt64
			}
			cache.UserDialogZset(user_id).Tx(c).ZAdd(peer_id, score)
			cache.DialogPts(user_id, peer_id).Tx(c).Set(msg.Id)
			if user_id != msg.FromId {
				cache.DialogUnread(user_id, peer_id).Tx(c).LuaIncrBy(1)
			}
		}
		// 发信人自动已读
		FastReadHistoryTx(c, msg.FromId, msg.ToId, msg.Id)
	} else if tot == def.IdChannel {
		channel_id := msg.ToId
		// 发信人自动已读
		FastReadHistoryTx(c, msg.FromId, channel_id, msg.Id)
		// 超级群消息计数+1
		cache.ChannelCounter(channel_id).Tx(c).LuaSegincr(msg.Id, def.SegCounterLimit)
	}
}

func FlushRedis(cmd *proto.Command) {
	c := redis.Begin("cache")
	flushRedisTx(c, cmd)
	redis.Commit(c)
}

// 批量持久化
func BulkFlushRedis(cmds []*proto.Command) {
	log.Info("BulkFlushRedis:", len(cmds))
	c := redis.Begin("cache")
	for _, cmd := range cmds {
		flushRedisTx(c, cmd)
	}
	redis.Commit(c)
}

// 持久化单条消息 db
func FlushDB(cmd *proto.Command) {
	msg := cmd.Message
	if msg == nil {
		return
	}
	// 修改对话状态, 保存消息
	tot := def.ToIdType(msg.ToId)
	if tot == def.IdUser || tot == def.IdGroup {
		input := &InputUserMessage{
			UserIds: cmd.UserIds,
			Msg:     msg,
		}
		InsertUserMessage(input)
	} else if tot == def.IdChannel {
		InsertChannelMessage(msg)
	}
}

// 批量持久化消息 mysql
func BulkFlushDB(cmds []*proto.Command) {
	var inputs []*InputUserMessage
	var msgs []*proto.Message
	for _, cmd := range cmds {
		msg := cmd.Message
		if msg == nil {
			return
		}
		tot := def.ToIdType(msg.ToId)
		if tot == def.IdUser || tot == def.IdGroup {
			input := &InputUserMessage{
				UserIds: cmd.UserIds,
				Msg:     msg,
			}
			inputs = append(inputs, input)
		} else if tot == def.IdChannel {
			msgs = append(msgs, msg)
		}
	}
	log.Println("BulkFlushDB:", len(inputs), len(msgs))
	if len(inputs) > 0 {
		BulkdInsertUserMessage(inputs)
	}
	if len(msgs) > 0 {
		BulkInserChannelMessage(msgs)
	}
}

type FlushFunc func(cmds []*proto.Command)

// 持久化单条消息
func FlushMessage(cmd *proto.Command) {
	FlushRedis(cmd)
	FlushDB(cmd)
}

// 消息刷写器, 定时定量刷入数据库
type Flusher struct {
	buf      []*proto.Command
	mu       *sync.RWMutex
	limit    int           // 积累到多少条开始批量插入
	interval time.Duration // 定时多少秒刷入
	h        FlushFunc
}

func NewFlusher(limit int, interval time.Duration, h FlushFunc) *Flusher {
	f := &Flusher{
		mu:       &sync.RWMutex{},
		limit:    limit,
		interval: interval,
		h:        h,
	}
	return f
}

func (f *Flusher) Push(cmd *proto.Command) {
	f.mu.Lock()
	f.buf = append(f.buf, cmd)
	f.mu.Unlock()
	if len(f.buf) >= f.limit {
		f.Flush()
	}
}

func (f *Flusher) Flush() {
	if len(f.buf) == 0 {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	buf := f.buf
	f.buf = f.buf[:0]
	log.Println("flush:", len(buf))
	if f.h != nil {
		f.h(buf)
	}
}

func (f *Flusher) Run() {
	for {
		time.Sleep(f.interval)
		f.Flush()
	}
}
