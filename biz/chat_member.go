package biz

import (
	"time"

	"zim.cn/base/redis"

	"zim.cn/biz/proto"

	"zim.cn/base/db"
	"zim.cn/biz/cache"
	"zim.cn/biz/def"
)

type MemberRaw struct {
	ChatId    string       `json:"chat_id"`
	UserId    string       `json:"user_id"`
	Name      string       `json:"name"`
	Role      def.ChatRole `json:"role"`
	Muted     bool         `json:"muted"`
	UpdatedAt int          `json:"updated_at"`
}

func loadMember(chat_id, user_id string) *MemberRaw {
	var r = &MemberRaw{}
	ok := db.Replica.QueryRow(`select name, role, muted, updated_at from chat_member where chat_id=? and user_id=? and deleted=0`,
		chat_id, user_id).Scan(&r.Name, &r.Role, &r.Muted, &r.UpdatedAt)
	if !ok {
		return nil
	}
	r.ChatId = chat_id
	r.UserId = user_id
	return r
}

func GetMember(chat_id, user_id string) *MemberRaw {
	key := cache.MemberKey(chat_id, user_id)
	r := key.Get()
	if r.Reply != nil {
		var m *MemberRaw
		err := r.Unmarshal(&m)
		if err == nil {
			return m
		}
	}
	m := loadMember(chat_id, user_id)
	key.Set(m)
	return m
}

func MemberRankScore(role def.ChatRole, updated_at int) float64 {
	return float64(int(role)*1000000000 + updated_at)
}

func LoadMembers(chat_id string) []*MemberRaw {
	var out []*MemberRaw
	rows := db.Replica.Query(`select user_id, name, role, muted, updated_at from chat_member where chat_id=? and deleted=0`, chat_id)
	defer rows.Close()
	for rows.Next() {
		var r = &MemberRaw{}
		rows.Scan(&r.UserId, &r.Name, &r.Role, &r.Muted, &r.UpdatedAt)
		r.ChatId = chat_id
		out = append(out, r)
	}
	return out
}

func GetChatMemberIds(chat_id string) []string {
	key := cache.ChatMemberZset(chat_id)
	member_ids := key.Do("ZREVRANGE", 0, -1).Strings()
	if len(member_ids) > 0 {
		return member_ids
	}
	if !key.Exists() {
		members := LoadMembers(chat_id)
		for _, m := range members {
			score := MemberRankScore(m.Role, m.UpdatedAt)
			key.ZAdd(m.UserId, score)
			member_ids = append(member_ids, m.UserId)
		}
	}
	return member_ids
}

func insertMember(chat_id string, user_ids []string, role def.ChatRole) {
	now := int(time.Now().Unix())
	tx := db.Primary.Begin()
	defer tx.Rollback()
	for _, user_id := range user_ids {
		ok := tx.Exec(`insert ignore into chat_member set chat_id=?, user_id=?, role=?, muted=0, deleted=0, created_at=?, updated_at=?`,
			chat_id, user_id, role, now, now).OK()
		if !ok {
			tx.Exec(`update chat_member set role=?, muted=0, deleted=0, updated_at=? where chat_id=? and user_id=?`,
				role, now, chat_id, user_id)
		}
	}
	tx.Commit()
	// cache
	score := MemberRankScore(role, now)
	ct := def.ToIdType(chat_id)
	c := redis.Begin("speed")
	for _, user_id := range user_ids {
		cache.MemberKey(chat_id, user_id).Tx(c).Del()
		if ct == def.IdChannel {
			cache.ChannelsOfUserKey(user_id).Tx(c).Del()
		}
	}
	redis.Commit(c)
	c = redis.Begin("cache")
	for _, user_id := range user_ids {
		cache.ChatMemberZset(chat_id).Tx(c).ZAdd(user_id, score)
	}
	redis.Commit(c)
}

func deleteMember(chat_id, user_id string) bool {
	ok := db.Primary.Exec(`update chat_member set muted=0, deleted=1 where chat_id=? and user_id=?`,
		chat_id, user_id).OK()
	if ok {
		cache.MemberKey(chat_id, user_id).Set(nil)
		cache.ChatMemberZset(chat_id).ZRem(user_id)
		ct := def.ToIdType(chat_id)
		if ct == def.IdChannel {
			cache.ChannelsOfUserKey(user_id).Del()
		}
	}
	return ok
}

func updateMemberMuted(chat_id, user_id string, muted bool) bool {
	ok := db.Primary.Exec(`update chat_member set muted=? where chat_id=? and user_id=?`,
		muted, chat_id, user_id).OK()
	if ok {
		cache.MemberKey(chat_id, user_id).Del()
	}
	return ok
}

// 成员禁言/解禁
// self_id: 操作者ID
// enable:true禁言 false解禁
func MuteMember(self_id, chat_id, user_id string, muted bool, duration int) bool {
	ok := updateMemberMuted(chat_id, user_id, muted)
	if !ok {
		return false
	}
	tluser := GetTinyUser(user_id)
	tip := &proto.TipMemberMuted{
		ChatId:   chat_id,
		Enable:   muted,
		User:     tluser,
		Duration: duration,
	}
	msg := &proto.Message{
		FromId: self_id,
		ToId:   chat_id,
		Type:   def.TipMemberMuted,
		Tip:    &proto.Tip{MemberMuted: tip},
	}
	SendMessage(msg)
	return true
}

func updateMemberRole(chat_id, user_id string, role int) bool {
	ok := db.Primary.Exec(`update chat_member set role=? where chat_id=? and user_id=?`,
		role, chat_id, user_id).OK()
	if ok {
		cache.MemberKey(chat_id, user_id).Del()
	}
	return ok
}

// 修改成员角色
func EditRole(self_id, chat_id, user_id string, role int) bool {
	ok := updateMemberRole(chat_id, user_id, role)
	if !ok {
		return ok
	}
	tluser := GetTinyUser(user_id)
	tip := &proto.TipEditRole{
		ChatId: chat_id,
		User:   tluser,
		Role:   role,
	}
	msg := &proto.Message{
		FromId: self_id,
		ToId:   chat_id,
		Type:   def.TipEditRole,
		Tip:    &proto.Tip{EditRole: tip},
	}
	SendMessage(msg)
	return true
}

func updateMemberName(chat_id, user_id string, name string) bool {
	ok := db.Primary.Exec(`update chat_member set name=? where chat_id=? and user_id=?`,
		name, chat_id, user_id).OK()
	if ok {
		cache.MemberKey(chat_id, user_id).Del()
	}
	return ok
}

// 修改成员昵称
func EditMemberName(chat_id, user_id string, name string) bool {
	ok := updateMemberName(chat_id, user_id, name)
	if !ok {
		return ok
	}
	PushEvMemberName(chat_id, user_id, name)
	return true
}

// 成员列表, 按入群时间排序
func loadMemberSlice(chat_id string, offset, limit int) []*MemberRaw {
	var out []*MemberRaw
	rows := db.Replica.Query(`select user_id, name, role, muted, updated_at from chat_member 
        where chat_id=? and deleted=0 limit ?,?`, chat_id, offset, limit)
	defer rows.Close()
	for rows.Next() {
		var r = &MemberRaw{}
		rows.Scan(&r.UserId, &r.Name, &r.Role, &r.Muted, &r.UpdatedAt)
		r.ChatId = chat_id
		out = append(out, r)
	}
	return out
}

// 成员列表API
func GetTLMembers(chat_id string, offset, limit int) []*proto.Member {
	var out = []*proto.Member{}
	raws := loadMemberSlice(chat_id, offset, limit)
	var user_ids []string
	for _, r := range raws {
		user_ids = append(user_ids, r.UserId)
	}
	userm := GetUserMap(user_ids)
	for _, r := range raws {
		tl := &proto.Member{
			Name:  r.Name,
			Muted: r.Muted,
			Role:  int(r.Role),
		}
		user := userm[r.UserId]
		if user != nil {
			tl.User = user.TL()
		}
		out = append(out, tl)
	}
	return out
}
