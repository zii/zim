package biz

import (
	"time"

	"zim.cn/base"

	"zim.cn/biz/proto"

	"zim.cn/biz/cache"

	"zim.cn/base/db"
)

type FriendRaw struct {
	UserId  string `json:"user_id"`
	PeerId  string `json:"peer_id"`
	Name    string `json:"name"`
	Blocked bool   `json:"blocked"`
}

func (r *FriendRaw) TL(peer *UserRaw) *proto.Friend {
	var out = &proto.Friend{
		UserId:  r.PeerId,
		User:    peer.TL(),
		Name:    r.Name,
		Blocked: r.Blocked,
	}
	name := r.Name
	if name == "" && peer != nil {
		name = peer.Name
	}
	out.Letter = base.PinyinInitials(name)
	return out
}

// self是不是已经添加了peer为单向好友
func IsFriend(self_id, peer_id string) bool {
	key := cache.IsFriendKey(self_id, peer_id)
	r := key.Get()
	if r.Reply != nil {
		return r.Bool()
	}
	var b bool
	db.Replica.Get(&b, `select 1 from friend where user_id=? and peer_id=?`, self_id, peer_id)
	key.Set(b)
	return b
}

// 添加单向好友(事务)
func AddFriendTx(tx *db.MustTx, self_id, peer_id string, name string) bool {
	now := int(time.Now().Unix())
	ok := tx.Exec(`insert ignore into friend set user_id=?, peer_id=?, name=?, blocked=0, created_at=?`,
		self_id, peer_id, name, now).OK()
	if !ok {
		ok = tx.Exec(`update friend set name=?, blocked=0, created_at=? where user_id=? and peer_id=?`,
			name, now, self_id, peer_id).OK()
	}
	return ok
}

// 添加单向好友
func AddFriend(self_id, peer_id string, name string) bool {
	tx := db.Primary.Begin()
	defer tx.Rollback()
	ok := AddFriendTx(tx, self_id, peer_id, name)
	tx.Commit()
	if ok {
		cache.IsFriendKey(self_id, peer_id).Set(true)
	}
	return ok
}

// 添加双向好友
// name: 我方备注昵称
// name2: 对方备注昵称
func AddFriendMutal(self_id, peer_id string, name, name2 string) bool {
	tx := db.Primary.Begin()
	defer tx.Rollback()
	ok1 := AddFriendTx(tx, self_id, peer_id, name2)
	var ok2 bool
	if !IsFriend(peer_id, self_id) {
		ok2 = AddFriendTx(tx, peer_id, self_id, name)
	}
	tx.Commit()
	ok := ok1 || ok2
	if ok {
		cache.IsFriendKey(self_id, peer_id).Set(true)
		cache.IsFriendKey(peer_id, self_id).Set(true)
	}
	return ok
}

// (业务层)单双向添加好友
func TLAddFriend(self_id, peer_id string, name string, mutal bool) bool {
	if !mutal {
		ok := AddFriend(self_id, peer_id, name)
		if !ok {
			return false
		}
		if IsFriend(peer_id, self_id) {
			ap := &FriendApplyRaw{
				FromId: self_id,
				ToId:   peer_id,
			}
			SendBecomeFriends(ap)
		}
		return true
	} else {
		ok := AddFriendMutal(self_id, peer_id, "", name)
		if !ok {
			return false
		}
		ap := &FriendApplyRaw{
			FromId: self_id,
			ToId:   peer_id,
		}
		SendBecomeFriends(ap)
		return true
	}
}

// 获取用户所有好友
func GetFriends(user_id string, blocked bool) []*FriendRaw {
	var out []*FriendRaw
	rows := db.Replica.Query(`select peer_id, name, blocked from friend where user_id=? and blocked=?`,
		user_id, blocked)
	defer rows.Close()
	for rows.Next() {
		var r = &FriendRaw{}
		r.UserId = user_id
		rows.Scan(&r.PeerId, &r.Name, &r.Blocked)
		out = append(out, r)
	}
	return out
}

func GetTLFriends(user_id string, blocked bool) []*proto.Friend {
	raws := GetFriends(user_id, blocked)
	var peer_ids []string
	for _, r := range raws {
		peer_ids = append(peer_ids, r.PeerId)
	}
	peerm := GetUserMap(peer_ids)
	var out = []*proto.Friend{}
	for _, r := range raws {
		peer := peerm[r.PeerId]
		tl := r.TL(peer)
		out = append(out, tl)
	}
	return out
}

func EditFriend(self_id, peer_id string, name string) bool {
	ok := db.Primary.Exec(`update friend set name=? where user_id=? and peer_id=?`,
		name, self_id, peer_id).OK()
	return ok
}

func RemoveFriend(self_id, peer_id string) bool {
	ok := db.Primary.Exec(`delete from friend where user_id=? and peer_id=?`, self_id, peer_id).OK()
	if !ok {
		return false
	}
	cache.IsFriendKey(self_id, peer_id).Del()
	// 删除对话
	TLDeleteDialog(self_id, peer_id, true)
	// 删除申请记录
	DeleteFriendApply2(self_id, peer_id)
	return true
}

func BlockFriend(self_id, peer_id string) bool {
	ok := db.Primary.Exec(`update friend set blocked=1 where user_id=? and peer_id=?`,
		self_id, peer_id).OK()
	return ok
}

func UnblockFriend(self_id, peer_id string) bool {
	ok := db.Primary.Exec(`update friend set blocked=0 where user_id=? and peer_id=?`,
		self_id, peer_id).OK()
	return ok
}

// 获取好友数量
func GetFriendCount(user_id string) int {
	var n int
	db.Replica.Get(&n, `select count(*) from friend where user_id=?`, user_id)
	return n
}
