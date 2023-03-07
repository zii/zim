// 纯粹为了加速用的key, 可以随时清库
package cache

import (
	"fmt"

	"zim.cn/base/redis"
)

// 用户
func User(user_id string) *redis.Key {
	return &redis.Key{
		DBName:   "speed",
		Key:      fmt.Sprintf("user:%s", user_id),
		Timeout:  24 * 3600,
		Critical: true,
		Encoding: "json",
	}
}

// 图片缓存
func Photo(photo_id int64) *redis.Key {
	return &redis.Key{
		DBName:   "speed",
		Key:      fmt.Sprintf("photo:%d", photo_id),
		Timeout:  24 * 3600,
		Encoding: "json",
	}
}

func SettingKey(name string) *redis.Key {
	key := &redis.Key{
		DBName:  "speed",
		Key:     fmt.Sprintf("setting:%s", name),
		Timeout: 86400,
	}
	return key
}

// 用户所在超级群ID
func ChannelsOfUserKey(user_id string) *redis.Key {
	return &redis.Key{
		DBName:   "speed",
		Key:      fmt.Sprintf("user:%s:channels", user_id),
		Timeout:  24 * 3600,
		Encoding: "json",
	}
}

// 用户消息缓存
func MessageKey(id int64) *redis.Key {
	key := &redis.Key{
		DBName:   "speed",
		Key:      fmt.Sprintf("msg:%d", id),
		Timeout:  86400,
		Encoding: "json",
	}
	return key
}

// 超级群消息缓存
func ChannelMsgKey(chat_id string, id int64) *redis.Key {
	key := &redis.Key{
		DBName:   "speed",
		Key:      fmt.Sprintf("channel:%s:msg:%d", chat_id, id),
		Timeout:  86400,
		Encoding: "json",
	}
	return key
}

// 群信息缓存
func ChatKey(id string) *redis.Key {
	key := &redis.Key{
		DBName:   "speed",
		Key:      fmt.Sprintf("chat:%s", id),
		Timeout:  86400,
		Encoding: "json",
	}
	return key
}

// 群成员信息缓存
func MemberKey(chat_id, user_id string) *redis.Key {
	key := &redis.Key{
		DBName:   "speed",
		Key:      fmt.Sprintf("chat:%s:member:%s", chat_id, user_id),
		Timeout:  86400,
		Encoding: "json",
	}
	return key
}

// peer_id是否在self_id的通讯录中
func IsFriendKey(self_id, peer_id string) *redis.Key {
	key := &redis.Key{
		DBName:  "speed",
		Key:     fmt.Sprintf("isfriend:%s:%s", self_id, peer_id),
		Timeout: 86400,
	}
	return key
}
