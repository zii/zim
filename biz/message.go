package biz

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"zim.cn/base/uuid"

	"zim.cn/base/log"

	redigo "github.com/gomodule/redigo/redis"
	"zim.cn/biz/kafka"

	"zim.cn/base/redis"

	"zim.cn/base/db"

	"zim.cn/biz/def"

	"zim.cn/biz/cache"
	"zim.cn/biz/proto"
)

func LoadMessage(msg_id int64) *proto.Message {
	var blob []byte
	db.Replica.Get(&blob, `select msg_blob from message where msg_id=?`, msg_id)
	var out *proto.Message
	json.Unmarshal(blob, &out)
	return out
}

func GetMessage(msg_id int64) *proto.Message {
	key := cache.MessageKey(msg_id)
	r := key.Get()
	if r.Reply != nil {
		var m *proto.Message
		err := r.Unmarshal(&m)
		if err == nil {
			return m
		}
	}
	m := LoadMessage(msg_id)
	key.Set(m)
	return m
}

func loadMessageMap(msg_ids []int64) map[int64]*proto.Message {
	var out = make(map[int64]*proto.Message)
	if len(msg_ids) == 0 {
		return out
	}
	q := fmt.Sprintf("select msg_id, msg_blob from message where msg_id in (%s)", db.JoinArray(msg_ids))
	rows := db.Replica.Query(q)
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

// 获得普通消息字典
func GetMessageMap(msg_ids []int64) map[int64]*proto.Message {
	var out = make(map[int64]*proto.Message)
	if len(msg_ids) == 0 {
		return out
	}
	var keys = make([]any, 0, len(msg_ids))
	for _, msg_id := range msg_ids {
		k := cache.MessageKey(msg_id).Key
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
		m := loadMessageMap(missing)
		c := redis.Begin("speed")
		for msg_id, msg := range m {
			out[msg_id] = msg
			cache.MessageKey(msg_id).Tx(c).Set(msg)
		}
		redis.Commit(c)
	}
	return out
}

// 获得普通消息列表
func GetMessages(msg_ids []int64) []*proto.Message {
	var out = []*proto.Message{}
	d := GetMessageMap(msg_ids)
	for _, id := range msg_ids {
		m := d[id]
		if m != nil {
			out = append(out, m)
		}
	}
	return out
}

func UpdateUserMessage(msg *proto.Message) bool {
	if msg == nil {
		return false
	}
	if msg.Id == 0 {
		return false
	}
	blob, err := json.Marshal(msg)
	if err != nil {
		log.Error("Marshal msg:", err)
		return false
	}
	ok := db.Primary.Exec(`update message set msg_blob=? where msg_id=?`, blob, msg.Id).OK()
	if ok {
		cache.MessageKey(msg.Id).Set(msg)
	}
	return ok
}

func delUserMessages(user_id string, msg_ids []int64) bool {
	if len(msg_ids) == 0 {
		return false
	}
	q := `delete from user_msgbox`
	p := db.Prepare()
	p.And("user_id=?", user_id)
	p.In("msg_id", msg_ids)
	q += p.Where()
	ok := db.Primary.Exec(q, p.Args()...).OK()
	// clear cache
	if ok {
		var keys []any
		for _, msg_id := range msg_ids {
			k := cache.MessageKey(msg_id).Key
			keys = append(keys, k)
		}
		redis.Do("speed", "del", keys...)
	}
	return ok
}

func nextMessageId() int64 {
	if def.UseMultiDC {
		return uuid.NextID("msg")
	}
	return cache.PtsKey().Incr().Int64()
}

func nextSeq() int64 {
	if def.UseMultiDC {
		return uuid.NextID("seq")
	}
	return cache.SeqKey().Incr().Int64()
}

func VerifyMessage(msg *proto.Message) error {
	if msg.ToId == "" {
		return errors.New("TO_ID_EMPTY")
	}
	if msg.Type == def.MsgRevoke {
		if msg.Elem == nil || msg.Elem.Revoke == nil {
			return errors.New("REVOKE_ELEM_EMPTY")
		}
		var src *proto.Message
		// 只有发送者本人能撤销
		if def.ToIdType(msg.ToId) != def.IdChannel {
			src = GetMessage(msg.Elem.Revoke.MsgId)
		} else {
			src = GetChannelMessage(msg.ToId, msg.Elem.Revoke.MsgId)
		}
		if src == nil {
			return errors.New("SRC_DELETED")
		}
		if src.FromId != msg.FromId || src.ToId != msg.ToId {
			return errors.New("ACCESS_DENIED")
		}
	}
	tot := def.ToIdType(msg.ToId)
	if tot != def.IdUser {
		chat := GetChat(msg.ToId)
		if chat == nil {
			return errors.New("CHAT_INVALID")
		}
		if chat.Deleted {
			return errors.New("CHAT_DISMISSED")
		}
		if chat.Muted {
			return errors.New("CHAT_MUTED")
		}
		fromm := GetMember(msg.ToId, msg.FromId)
		if fromm == nil {
			return errors.New("NOT_MEMBER")
		}
		if fromm.Muted {
			return errors.New("MUTED")
		}
	}
	return nil
}

// cc: 抄送给额外用户
func SendMessage(msg *proto.Message, cc ...string) error {
	if msg == nil {
		return errors.New("MSG_NIL")
	}
	ct := msg.Type.Class()
	if ct == def.Msg && msg.Elem == nil {
		return errors.New("ELEM_NIL")
	} else if ct == def.Tip && msg.Tip == nil {
		return errors.New("TIP_NIL")
	} else if ct == def.Event && msg.Event == nil {
		return errors.New("EVENT_NIL")
	}
	if msg.FromUser == nil && def.ToIdType(msg.FromId) != def.IdSys {
		fu := GetUser(msg.FromId)
		if fu == nil {
			return errors.New("FROM_ID_INVALID")
		}
		if fu.Banned() {
			return errors.New("FROM_USER_BANNED")
		}
		msg.FromUser = fu.TL()
	}
	if msg.CreatedAt == 0 {
		msg.CreatedAt = int(time.Now().Unix())
	}
	cmd := &proto.Command{
		Op:      def.OpSend,
		Message: msg,
	}
	tot := def.ToIdType(msg.ToId)
	if tot == def.IdUser {
		msg.Id = nextMessageId()
		if def.ToIdType(msg.FromId) != def.IdSys {
			cmd.UserIds = append(cmd.UserIds, msg.FromId)
		}
		if msg.ToId != msg.FromId {
			cmd.UserIds = append(cmd.UserIds, msg.ToId)
		}
		cmd.UserIds = append(cmd.UserIds, cc...)
		PushCommand(cmd)
	} else if tot == def.IdGroup {
		msg.Id = nextMessageId()
		member_ids := GetChatMemberIds(msg.ToId)
		cmd.UserIds = member_ids
		cmd.UserIds = append(cmd.UserIds, cc...)
		PushCommand(cmd)
	} else if tot == def.IdChannel {
		msg.Id = nextChannelMessageId(msg.ToId)
		cmd.UserIds = append(cmd.UserIds, cc...)
		PushCommand(cmd)
	} else {
		log.Warn("INVALID TOT:", tot, msg.ToId)
		return errors.New("INVALID_TOTYPE")
	}

	// 特殊处理: 撤销消息
	if msg.Type == def.MsgRevoke && msg.Elem != nil && msg.Elem.Revoke != nil {
		ok := RevokeMessage(msg.ToId, msg.Elem.Revoke.MsgId)
		if !ok {
			return nil
		}
	}

	// 持久化消息
	if msg.Type < def.Event {
		// if kafka is disabled, execute in current process
		if kafka.P1 == nil {
			go FlushMessage(cmd)
		} else {
			var key string
			if strings.Compare(msg.FromId, msg.ToId) > 0 {
				key = fmt.Sprintf("%s:%s", msg.FromId, msg.ToId)
			} else {
				key = fmt.Sprintf("%s:%s", msg.ToId, msg.FromId)
			}
			data, _ := json.Marshal(cmd)
			err := kafka.P1.SendMessage(key, data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 标记对话消息已读 返回(ok, 对话最新消息ID, 剩余未读数)
// ok: false表示无任何改变
func readHistory(user_id, peer_id string, max_id int64) (bool, int64, int) {
	pts := GetDialogPts(user_id, peer_id)
	if max_id > pts || max_id <= 0 {
		max_id = pts
	}
	ok := cache.DialogRead(user_id, peer_id).LuaSetHi(max_id).Bool()
	pt := def.ToIdType(peer_id)
	if pt == def.IdChannel {
		unread := int(pts - max_id)
		if unread < 0 {
			unread = 0
		}
		return ok, pts, unread
	}
	if max_id >= pts {
		cache.DialogUnread(user_id, peer_id).Set(0)
		return ok, pts, 0
	} else {
		unread := queryDialogUnread(user_id, peer_id, max_id)
		cache.DialogUnread(user_id, peer_id).Set(unread)
		return ok, pts, unread
	}
}

// 快速标记, 未读数直接清零
func FastReadHistoryTx(tx redigo.Conn, user_id, peer_id string, max_id int64) (int64, int) {
	if def.ToIdType(user_id) != def.IdUser {
		return 0, 0
	}
	cache.DialogRead(user_id, peer_id).Tx(tx).Set(max_id)
	pt := def.ToIdType(peer_id)
	if pt == def.IdChannel {
		return max_id, 0
	}
	cache.DialogUnread(user_id, peer_id).Tx(tx).Set(0)
	return max_id, 0
}

// 发送已读回执
func SendReceipt(user_id, to_id string, max_id int64) bool {
	pt := def.ToIdType(to_id)
	if pt == def.IdUser {
		ok := cache.UserReceiptMaxId(to_id, user_id).LuaSetHi(max_id).Bool()
		if ok {
			PushEvReceipt([]string{to_id}, user_id, max_id)
		}
	} else if pt == def.IdGroup {
		chat_id := to_id
		ok := cache.ChatReceiptMaxId(chat_id).LuaSetHi(max_id).Bool()
		if ok {
			member_ids := GetChatMemberIds(chat_id)
			PushEvReceipt(member_ids, chat_id, max_id)
		}
	} else if pt == def.IdChannel {
		chat_id := to_id
		ok := cache.ChatReceiptMaxId(chat_id).LuaSetHi(max_id).Bool()
		if ok {
			PushEvReceipt(nil, chat_id, max_id)
		}
	}
	return true
}

func ReadHistory(user_id, peer_id string, max_id int64) (int64, int) {
	ok, pts, unread := readHistory(user_id, peer_id, max_id)
	if ok {
		if max_id <= 0 || max_id > pts {
			max_id = pts
		}
		PushEvHasRead(user_id, peer_id, pts, unread)
		SendReceipt(user_id, peer_id, max_id)
	}
	return pts, unread
}

type InputUserMessage struct {
	UserIds []string
	Msg     *proto.Message
}

// 插入用户信箱
func InsertUserMessage(input *InputUserMessage) {
	user_ids := input.UserIds
	msg := input.Msg
	if len(user_ids) == 0 {
		return
	}
	blob, err := json.Marshal(msg)
	if err != nil {
		log.Error("encode msg:", err)
		return
	}
	now := int(time.Now().Unix())
	ok := db.Primary.Exec(`insert ignore into message set msg_id=?, msg_blob=?, from_id=?, to_id=?, created_at=?`,
		msg.Id, blob, msg.FromId, msg.ToId, now).OK()
	if !ok {
		return
	}
	var args []any
	q := `insert into user_msgbox(user_id, peer_id, msg_id, from_id, to_id, created_at) values`
	for _, user_id := range user_ids {
		q += `(?,?,?,?,?,?),`
		args = append(args, user_id, msg.GetPeerId(user_id), msg.Id, msg.FromId, msg.ToId, now)
	}
	q = q[:len(q)-1]
	db.Primary.Exec(q, args...)
}

// 批量插入用户消息
func BulkdInsertUserMessage(inputs []*InputUserMessage) {
	now := int(time.Now().Unix())
	{
		var args []any
		q := `insert ignore into message(msg_id, msg_blob, from_id, to_id, created_at) values`
		for _, in := range inputs {
			msg := in.Msg
			blob, err := json.Marshal(msg)
			if err != nil {
				log.Error("encode msg:", err)
				continue
			}
			q += `(?,?,?,?,?),`
			args = append(args, msg.Id, blob, msg.FromId, msg.ToId, now)
		}
		if len(args) > 0 {
			q = q[:len(q)-1]
			db.Primary.Exec(q, args...)
		}
	}
	{
		var args []any
		q := `insert ignore into user_msgbox(user_id, peer_id, msg_id, from_id, to_id, created_at) values`
		for _, in := range inputs {
			msg := in.Msg
			for _, user_id := range in.UserIds {
				q += `(?,?,?,?,?,?),`
				args = append(args, user_id, msg.GetPeerId(user_id), msg.Id, msg.FromId, msg.ToId, now)
			}
		}
		if len(args) > 0 {
			q = q[:len(q)-1]
			db.Primary.Exec(q, args...)
		}
	}
}

// 插入超级群消息
func InsertChannelMessage(msg *proto.Message) {
	blob, err := json.Marshal(msg)
	if err != nil {
		log.Error("encode msg:", err)
		return
	}
	if def.ToIdType(msg.ToId) != def.IdChannel {
		return
	}
	now := int(time.Now().Unix())
	db.Primary.Exec(`insert into channel_msgbox set chat_id=?, msg_id=?, msg_blob=?, from_id=?, created_at=?`,
		msg.ToId, msg.Id, blob, msg.FromId, now)
}

// 批量插入超级群消息
func BulkInserChannelMessage(msgs []*proto.Message) {
	var q = `insert into channel_msgbox(chat_id, msg_id, msg_blob, from_id, created_at) values`
	var args []any
	now := int(time.Now().Unix())
	for _, msg := range msgs {
		if def.ToIdType(msg.ToId) != def.IdChannel {
			continue
		}
		blob, err := json.Marshal(msg)
		if err != nil {
			log.Error("encode msg:", err)
			continue
		}
		q += `(?,?,?,?,?),`
		args = append(args, msg.ToId, msg.Id, blob, msg.FromId, now)
	}
	if len(args) > 0 {
		q = q[:len(q)-1]
		db.Primary.Exec(q, args...)
	}
}

// 撤销用户消息
func RevokeUserMessage(msg_id int64) bool {
	msg := GetMessage(msg_id)
	if msg == nil {
		return false
	}
	if msg.Revoked {
		return false
	}
	msg.Id = msg_id
	msg.Revoked = true
	return UpdateUserMessage(msg)
}

// 撤消超级群消息
func RevokeChannelMessage(chat_id string, msg_id int64) bool {
	msg := GetChannelMessage(chat_id, msg_id)
	if msg == nil {
		return false
	}
	if msg.Revoked {
		return false
	}
	msg.Id = msg_id
	msg.Revoked = true
	return UpdateChannelMessage(msg)
}

func RevokeMessage(to_id string, msg_id int64) bool {
	tot := def.ToIdType(to_id)
	if tot == def.IdChannel {
		return RevokeChannelMessage(to_id, msg_id)
	}
	return RevokeUserMessage(msg_id)
}

// 加载历史消息 区间=(min_id, max_id)
func LoadHistory(user_id, peer_id string, min_id, max_id int64, offset, limit int) []*proto.Message {
	peert := def.ToIdType(peer_id)
	var rows *db.MustRows
	if peert == def.IdChannel {
		q := `select msg_blob from channel_msgbox`
		p := db.Prepare()
		p.And("chat_id=? and msg_id>?", peer_id, min_id)
		if max_id > 0 {
			p.And("msg_id<?", max_id)
		}
		p.Sort("msg_id desc")
		p.Slice(offset, limit)
		q += p.Clause()
		rows = db.Replica.Query(q, p.Args()...)
	} else {
		q := `select m.msg_blob from user_msgbox b inner join message m on m.msg_id=b.msg_id`
		p := db.Prepare()
		p.And("b.user_id=? and b.peer_id=? and b.msg_id>?", user_id, peer_id, min_id)
		if max_id > 0 {
			p.And("b.msg_id<?", max_id)
		}
		p.Sort("b.msg_id desc")
		p.Slice(offset, limit)
		q += p.Clause()
		rows = db.Replica.Query(q, p.Args()...)
	}
	defer rows.Close()
	var out = []*proto.Message{}
	for rows.Next() {
		var blob []byte
		rows.Scan(&blob)
		var msg *proto.Message
		json.Unmarshal(blob, &msg)
		out = append(out, msg)
	}
	return out
}

// 单向删除用户自己的消息
func DeleteUserMessages(self_id, peer_id string, msg_ids []int64) bool {
	ok := delUserMessages(self_id, msg_ids)
	if !ok {
		return false
	}
	ev := &proto.Event{
		MsgDeleted: &proto.EvMsgDeleted{
			PeerId: peer_id,
			MsgId:  msg_ids,
		},
	}
	PushUserOfflineEvent(self_id, peer_id, def.EvMsgDeleted, ev)
	return true
}

// 删除超级群消息
func DeleteChannelMessages(chat_id string, msg_ids []int64) bool {
	ok := delChannelMessages(chat_id, msg_ids)
	if !ok {
		return false
	}
	ev := &proto.Event{
		MsgDeleted: &proto.EvMsgDeleted{
			PeerId: chat_id,
			MsgId:  msg_ids,
		},
	}
	PushChannelOfflineEvent(chat_id, def.EvMsgDeleted, ev)
	return true
}

// 是不是可转发的消息
func Forwardable(msg *proto.Message) bool {
	if msg.Type.Class() != def.Msg {
		return false
	}
	if msg.Type == def.MsgRevoke {
		return false
	}
	return true
}

// 逐条转发多个消息
func ForwardMessages(self_id string, to_ids []string, peer_id string, msg_ids []int64) error {
	pt := def.ToIdType(peer_id)
	var msgs []*proto.Message
	if pt == def.IdChannel {
		msgs = GetChannelMessages(peer_id, msg_ids)
	} else {
		msgs = GetMessages(msg_ids)
	}
	if len(msgs) == 0 {
		return nil
	}
	for _, m := range msgs {
		if !Forwardable(m) {
			return fmt.Errorf("不可转发的消息类型:%s:%d:%d", peer_id, m.Id, m.Type)
		}
		if m.FwdHeader == nil {
			m.FwdHeader = &proto.FwdHeader{
				FromId: m.FromId,
			}
			if m.FromUser != nil {
				m.FwdHeader.FromName = m.FromUser.Name
			}
			if def.ToIdType(m.ToId) != def.IdUser {
				m.FwdHeader.ChatId = m.ToId
			}
			m.FwdHeader.MsgId = m.Id
		}
		m.FromId = self_id
		for _, to_id := range to_ids {
			m.ToId = to_id
			err := SendMessage(m)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 发送聊天记录(合并消息)
func SendChatlog(self_id string, to_ids []string, peer_id string, msg_ids []int64) error {
	elem := &proto.Elem{
		ChatLog: &proto.ChatLogElem{
			Title: "",
		},
	}
	pt := def.ToIdType(peer_id)
	if pt == def.IdChannel {
		elem.ChatLog.Msgs = GetChannelMessages(peer_id, msg_ids)
	} else {
		elem.ChatLog.Msgs = GetMessages(msg_ids)
	}
	msg := &proto.Message{
		Type:   def.MsgChatlog,
		FromId: self_id,
		Elem:   elem,
	}
	for _, to_id := range to_ids {
		msg.ToId = to_id
		err := SendMessage(msg)
		if err != nil {
			return err
		}
	}
	return nil
}
