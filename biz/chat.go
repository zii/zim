package biz

import (
	"encoding/json"
	"strings"
	"time"

	"zim.cn/base/redis"

	"zim.cn/biz/cache"

	"zim.cn/biz/proto"

	"zim.cn/base/db"
	"zim.cn/biz/def"
)

type ChatRaw struct {
	ChatId  string `json:"chat_id"`
	OwnerId string `json:"owner"`
	Type    int    `json:"type"`
	Title   string `json:"title"`
	About   string `json:"about"`
	Photo   string `json:"photo"`
	Maxp    int    `json:"maxp"`
	Muted   bool   `json:"muted"`
	Deleted bool   `json:"deleted"`
}

func (r *ChatRaw) TL() *proto.Chat {
	if r == nil {
		return nil
	}
	o := &proto.Chat{
		Id:      r.ChatId,
		Type:    r.Type,
		Title:   r.Title,
		About:   r.About,
		OwnerId: r.OwnerId,
		Photo:   r.Photo,
		Maxp:    r.Maxp,
		Muted:   r.Muted,
	}
	return o
}

func LoadChat(id string) *ChatRaw {
	var r = &ChatRaw{}
	ok := db.Replica.QueryRow(`select owner_id, type, title, about, photo, maxp, muted, deleted from chat where chat_id=?`,
		id).Scan(&r.OwnerId, &r.Type, &r.Title, &r.About, &r.Photo, &r.Maxp, &r.Muted, &r.Deleted)
	if !ok {
		return nil
	}
	r.ChatId = id
	return r
}

func GetChat(id string) *ChatRaw {
	key := cache.ChatKey(id)
	r := key.Get()
	if r.Reply != nil {
		var out *ChatRaw
		err := r.Unmarshal(&out)
		if err == nil {
			return out
		}
	}
	chat := LoadChat(id)
	key.Set(chat)
	return chat
}

func GetTLChat(id string) *proto.Chat {
	c := GetChat(id)
	return c.TL()
}

func LoadChats(chat_ids []string) []*ChatRaw {
	var out []*ChatRaw
	if len(chat_ids) == 0 {
		return out
	}
	q := `select chat_id, owner_id, type, title, about, photo, maxp, muted, deleted from chat`
	p := db.Prepare()
	p.In("chat_id", chat_ids)
	q += p.Clause()
	rows := db.Replica.Query(q, p.Args()...)
	defer rows.Close()
	for rows.Next() {
		var c = &ChatRaw{}
		rows.Scan(&c.ChatId, &c.OwnerId, &c.Type, &c.Title, &c.About, &c.Photo, &c.Maxp, &c.Muted, &c.Deleted)
		out = append(out, c)
	}
	return out
}

func GetChatMap(chat_ids []string) map[string]*ChatRaw {
	var out = make(map[string]*ChatRaw)
	if len(chat_ids) == 0 {
		return out
	}
	var keys []any
	for _, chat_id := range chat_ids {
		k := cache.ChatKey(chat_id).Key
		keys = append(keys, k)
	}
	vals := redis.Do("speed", "MGET", keys...).Strings()
	var missing []string
	for i, s := range vals {
		chat_id := chat_ids[i]
		if s == "" {
			missing = append(missing, chat_id)
			continue
		}
		var c *ChatRaw
		err := json.Unmarshal([]byte(s), &c)
		if err != nil {
			missing = append(missing, chat_id)
			continue
		}
		out[chat_id] = c
	}
	if len(missing) > 0 {
		chats := LoadChats(missing)
		for _, c := range chats {
			out[c.ChatId] = c
			cache.ChatKey(c.ChatId).Set(c)
		}
	}
	return out
}

func GetChats(chat_ids []string) []*ChatRaw {
	var chats []*ChatRaw
	m := GetChatMap(chat_ids)
	for _, chat_id := range chat_ids {
		c := m[chat_id]
		if c != nil {
			chats = append(chats, c)
		}
	}
	return chats
}

func GetTLChats(chat_ids []string) []*proto.Chat {
	var out = []*proto.Chat{}
	chats := GetChats(chat_ids)
	for _, c := range chats {
		out = append(out, c.TL())
	}
	return out
}

func GetTLChatMap(chat_ids []string) map[string]*proto.Chat {
	var out = make(map[string]*proto.Chat)
	m := GetChatMap(chat_ids)
	for chat_id, raw := range m {
		out[chat_id] = raw.TL()
	}
	return out
}

// 默认群成员上限
func DefaultMemberLimit(typ int) int {
	if typ == def.TypeGroup {
		return 200
	} else {
		return 1000000
	}
}

func CreateChat(owner_id string, typ int, title, about string, init_members []string, maxp int) string {
	var chat_id string
	if typ == def.TypeGroup {
		chat_id = GenerateAccid(def.IdGroup)
	} else if typ == def.TypeChannel {
		chat_id = GenerateAccid(def.IdChannel)
	} else {
		return ""
	}
	now := int(time.Now().Unix())
	tx := db.Primary.Begin()
	defer tx.Rollback()
	tx.Exec(`insert into chat set chat_id=?, owner_id=?, type=?, title=?, about=?, photo=?, maxp=?, muted=0, created_at=?`,
		chat_id, owner_id, typ, title, about, "", maxp, now)
	tx.Exec(`insert into chat_member set chat_id=?, user_id=?, role=?, created_at=?, updated_at=?`,
		chat_id, owner_id, def.RoleOwner, now, now)
	for _, member_id := range init_members {
		tx.Exec(`insert into chat_member set chat_id=?, user_id=?, role=?, created_at=?, updated_at=?`,
			chat_id, member_id, def.RoleMember, now, now)
	}
	tx.Commit()
	// 创建完了发个消息
	tlchat := &proto.Chat{
		Id:      chat_id,
		Type:    typ,
		Title:   title,
		About:   about,
		OwnerId: owner_id,
		Maxp:    maxp,
	}
	owner := GetUser(owner_id)
	tlusers := GetTLUsers(init_members)
	msg := &proto.Message{
		Type:   def.TipChatCreated,
		FromId: owner_id,
		ToId:   chat_id,
		Tip: &proto.Tip{
			ChatCreated: &proto.TipChatCreated{
				Chat:        tlchat,
				Creator:     owner.TL(),
				InitMembers: tlusers,
			},
		},
	}
	key := cache.ChatMemberZset(chat_id)
	score := MemberRankScore(def.RoleOwner, now)
	key.ZAdd(owner_id, score)
	for _, member_id := range init_members {
		score := MemberRankScore(def.RoleMember, now)
		key.ZAdd(member_id, score)
	}
	SendMessage(msg)
	// clear cache
	if typ == def.TypeChannel {
		cache.ChannelsOfUserKey(owner_id).Del()
		for _, member_id := range init_members {
			cache.ChannelsOfUserKey(member_id).Del()
		}
	}
	return chat_id
}

// 添加群成员
func AddMember(from_id, chat_id string, user_ids []string) error {
	role := def.RoleMember
	insertMember(chat_id, user_ids, role)

	from := GetUser(from_id)
	users := GetTLUsers(user_ids)
	msg := &proto.Message{
		Type:   def.TipMemberEnter,
		FromId: from_id,
		ToId:   chat_id,
		Tip: &proto.Tip{
			MemberEnter: &proto.TipMemberEnter{
				ChatId:  chat_id,
				Users:   users,
				Role:    role,
				Inviter: from.TinyTL(),
			},
		},
	}
	return SendMessage(msg)
}

// 移除群成员
func DeleteMember(chat_id, user_id string, kick bool) bool {
	ok := deleteMember(chat_id, user_id)
	if !ok {
		return false
	}
	user := GetTinyUser(user_id)
	msg := &proto.Message{
		FromId: user_id,
		ToId:   chat_id,
		Tip:    &proto.Tip{},
	}
	var cc []string
	if kick {
		msg.Type = def.TipMemberKicked
		msg.Tip.MemberKicked = &proto.TipMemberKicked{
			ChatId: chat_id,
			User:   user,
		}
		cc = append(cc, user_id)
	} else {
		msg.Type = def.TipMemberQuit
		msg.Tip.MemberQuit = &proto.TipMemberQuit{
			ChatId: chat_id,
			User:   user,
		}
	}
	SendMessage(msg, cc...)
	if !kick {
		PushEvQuitChat(user_id, chat_id)
		RemoveDialog(user_id, chat_id)
	}
	return true
}

func updateChatMuted(chat_id string, muted bool) bool {
	ok := db.Primary.Exec(`update chat set muted=? where chat_id=?`, muted, chat_id).OK()
	if ok {
		cache.ChatKey(chat_id).Del()
	}
	return ok
}

// 全员禁言/解禁 enable:true禁言 false解禁
func MuteChat(owner_id string, chat_id string, muted bool, duration int) bool {
	ok := updateChatMuted(chat_id, muted)
	if !ok {
		return false
	}
	tip := &proto.TipChatMuted{
		ChatId:   chat_id,
		Duration: duration,
		Enable:   muted,
	}
	msg := &proto.Message{
		FromId: owner_id,
		ToId:   chat_id,
		Type:   def.TipChatMuted,
		Tip: &proto.Tip{
			ChatMuted: tip,
		},
	}
	SendMessage(msg)
	return true
}

func EditChatInfo(old, nue *ChatRaw) bool {
	q := `update chat set `
	var args []any
	var conds []string
	if old.Title != nue.Title {
		conds = append(conds, "title=?")
		args = append(args, nue.Title)
	}
	// 是否发送群公告消息
	var send_about bool
	if old.About != nue.About {
		conds = append(conds, "about=?")
		args = append(args, nue.About)
		if nue.About != "" {
			send_about = true
		}
	}
	if old.Photo != nue.Photo {
		conds = append(conds, "photo=?")
		args = append(args, nue.Photo)
	}
	if len(conds) == 0 {
		return false
	}
	q += strings.Join(conds, ", ")
	q += " where chat_id=?"
	args = append(args, nue.ChatId)
	ok := db.Primary.Exec(q, args...).OK()
	if ok {
		cache.ChatKey(nue.ChatId).Del()
		if send_about {
			text := "公告:\n" + nue.About
			msg := &proto.Message{
				FromId: nue.OwnerId,
				ToId:   nue.ChatId,
				Elem: &proto.Elem{
					Text: text,
				},
			}
			SendMessage(msg)
		}
	}
	return ok
}

func updateOwner(chat *ChatRaw, user_id string) bool {
	if chat.ChatId == "" {
		return false
	}
	if chat.OwnerId == user_id {
		return false
	}
	if user_id == "" {
		return false
	}
	tx := db.Primary.Begin()
	defer tx.Rollback()
	ok := tx.Exec(`update chat set owner_id=? where chat_id=?`, user_id, chat.ChatId).OK()
	if !ok {
		return false
	}
	tx.Exec(`update chat_member set role=? where chat_id=? and user_id=?`,
		def.RoleMember, chat.ChatId, chat.OwnerId)
	tx.Exec(`update chat_member set role=?, muted=0 where chat_id=? and user_id=?`,
		def.RoleOwner, chat.ChatId, user_id)
	tx.Commit()
	// clear cache
	cache.ChatKey(chat.ChatId).Del()
	cache.MemberKey(chat.ChatId, chat.OwnerId).Del()
	cache.MemberKey(chat.ChatId, user_id).Del()
	return true
}

func TransferOwner(chat *ChatRaw, user_id string) bool {
	ok := updateOwner(chat, user_id)
	if !ok {
		return false
	}
	tip := &proto.TipOwnerTransfer{
		ChatId:   chat.ChatId,
		OldOwner: GetTLUser(chat.OwnerId),
		NewOwner: GetTLUser(user_id),
	}
	msg := &proto.Message{
		FromId: chat.OwnerId,
		ToId:   chat.ChatId,
		Type:   def.TipOwnerTransfer,
		Tip: &proto.Tip{
			OwnerTransfer: tip,
		},
	}
	SendMessage(msg)
	return true
}

func deleteChat(chat_id string) bool {
	ok := db.Primary.Exec(`update chat set deleted=1 where chat_id=?`, chat_id).OK()
	if ok {
		cache.ChatKey(chat_id).Del()
	}
	return ok
}

func DismissChat(chat *ChatRaw) bool {
	if chat.Deleted {
		return false
	}
	ok := deleteChat(chat.ChatId)
	if !ok {
		return false
	}
	tip := &proto.TipDismissed{
		ChatId: chat.ChatId,
	}
	msg := &proto.Message{
		FromId: chat.OwnerId,
		ToId:   chat.ChatId,
		Type:   def.TipDismissed,
		Tip: &proto.Tip{
			Dismissed: tip,
		},
	}
	SendMessage(msg)
	return true
}

// 分页查询用户所在群列表
func SearchUserTLChats(self_id string, offset, limit int) []*proto.Chat {
	var chat_ids []string
	q := `select chat_id from chat_member where user_id=? and deleted=0 order by id desc limit ?, ?`
	rows := db.Replica.Query(q, self_id, offset, limit)
	defer rows.Close()
	for rows.Next() {
		var chat_id string
		rows.Scan(&chat_id)
		chat_ids = append(chat_ids, chat_id)
	}
	out := GetTLChats(chat_ids)
	return out
}

// 批量获取群成员数
func MGetMemberCount(chat_ids []string) map[string]int {
	c := redis.Begin("cache")
	for _, chat_id := range chat_ids {
		cache.ChatMemberZset(chat_id).Tx(c).ZCard()
	}
	r := redis.Commit(c)
	out := make(map[string]int)
	for i, n := range r.Ints() {
		chat_id := chat_ids[i]
		out[chat_id] = n
	}
	return out
}

// 单独获取群成员数
func GetMemberCount(chat_id string) int {
	return cache.ChatMemberZset(chat_id).ZCard()
}
