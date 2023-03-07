package biz

import (
	"encoding/json"
	"strings"

	"zim.cn/biz/def"

	"zim.cn/base/db"
	"zim.cn/base/redis"
	"zim.cn/biz/cache"
	"zim.cn/biz/proto"
)

type UserRaw struct {
	UserId    string `json:"uid"`
	Name      string `json:"name"`
	Photo     string `json:"photo"`
	Ex        string `json:"ex"`
	Status    int    `json:"stat"`
	CreatedAt int    `json:"ct"`
}

func (r *UserRaw) TL() *proto.User {
	if r == nil {
		return nil
	}
	tl := &proto.User{
		Id:     r.UserId,
		Name:   r.Name,
		Photo:  r.Photo,
		Ex:     r.Ex,
		Status: r.Status,
	}
	return tl
}

func (r *UserRaw) TinyTL() *proto.TinyUser {
	if r == nil {
		return nil
	}
	tl := &proto.TinyUser{
		Id:   r.UserId,
		Name: r.Name,
	}
	return tl
}

// 是否禁用
func (r *UserRaw) Banned() bool {
	return r.Status == def.UserBanned
}

// 是否正常
func (r *UserRaw) OK() bool {
	return r.Status == def.UserOK
}

func LoadUser(user_id string) *UserRaw {
	var u = &UserRaw{}
	ok := db.Replica.QueryRow(`select user_id, name, photo, ex, status, created_at from user where user_id=?`,
		user_id).Scan(&u.UserId, &u.Name, &u.Photo, &u.Ex, &u.Status, &u.CreatedAt)
	if !ok {
		return nil
	}
	return u
}

func GetUser(user_id string) *UserRaw {
	key := cache.User(user_id)
	r := key.Get()
	if r.Reply != nil {
		var u *UserRaw
		err := r.Unmarshal(&u)
		if err == nil {
			return u
		}
	}
	u := LoadUser(user_id)
	key.Set(u)
	return u
}

func GetTLUser(user_id string) *proto.User {
	r := GetUser(user_id)
	return r.TL()
}

func GetTinyUser(user_id string) *proto.TinyUser {
	r := GetUser(user_id)
	return r.TinyTL()
}

func LoadUsers(user_ids []string) []*UserRaw {
	var out []*UserRaw
	if len(user_ids) == 0 {
		return out
	}
	q := `select user_id, name, photo, ex, status, created_at from user`
	p := db.Prepare()
	p.In("user_id", user_ids)
	q += p.Clause()
	rows := db.Replica.Query(q, p.Args()...)
	defer rows.Close()
	for rows.Next() {
		var u = &UserRaw{}
		rows.Scan(&u.UserId, &u.Name, &u.Photo, &u.Ex, &u.Status, &u.CreatedAt)
		out = append(out, u)
	}
	return out
}

func GetUserMap(user_ids []string) map[string]*UserRaw {
	var out = make(map[string]*UserRaw)
	if len(user_ids) == 0 {
		return out
	}
	var keys []any
	for _, user_id := range user_ids {
		k := cache.User(user_id).Key
		keys = append(keys, k)
	}
	var missing []string
	vals := redis.Do("speed", "MGET", keys...).Strings()
	for i, s := range vals {
		user_id := user_ids[i]
		if s == "" {
			missing = append(missing, user_id)
			continue
		}
		var u *UserRaw
		err := json.Unmarshal([]byte(s), &u)
		if err != nil {
			missing = append(missing, user_id)
			continue
		}
		out[user_id] = u
	}
	if len(missing) > 0 {
		users := LoadUsers(missing)
		for _, u := range users {
			out[u.UserId] = u
			cache.User(u.UserId).Set(u)
		}
	}
	return out
}

func GetUsers(user_ids []string) []*UserRaw {
	var out []*UserRaw
	m := GetUserMap(user_ids)
	for _, user_id := range user_ids {
		u := m[user_id]
		if u != nil {
			out = append(out, u)
		}
	}
	return out
}

func GetTLUsers(user_ids []string) []*proto.User {
	var out []*proto.User
	users := GetUsers(user_ids)
	for _, u := range users {
		out = append(out, u.TL())
	}
	return out
}

func GetTLUserMap(user_ids []string) map[string]*proto.User {
	var out = make(map[string]*proto.User)
	m := GetUserMap(user_ids)
	for user_id, raw := range m {
		out[user_id] = raw.TL()
	}
	return out
}

func loadChannelIdsOfUser(user_id string) []string {
	var out []string
	rows := db.Replica.Query(`select m.chat_id from chat_member m
inner join chat c on c.chat_id=m.chat_id
where m.user_id=? and c.type=2 and m.deleted=0`, user_id)
	defer rows.Close()
	for rows.Next() {
		var chat_id string
		rows.Scan(&chat_id)
		out = append(out, chat_id)
	}
	return out
}

func GetChannelIdsOfUser(user_id string) []string {
	key := cache.ChannelsOfUserKey(user_id)
	r := key.Get()
	if r.Reply != nil {
		var out []string
		err := r.Unmarshal(&out)
		if err == nil {
			return out
		}
	}
	chat_ids := loadChannelIdsOfUser(user_id)
	key.Set(chat_ids)
	return chat_ids
}

func EditUser(user *UserRaw, name, photo, ex string) bool {
	var fields []string
	var args []any
	if name != "" && user.Name != name {
		fields = append(fields, "name=?")
		args = append(args, name)
	}
	if photo != "" && user.Photo != photo {
		fields = append(fields, "photo=?")
		args = append(args, photo)
	}
	if ex != "" && user.Ex != ex {
		fields = append(fields, "ex=?")
		args = append(args, ex)
	}
	if len(fields) == 0 {
		return false
	}
	q := `update user set ` + strings.Join(fields, ", ") + " where user_id=?"
	args = append(args, user.UserId)
	ok := db.Primary.Exec(q, args...).OK()
	if ok {
		cache.User(user.UserId).Del()
	}
	return ok
}

func SetUserStatus(user_id string, status int) bool {
	ok := db.Primary.Exec(`update user set status=? where user_id=?`, status, user_id).OK()
	if ok {
		cache.User(user_id).Del()
	}
	return ok
}
