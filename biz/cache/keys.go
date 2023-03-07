// 重要的key
// 设计原则:
// 1. 主库尽量不要产生大集合(map/set/zset), 因为90%都是冷数据, 而map没有超时机制
// 2. 数据库里能查到,就做成临时key,放到speed库
package cache

import (
	"fmt"

	"zim.cn/base/redis"
)

type CookieValue struct {
	UserId   string `json:"id"`
	ExpireAt int    `json:"ex"`
	// device params
	Appkey   string `json:"appkey"`
	Platform int    `json:"platform"`
	DeviceId string `json:"device_id"` // 设备ID
}

// cookie:{token} = {"user_id":用户ID, "expire":过期时间}
func CookieKey(token string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Key:      fmt.Sprintf("cookie:%s", token),
		Timeout:  7 * 86400,
		Encoding: "json",
		Critical: true,
	}
}

// user:{用户ID}:cookie:zset = {token}:{过期时间}
// 只用来限制用户cookie数量, 不作它用
func UserCookiesKey(user_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Key:      fmt.Sprintf("user:%s:cookie:zset", user_id),
		Critical: true,
	}
}

// 用户token数
func TokenCount(user_id string) int {
	return UserCookiesKey(user_id).ZCard()
}

// smscode:{phone}:{code} = hash;
func SmsCodeKey(phone_number string, code string) *redis.Key {
	/* 用户基本信息 */
	return &redis.Key{
		DBName:   "cache",
		Key:      fmt.Sprintf("smscode:%s:%s", phone_number, code),
		Timeout:  120,
		Critical: true,
	}
}

// 本地缓存修订版本号
func LocalRevisonNo(name string) *redis.Key {
	key := &redis.Key{
		DBName:   "cache",
		Key:      fmt.Sprintf("local:%s:revision", name),
		Critical: true,
	}
	return key
}

// 本地缓存修订号对应的key名
func LocalRevisionK(name string, no int) *redis.Key {
	key := &redis.Key{
		DBName:  "cache",
		Key:     fmt.Sprintf("local:%s:rev:%d:k", name, no),
		Timeout: 86400,
	}
	return key
}

func UserIdCounter() *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Key:      fmt.Sprintf("user_id:counter"),
		Critical: true,
	}
}

// 消息ID生成器
func PtsKey() *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Key:      "pts",
		Critical: true,
	}
}

// 事件ID生成器
func SeqKey() *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Key:      "seq",
		Critical: true,
	}
}

// 用户已读最大对话消息ID(inbox id)
func DialogRead(user_id, peer_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("dialog:%s:%s:read", user_id, peer_id),
	}
}

// 对话最新消息ID
func DialogPts(user_id, peer_id string) *redis.Key {
	return &redis.Key{
		DBName:  "cache",
		Key:     fmt.Sprintf("dialog:%s:%s:pts", user_id, peer_id),
		Timeout: 7 * 86400,
	}
}

// 对话未读数
func DialogUnread(user_id, peer_id string) *redis.Key {
	return &redis.Key{
		DBName:  "cache",
		Key:     fmt.Sprintf("dialog:%s:%s:unread", user_id, peer_id),
		Timeout: 7 * 86400,
	}
}

// 超级群最新消息ID
func ChannelPts(chat_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("channel:%s:pts", chat_id),
	}
}

// 超级群最新消息时间 用于对话列表排序
func ChannelTime(chat_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("channel:%s:time", chat_id),
	}
}

// 用户删掉超级群对话时的时间字典 用于对话列表排序
func UserDDCMap(user_id, channel_id string) *redis.Key {
	key := &redis.Key{
		DBName: "cache",
		Key:    fmt.Sprintf("user:%s:ddc:map", user_id),
	}
	if channel_id != "" {
		key.SecondaryKey = channel_id
	}
	return key
}

// 群成员列表
// key: user_id
// score=role*1000000000+入群时间/上线时间
func ChatMemberZset(chat_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("chat:%s:member:zset", chat_id),
	}
}

// 用户对话列表 score=最新消息时间
func UserDialogZset(user_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("user:%s:dialog:zset", user_id),
	}
}

// 用户置顶对话 {peer_id:置顶时间}
func UserPinnedDialogMap(user_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("user:%s:pinned:dialog:map", user_id),
	}
}

// 用户免打扰设置 {peer_id:PeerNotifySetting}
func UserPeerNotifyMap(user_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("user:%s:pnotify:map", user_id),
		Encoding: "json",
	}
}

// 单聊对话最大已读回执消息ID
func UserReceiptMaxId(user_id, peer_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("user:%s:%s:receipt", user_id, peer_id),
	}
}

// 群聊最大已读回执消息ID
func ChatReceiptMaxId(chat_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("chat:%s:receipt", chat_id),
	}
}

// 用户/普通群对话最新事件ID
func UserSeq(user_id, peer_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("user:%s:%s:seq", user_id, peer_id),
	}
}

// 超级群最新事件ID
func ChannelSeq(chat_id string) *redis.Key {
	return &redis.Key{
		DBName:   "cache",
		Critical: true,
		Key:      fmt.Sprintf("channel:%s:seq", chat_id),
	}
}

// 用户在线设备 {token:platform}
func UserOnlineMap(user_id string, token string) *redis.Key {
	return &redis.Key{
		DBName:       "cache",
		Key:          fmt.Sprintf("user:%s:online:map", user_id),
		SecondaryKey: token,
	}
}

/*
超级群消息分段计数器 内部实现参考:lua_segincr
*/
func ChannelCounter(chat_id string) *redis.Key {
	key := &redis.Key{
		DBName:   "cache",
		Key:      fmt.Sprintf("channel:%s:counter:map", chat_id),
		Critical: true,
	}
	return key
}
