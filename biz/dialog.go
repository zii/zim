package biz

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	"zim.cn/base"

	"zim.cn/base/db"
	"zim.cn/base/redis"
	"zim.cn/biz/cache"
	"zim.cn/biz/def"
	"zim.cn/biz/proto"
)

// 对话是否存在
func IsDialogExists(user_id, peer_id string) bool {
	return cache.UserDialogZset(user_id).ZExists(peer_id)
}

// 对话中我已读的最大消息ID
func GetDialogReadId(user_id, peer_id string) int64 {
	return cache.DialogRead(user_id, peer_id).Get().Int64()
}

func loadDialogPts(user_id, peer_id string) int64 {
	var pts int64
	peert := def.ToIdType(peer_id)
	if peert == def.IdChannel {
		db.Replica.Get(&pts, `select coalesce(max(msg_id), 0) from channel_msgbox where chat_id=?`, peer_id)
	} else {
		db.Replica.Get(&pts, `select coalesce(max(msg_id), 0) from user_msgbox where user_id=? and peer_id=?`,
			user_id, peer_id)
	}
	return pts
}

// 查询对话最大消息ID
func GetDialogPts(user_id, peer_id string) int64 {
	pt := def.ToIdType(peer_id)
	if pt == def.IdChannel {
		return GetChannelPts(peer_id)
	}
	key := cache.DialogPts(user_id, peer_id)
	r := key.Get()
	if r.Reply != nil {
		return r.Int64()
	}
	pts := loadDialogPts(user_id, peer_id)
	key.Set(pts)
	return pts
}

func loadUserDialogPtsMap(user_id string, peer_ids []string) map[string]int64 {
	var out = make(map[string]int64)
	if len(peer_ids) == 0 {
		return out
	}
	q := `select peer_id, max(msg_id) from user_msgbox`
	p := db.Prepare()
	p.And("user_id=?", user_id)
	p.In("peer_id", peer_ids)
	q += p.Where() + ` group by peer_id`
	rows := db.Replica.Query(q, p.Args()...)
	defer rows.Close()
	for rows.Next() {
		var peer_id string
		var pts int64
		rows.Scan(&peer_id, &pts)
		out[peer_id] = pts
	}
	return out
}

func GetUserDialogPtsMap(self_id string, peer_ids []string) map[string]int64 {
	var out = make(map[string]int64)
	if len(peer_ids) == 0 {
		return out
	}
	var keys []any
	for _, peer_id := range peer_ids {
		k := cache.DialogPts(self_id, peer_id).Key
		keys = append(keys, k)
	}
	vals := redis.Do("cache", "MGET", keys...).Strings()
	var missing []string
	for i, s := range vals {
		peer_id := peer_ids[i]
		if s == "" {
			missing = append(missing, peer_id)
			continue
		}
		pts, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			missing = append(missing, peer_id)
			continue
		}
		out[peer_id] = pts
	}
	if len(missing) > 0 {
		m := loadUserDialogPtsMap(self_id, missing)
		c := redis.Begin("cache")
		for peer_id, pts := range m {
			out[peer_id] = pts
			cache.DialogPts(self_id, peer_id).Tx(c).Set(pts)
		}
		redis.Commit(c)
	}
	return out
}

func GetDialogPtsMap(self_id string, peer_ids []string) map[string]int64 {
	if len(peer_ids) == 0 {
		return make(map[string]int64)
	}
	var user_ids []string
	var chat_ids []string
	for _, peer_id := range peer_ids {
		if def.ToIdType(peer_id) == def.IdChannel {
			chat_ids = append(chat_ids, peer_id)
		} else {
			user_ids = append(user_ids, peer_id)
		}
	}
	um := GetUserDialogPtsMap(self_id, user_ids)
	cm := GetChannelPtsMap(chat_ids)
	for chat_id, pts := range cm {
		um[chat_id] = pts
	}
	return um
}

// 查询剩余未读数
func queryDialogUnread(user_id, peer_id string, min_id int64) int {
	var n int
	peert := def.ToIdType(peer_id)
	if peert == def.IdChannel {
		db.Replica.Get(&n, `select count(*) from channel_msgbox where chat_id=? and msg_id>?`, peer_id, min_id)
	} else {
		db.Replica.Get(&n, `select count(*) from user_msgbox where user_id=? and peer_id=? and msg_id>?`,
			user_id, peer_id, min_id)
	}
	return n
}

func loadDialogUnread(user_id, peer_id string) int {
	min_id := GetDialogReadId(user_id, peer_id)
	n := queryDialogUnread(user_id, peer_id, min_id)
	return n
}

func GetUserDialogUnread(user_id, peer_id string) int {
	key := cache.DialogUnread(user_id, peer_id)
	r := key.Get()
	if r.Reply != nil {
		return r.Int()
	}
	n := loadDialogUnread(user_id, peer_id)
	key.Set(n)
	return n
}

func GetUserDialogUnreadMap(user_id string, peer_ids []string) map[string]int {
	var out = make(map[string]int)
	if len(peer_ids) == 0 {
		return out
	}
	var keys []any
	for _, peer_id := range peer_ids {
		k := cache.DialogUnread(user_id, peer_id).Key
		keys = append(keys, k)
	}
	vals := redis.Do("cache", "MGET", keys...).Strings()
	var missing []string
	for i, s := range vals {
		peer_id := peer_ids[i]
		if s == "" {
			missing = append(missing, peer_id)
			continue
		}
		out[peer_id], _ = strconv.Atoi(s)
	}
	for _, peer_id := range missing {
		n := GetUserDialogUnread(user_id, peer_id)
		cache.DialogUnread(user_id, peer_id).Set(n)
		out[peer_id] = n
	}
	return out
}

func GetUserSeqMap(self_id string, peer_ids []string) map[string]int64 {
	var out = make(map[string]int64)
	if len(peer_ids) == 0 {
		return out
	}
	var keys []any
	for _, peer_id := range peer_ids {
		k := cache.UserSeq(self_id, peer_id).Key
		keys = append(keys, k)
	}
	vals := redis.Do("cache", "MGET", keys...).Int64s()
	for i, seq := range vals {
		peer_id := peer_ids[i]
		out[peer_id] = seq
	}
	return out
}

func GetChannelSeqMap(chat_ids []string) map[string]int64 {
	var out = make(map[string]int64)
	if len(chat_ids) == 0 {
		return out
	}
	var keys []any
	for _, chat_id := range chat_ids {
		k := cache.ChannelSeq(chat_id).Key
		keys = append(keys, k)
	}
	vals := redis.Do("cache", "MGET", keys...).Int64s()
	for i, seq := range vals {
		chat_id := chat_ids[i]
		out[chat_id] = seq
	}
	return out
}

func GetUserReceiptMap(self_id string, peer_ids []string) map[string]int64 {
	var out = make(map[string]int64)
	if len(peer_ids) == 0 {
		return out
	}
	var keys []any
	for _, peer_id := range peer_ids {
		k := cache.UserReceiptMaxId(self_id, peer_id).Key
		keys = append(keys, k)
	}
	vals := redis.Do("cache", "MGET", keys...).Int64s()
	for i, max_id := range vals {
		peer_id := peer_ids[i]
		out[peer_id] = max_id
	}
	return out
}

func GetChatReceiptMap(chat_ids []string) map[string]int64 {
	var out = make(map[string]int64)
	if len(chat_ids) == 0 {
		return out
	}
	var keys []any
	for _, chat_id := range chat_ids {
		k := cache.ChatReceiptMaxId(chat_id).Key
		keys = append(keys, k)
	}
	vals := redis.Do("cache", "MGET", keys...).Int64s()
	for i, max_id := range vals {
		chat_id := chat_ids[i]
		out[chat_id] = max_id
	}
	return out
}

// 批量获得对话已读最大消息ID
func GetUserReadIdMap(self_id string, peer_ids []string) map[string]int64 {
	var out = make(map[string]int64)
	if len(peer_ids) == 0 {
		return out
	}
	var keys []any
	for _, peer_id := range peer_ids {
		k := cache.DialogRead(self_id, peer_id).Key
		keys = append(keys, k)
	}
	vals := redis.Do("cache", "MGET", keys...).Int64s()
	for i, max_id := range vals {
		peer_id := peer_ids[i]
		out[peer_id] = max_id
	}
	return out
}

// 填充对话列表
// 超级群对话并非实时同步的, 所以每次获取对话列表前都要先通过计算把用户的超级群对话合并到主列表
// 算法: 获取我所有的超级群ID; 获取所有超级群时间; 加入对话列表
func PaddingDialog(self_id string) {
	chat_ids := GetChannelIdsOfUser(self_id)
	if len(chat_ids) == 0 {
		return
	}
	tm := GetChannelTimeMap(chat_ids)
	ddc := cache.UserDDCMap(self_id, "").HGetAll().IntMap()
	c := redis.Begin("cache")
	for chat_id, t := range tm {
		del_at := ddc[chat_id]
		if t > del_at {
			cache.UserDialogZset(self_id).Tx(c).ZAdd(chat_id, t)
			cache.UserDDCMap(self_id, chat_id).HDel()
		}
	}
	redis.Commit(c)
}

func SearchTLDialogs(self_id string, offset, limit int) *proto.Dialogs {
	var out = &proto.Dialogs{
		Dialogs: []*proto.Dialog{},
	}
	key := cache.UserDialogZset(self_id)
	if offset == 0 {
		PaddingDialog(self_id)
		out.Total = key.ZCard()
	}
	kname := key.Key
	// 将置顶对话提升到最高排名
	pinned_ids := GetPinnedDialogIds(self_id)
	if offset == 0 {
		c := redis.Begin("cache")
		for _, peer_id := range pinned_ids {
			cache.UserDialogZset(self_id).Tx(c).ZAddXX(peer_id, math.MaxInt64)
		}
		redis.Commit(c)
	}
	peer_ids := redis.Do("cache", "ZREVRANGE", kname, offset, limit-1).Strings()
	var user_ids []string
	var chat_ids []string
	var ug_ids []string
	var channel_ids []string
	for _, peer_id := range peer_ids {
		peert := def.ToIdType(peer_id)
		if peert == def.IdUser {
			user_ids = append(user_ids, peer_id)
		} else {
			chat_ids = append(chat_ids, peer_id)
		}
		if peert != def.IdChannel {
			ug_ids = append(ug_ids, peer_id)
		} else {
			channel_ids = append(channel_ids, peer_id)
		}
	}
	var userm = GetTLUserMap(user_ids)
	var chatm = GetTLChatMap(chat_ids)
	var unreadm = GetUserDialogUnreadMap(self_id, ug_ids)
	var creadm = GetUserReadIdMap(self_id, channel_ids)
	var uptsm = GetUserDialogPtsMap(self_id, ug_ids)
	var cptsm = GetChannelPtsMap(channel_ids)
	var ureceiptm = GetUserReceiptMap(self_id, user_ids)
	var creceiptm = GetChatReceiptMap(chat_ids)
	var u_msg_ids []int64
	var c_msg_ids []*InputChanMsg
	for _, pts := range uptsm {
		if pts != 0 {
			u_msg_ids = append(u_msg_ids, pts)
		}
	}
	for chat_id, pts := range cptsm {
		if pts != 0 {
			in := &InputChanMsg{
				ChatId: chat_id,
				MsgId:  pts,
			}
			c_msg_ids = append(c_msg_ids, in)
		}
	}
	var umsgm = GetMessageMap(u_msg_ids)
	var cmsgm = GetChannelsMessageMap(c_msg_ids)
	var useqm = GetUserSeqMap(self_id, ug_ids)
	var cseqm = GetChannelSeqMap(channel_ids)
	var notifym = GetPeerNotifySettingMap(self_id, peer_ids)
	for _, peer_id := range peer_ids {
		d := &proto.Dialog{
			PeerId: peer_id,
		}
		pt := def.ToIdType(peer_id)
		if pt == def.IdUser {
			d.PeerUser = userm[peer_id]
			d.ReceiptMaxId = ureceiptm[peer_id]
		} else {
			d.PeerChat = chatm[peer_id]
			d.ReceiptMaxId = creceiptm[peer_id]
		}
		if pt != def.IdChannel {
			pts := uptsm[peer_id]
			d.TopMessage = umsgm[pts]
			d.Unread = unreadm[peer_id]
			d.Pts = pts
			d.Seq = useqm[peer_id]
		} else {
			pts := cptsm[peer_id]
			d.TopMessage = cmsgm[InputChanMsg{ChatId: peer_id, MsgId: pts}]
			//d.Unread = int(pts - creadm[peer_id]) // deprecated: 分布式环境下消息ID不连续
			d.Unread = GetChannelUnread(peer_id, creadm[peer_id])
			d.Pts = pts
			d.Seq = cseqm[peer_id]
		}
		d.Pinned = base.InArray(peer_id, pinned_ids)
		d.NotifySetting = notifym[peer_id].TL()
		out.Dialogs = append(out.Dialogs, d)
	}
	return out
}

func RemoveDialog(user_id, peer_id string) {
	c := redis.Begin("cache")
	cache.DialogPts(user_id, peer_id).Tx(c).Del()
	cache.DialogUnread(user_id, peer_id).Tx(c).Del()
	cache.DialogRead(user_id, peer_id).Tx(c).Del()
	cache.UserDialogZset(user_id).Tx(c).ZRem(peer_id)
	cache.UserPinnedDialogMap(user_id).HDel(peer_id)
	cache.UserPeerNotifyMap(user_id).HDel(peer_id)
	if def.ToIdType(peer_id) == def.IdChannel {
		now := int(time.Now().Unix())
		cache.UserDDCMap(user_id, peer_id).Tx(c).HSet(now)
	}
	redis.Commit(c)
}

// 删除对话聊天记录
func ClearMsgbox(self_id, peer_id string) bool {
	ok := db.Primary.Exec(`delete from user_msgbox where user_id=? and peer_id=?`,
		self_id, peer_id).OK()
	return ok
}

// 清空对话聊天记录, 发消息
// 同时标记已读
func TLClearDialog(user_id, peer_id string) {
	if def.ToIdType(peer_id) != def.IdChannel {
		ClearMsgbox(user_id, peer_id)
	}
	pts := GetDialogPts(user_id, peer_id)
	c := redis.Begin("cache")
	cache.DialogUnread(user_id, peer_id).Tx(c).Set(0)
	cache.DialogRead(user_id, peer_id).Tx(c).Set(pts)
	redis.Commit(c)

	ev := &proto.Event{
		DialogClear: &proto.EvDialogClear{
			PeerId: peer_id,
			MaxId:  pts,
		},
	}
	PushUserOfflineEvent(user_id, peer_id, def.EvDialogClear, ev)
}

// 删除对话, 清空聊天记录, 删除免打扰设置
// 同时发送EvDialogDeleted事件
func TLDeleteDialog(user_id, peer_id string, clear bool) {
	if clear && def.ToIdType(peer_id) != def.IdChannel {
		ClearMsgbox(user_id, peer_id)
	}
	pts := GetDialogPts(user_id, peer_id)
	RemoveDialog(user_id, peer_id)
	PushEvDialogDeleted(user_id, peer_id, clear, pts)
}

// 获取对话最新消息发送时间
func GetTopMessageTime(self_id, peer_id string) int {
	pts := GetDialogPts(self_id, peer_id)
	if def.ToIdType(peer_id) == def.IdChannel {
		msg := GetChannelMessage(peer_id, pts)
		if msg != nil {
			return msg.CreatedAt
		}
	} else {
		msg := GetMessage(pts)
		if msg != nil {
			return msg.CreatedAt
		}
	}
	return 0
}

// 置顶/取消置顶对话
func TLPinDialog(self_id, peer_id string, pinned bool) bool {
	var ok bool
	if pinned {
		now := int(time.Now().Unix())
		ok = cache.UserPinnedDialogMap(self_id).HSetNX2(peer_id, now).Bool()
	} else {
		ok = cache.UserPinnedDialogMap(self_id).HDel(peer_id).Bool()
		// 取消置顶: 恢复消息发送时间作为排名权重
		send_at := GetTopMessageTime(self_id, peer_id)
		cache.UserDialogZset(self_id).ZAddXX(peer_id, send_at)
	}
	if !ok {
		return false
	}
	// send sync event
	PushEvPinDialog(self_id, peer_id, pinned)
	return true
}

// 获取置顶对话ID列表 返回[peer_id]
func GetPinnedDialogIds(self_id string) []string {
	return cache.UserPinnedDialogMap(self_id).HKeys().Strings()
}

type PeerNotifySetting struct {
	Badge bool `json:"badge"` // 是否将未读数显示为小红点(true显示小红点 false显示未读数)
}

func (s *PeerNotifySetting) TL() *proto.PeerNotifySetting {
	if s == nil {
		return &proto.PeerNotifySetting{}
	}
	return &proto.PeerNotifySetting{
		Badge: s.Badge,
	}
}

func (s *PeerNotifySetting) Equal(a *PeerNotifySetting) bool {
	if s == nil || a == nil {
		return s == a
	}
	return *s == *a
}

// 获取用户对话免打扰设置
func GetPeerNotifySetting(self_id, peer_id string) *PeerNotifySetting {
	var out *PeerNotifySetting
	cache.UserPeerNotifyMap(self_id).HGet(peer_id).Unmarshal(&out)
	return out
}

// 批量获取对话免打扰设置
func GetPeerNotifySettingMap(self_id string, peer_ids []string) map[string]*PeerNotifySetting {
	var out = make(map[string]*PeerNotifySetting)
	if len(peer_ids) == 0 {
		return out
	}
	var args []any
	for _, peer_id := range peer_ids {
		args = append(args, peer_id)
	}
	vals := cache.UserPeerNotifyMap(self_id).HMGet(args...).Strings()
	for i, val := range vals {
		peer_id := peer_ids[i]
		var s *PeerNotifySetting
		json.Unmarshal([]byte(val), &s)
		out[peer_id] = s
	}
	return out
}

// 更新用户免打扰设置
func UpdatePeerNotifySetting(self_id, peer_id string, s *PeerNotifySetting) bool {
	old := GetPeerNotifySetting(self_id, peer_id)
	if old.Equal(s) {
		return false
	}
	cache.UserPeerNotifyMap(self_id).HSet2(peer_id, s).Bool()
	PushEvUpdatePeerNotify(self_id, peer_id, s)
	return true
}
