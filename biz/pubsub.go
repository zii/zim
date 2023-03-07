package biz

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"zim.cn/biz/cache"

	"zim.cn/biz/proto"

	"zim.cn/biz/def"

	"zim.cn/base"

	redigo "github.com/gomodule/redigo/redis"
	"zim.cn/base/redis"
)

func TestSubscribe() {
	c := redis.GetRedisPoolClient("pubsub")
	sub := redigo.PubSubConn{Conn: c}
	err := sub.Subscribe(def.MESSAGE_CHANNEL)
	base.Raise(err)
	for {
		v := sub.Receive()
		switch r := v.(type) {
		case redigo.Message:
			fmt.Println("message:", r.Channel, r.Data)
		case redigo.Subscription:
			fmt.Println("subscr:", r.Channel, r.Kind, r.Count)
		case error:
			fmt.Println("err:", err)
			return
		}
	}
}

// 结果: 批量10000条:86ms 逐条10000条:5.4s
func TestPipeline(mode int, n int) {
	if mode == 1 {
		// 批量插入
		c := redis.GetRedisPoolClient("cache")
		c.Send("MULTI")
		for i := 0; i < n; i++ {
			err := c.Send("SET", base.GenerateStringNonce(4), base.GenerateStringNonce(4), "EX", 30)
			base.Raise(err)
		}
		r, err := c.Do("EXEC")
		base.Raise(err)
		fmt.Println("r:", len(r.([]any)))
	} else if mode == 2 {
		// 批量插入
		c := redis.Begin("cache")
		for i := 0; i < n; i++ {
			//cache.ChannelPts(base.GenerateStringNonce(4)).Tx(c).Set(base.GenerateStringNonce(4))
			cache.UserDialogZset(base.GenerateStringNonce(4)).Tx(c).ZAdd(base.GenerateStringNonce(4), rand.Float64())
		}
		r := redis.Commit(c)
		fmt.Println("r:", r.Strings())
	} else {
		// 逐条插入
		for i := 0; i < n; i++ {
			redis.Do("cache", "SET", base.GenerateStringNonce(4), base.GenerateStringNonce(4), "EX", 30)
		}
	}
}

// 实时推送消息给gateway
func PushCommand(cmd *proto.Command) {
	if cmd == nil {
		return
	}
	b, _ := json.Marshal(cmd)
	redis.Do("cache", "PUBLISH", def.MESSAGE_CHANNEL, b)
}

// 推送强制下线命令
func PushCmdDisconnect(user_id string, token string, reason string) {
	cmd := &proto.Command{
		Op: def.OpDisconnect,
		CmdDisconnect: &proto.CmdDisconnect{
			UserId: user_id,
			Token:  token,
			Reason: reason,
		},
	}
	PushCommand(cmd)
}

// 多端同步事件: 自己已读对话消息
func PushEvHasRead(user_id, peer_id string, max_id int64, unread int) {
	msg := &proto.Message{
		Type: def.EvHasRead,
		Event: &proto.Event{
			HasRead: &proto.EvHasRead{
				PeerId: peer_id,
				MaxId:  max_id,
				Unread: unread,
			},
		},
	}
	cmd := &proto.Command{
		Op:      def.OpSend,
		UserIds: []string{user_id},
		Message: msg,
	}
	PushCommand(cmd)
}

// 多端同步事件: 主动退群
func PushEvQuitChat(user_id, chat_id string) {
	msg := &proto.Message{
		Type: def.EvQuitChat,
		Event: &proto.Event{
			QuitChat: &proto.EvQuitChat{
				ChatId: chat_id,
			},
		},
	}
	cmd := &proto.Command{
		Op:      def.OpSend,
		UserIds: []string{user_id},
		Message: msg,
	}
	PushCommand(cmd)
}

// 推送已读回执事件
// user_ids: 单聊: 自己ID, 普通群: 所有成员ID, 超级群: 空
func PushEvReceipt(user_ids []string, peer_id string, max_id int64) {
	msg := &proto.Message{
		Type: def.EvReceipt,
		Event: &proto.Event{
			Receipt: &proto.EvReceipt{
				PeerId: peer_id,
				MaxId:  max_id,
			},
		},
	}
	if def.ToIdType(peer_id) == def.IdChannel {
		msg.ToId = peer_id
	}
	cmd := &proto.Command{
		Op:      def.OpSend,
		UserIds: user_ids,
		Message: msg,
	}
	PushCommand(cmd)
}

// 推送修改成员信息事件
func PushEvMemberName(chat_id string, member_id string, name string) {
	msg := &proto.Message{
		Type: def.EvMemberName,
		Event: &proto.Event{
			MemberName: &proto.EvMemberName{
				ChatId: chat_id,
				UserId: member_id,
				Name:   name,
			},
		},
	}
	var user_ids []string
	if def.ToIdType(chat_id) == def.IdChannel {
		msg.ToId = chat_id
	} else {
		user_ids = GetChatMemberIds(chat_id)
	}
	cmd := &proto.Command{
		Op:      def.OpSend,
		UserIds: user_ids,
		Message: msg,
	}
	PushCommand(cmd)
}

// 多端同步事件: 推送删除对话
func PushEvDialogDeleted(self_id, peer_id string, clear bool, max_id int64) {
	msg := &proto.Message{
		Type: def.EvDialogDeleted,
		Event: &proto.Event{
			DialogDeleted: &proto.EvDialogDeleted{
				PeerId: peer_id,
				Clear:  clear,
				MaxId:  max_id,
			},
		},
	}
	cmd := &proto.Command{
		Op:      def.OpSend,
		UserIds: []string{self_id},
		Message: msg,
	}
	PushCommand(cmd)
}

// 推送单人离线事件
func PushUserOfflineEvent(self_id, peer_id string, msg_type def.MsgType, ev *proto.Event) {
	if ev.Seq == 0 {
		ev.Seq = nextSeq()
	}
	msg := &proto.Message{
		Type:  msg_type,
		Event: ev,
	}
	// 发送在线事件
	cmd := &proto.Command{
		Op:      def.OpSend,
		UserIds: []string{self_id},
		Message: msg,
	}
	PushCommand(cmd)
	// 保存redis
	cache.UserSeq(self_id, peer_id).Set(ev.Seq)
	// 保存离线事件
	// 优化:如果用户只有一个token, 就不用保存离线事件, 因为不需要多端同步
	if cache.TokenCount(self_id) > 1 {
		InsertEvent(self_id, peer_id, ev.Seq, msg)
	}
}

// 推送超级群离线事件
func PushChannelOfflineEvent(chat_id string, msg_type def.MsgType, ev *proto.Event) {
	if def.ToIdType(chat_id) != def.IdChannel {
		return
	}
	if ev.Seq == 0 {
		ev.Seq = nextSeq()
	}
	msg := &proto.Message{
		Type:  msg_type,
		ToId:  chat_id,
		Event: ev,
	}
	// 发送在线事件
	cmd := &proto.Command{
		Op:      def.OpSend,
		Message: msg,
	}
	PushCommand(cmd)
	// 保存redis
	cache.ChannelSeq(chat_id).Set(ev.Seq)
	// 保存离线事件
	InsertEvent(chat_id, "", ev.Seq, msg)
}

// 多端同步事件: 推送置顶对话
func PushEvPinDialog(user_id, peer_id string, pinned bool) {
	msg := &proto.Message{
		Type: def.EvPinDialog,
		Event: &proto.Event{
			PinDialog: &proto.EvPinDialog{
				PeerId: peer_id,
				Pinned: pinned,
			},
		},
	}
	cmd := &proto.Command{
		Op:      def.OpSend,
		UserIds: []string{user_id},
		Message: msg,
	}
	PushCommand(cmd)
}

// 多端同步事件: 推送更新对话免打扰设置
func PushEvUpdatePeerNotify(user_id, peer_id string, s *PeerNotifySetting) {
	msg := &proto.Message{
		Type: def.EvUpdatePeerNotify,
		Event: &proto.Event{
			PeerNotify: &proto.EvUpdatePeerNotify{
				PeerId:        peer_id,
				NotifySetting: s.TL(),
			},
		},
	}
	cmd := &proto.Command{
		Op:      def.OpSend,
		UserIds: []string{user_id},
		Message: msg,
	}
	PushCommand(cmd)
}
