package biz

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"zim.cn/biz/def"

	"zim.cn/base"

	"zim.cn/base/db"
	"zim.cn/biz/proto"
)

type FriendApplyRaw struct {
	UserId    string         `json:"user_id"`
	FromId    string         `json:"from_id"`
	ToId      string         `json:"to_id"`
	Hash      string         `json:"hash"`
	Greets    []*proto.Greet `json:"greets,omitempty"`
	Name      string         `json:"name"`
	Status    int            `json:"status"`
	UpdatedAt int            `json:"updated_at"`
}

func (r *FriendApplyRaw) Copy() *FriendApplyRaw {
	b := *r
	return &b
}

func (r *FriendApplyRaw) PeerId() string {
	if r.UserId == r.FromId {
		return r.ToId
	}
	return r.FromId
}

func (r *FriendApplyRaw) TL(peer_user *UserRaw) *proto.FriendApply {
	var out = &proto.FriendApply{
		Hash:      r.Hash,
		FromId:    r.FromId,
		PeerUser:  peer_user.TL(),
		ToId:      r.ToId,
		Greets:    r.Greets,
		Name:      r.Name,
		Status:    r.Status,
		UpdatedAt: r.UpdatedAt,
	}
	return out
}

func GetFriendApply(user_id, hash string) *FriendApplyRaw {
	var r = &FriendApplyRaw{}
	var gs sql.NullString
	ok := db.Replica.QueryRow(`select from_id, to_id, greets, name, status, updated_at from friend_apply
	where user_id=? and hash=?`, user_id, hash).Scan(&r.FromId, &r.ToId, &gs, &r.Name, &r.Status, &r.UpdatedAt)
	if !ok {
		return nil
	}
	r.UserId = user_id
	r.Hash = hash
	json.Unmarshal([]byte(gs.String), &r.Greets)
	return r
}

// 获得双方申请记录 {user_id:*FriendApplyRaw}
func GetFriendApplys(from_id, to_id string) map[string]*FriendApplyRaw {
	var out = make(map[string]*FriendApplyRaw)
	rows := db.Replica.Query(`select user_id, hash, greets, name, status, updated_at from friend_apply
	where from_id=? and to_id=?`, from_id, to_id)
	defer rows.Close()
	for rows.Next() {
		var a = &FriendApplyRaw{}
		var gs sql.NullString
		rows.Scan(&a.UserId, &a.Hash, &gs, &a.Name, &a.Status, &a.UpdatedAt)
		a.FromId = from_id
		a.ToId = to_id
		json.Unmarshal([]byte(gs.String), &a.Greets)
		out[a.UserId] = a
	}
	return out
}

// 添加单方好友申请记录
func addFriendApply(user_id, from_id, to_id string, greet, name string, status int, raw *FriendApplyRaw) *FriendApplyRaw {
	now := int(time.Now().Unix())
	hash := base.Md5String([]byte(fmt.Sprintf("%s:%s", from_id, to_id)))
	if raw == nil {
		greets := []*proto.Greet{
			{greet, now},
		}
		greets_json := base.JsonString(greets)
		db.Primary.Exec(`insert into friend_apply set user_id=?, from_id=?, to_id=?, hash=?, greets=?, name=?, 
		status=?, updated_at=?`, user_id, from_id, to_id, hash, greets_json, name, status, now)
		out := &FriendApplyRaw{
			UserId:    user_id,
			FromId:    from_id,
			ToId:      to_id,
			Hash:      hash,
			Greets:    greets,
			Name:      name,
			Status:    status,
			UpdatedAt: now,
		}
		return out
	} else {
		raw := raw.Copy()
		raw.UserId = user_id
		raw.FromId = from_id
		raw.ToId = to_id
		raw.Hash = hash
		raw.Greets = append(raw.Greets, &proto.Greet{
			Text: greet,
			Time: now,
		})
		if len(raw.Greets) > def.MaxGreetNum {
			raw.Greets = raw.Greets[len(raw.Greets)-def.MaxGreetNum:]
		}
		greets_json := base.JsonString(raw.Greets)
		raw.Name = name
		raw.Status = status
		raw.UpdatedAt = now
		db.Primary.Exec(`update friend_apply set greets=?, name=?, status=?, updated_at=? where user_id=? and hash=?`,
			greets_json, name, status, now, user_id, hash)
		return raw
	}
}

// 添加双方好友申请记录, 返回邀请者申请记录
func AddFriendApplys(from_id, to_id string, greet, name string, status int) *FriendApplyRaw {
	apm := GetFriendApplys(from_id, to_id)
	var user_ids = []string{from_id, to_id}
	var out *FriendApplyRaw
	for _, user_id := range user_ids {
		ap := apm[user_id]
		na := ""
		if user_id == from_id {
			na = name
		}
		ap = addFriendApply(user_id, from_id, to_id, greet, na, status, ap)
		if user_id == from_id {
			out = ap
		}
	}
	return out
}

func SendBecomeFriends(ap *FriendApplyRaw) {
	msg := &proto.Message{
		Type:   def.TipBecomeFriends,
		FromId: ap.ToId,
		ToId:   ap.FromId,
		Tip: &proto.Tip{
			BecomeFriends: &proto.TipBecomeFriends{
				Greets: ap.Greets,
				FromId: ap.FromId,
				ToId:   ap.ToId,
			},
		},
	}
	SendMessage(msg)
}

func SendFriendApply(to_id string) {
	msg := &proto.Message{
		Type:   def.MsgText,
		FromId: def.IdFriend,
		ToId:   to_id,
		Elem: &proto.Elem{
			Text: "apply",
		},
	}
	SendMessage(msg)
}

// 邀请好友
// from_id: 邀请人ID
// to_id: 被邀请人ID
// greet: 招呼语
// name: 备注名称
// 先插入申请记录; 如果对方已经加了我, 直接成为好友; 否则用#friend账号给to_id发送申请消息
// 返回邀请者的最新邀请记录
func InviteFriend(from_id, to_id string, greet, name string) (*proto.FriendApply, error) {
	var ap *FriendApplyRaw
	if IsFriend(to_id, from_id) {
		if IsFriend(from_id, to_id) {
			return nil, errors.New("DUPLICATED_APPLY")
		}
		ap = AddFriendApplys(from_id, to_id, greet, name, def.FriendApplyAccept)
		SendBecomeFriends(ap)
	} else {
		ap = AddFriendApplys(from_id, to_id, greet, name, def.FriendApplyWait)
		SendFriendApply(to_id)
	}
	peer_id := ap.PeerId()
	peer_user := GetUser(peer_id)
	out := ap.TL(peer_user)
	return out, nil
}

func UpdateApplyStatus(hash string, status int) bool {
	ok := db.Primary.Exec(`update friend_apply set status=? where hash=?`,
		status, hash).OK()
	return ok
}

// 接受好友申请
func AcceptFriend(self_id, hash string) (bool, error) {
	ap := GetFriendApply(self_id, hash)
	if ap == nil {
		return false, errors.New("APPLY_NOTFOUND")
	}
	if ap.Status == def.FriendApplyAccept {
		return false, nil
	}
	if ap.ToId != self_id {
		return false, errors.New("ACCESS_DENIED")
	}
	to := GetUser(ap.ToId)
	if to == nil {
		return false, errors.New("USER_ID_INVALID")
	}
	if to.Banned() {
		return false, errors.New("USER_FORBIDDEN")
	}
	ap.Status = def.FriendApplyAccept
	if ok := UpdateApplyStatus(hash, def.FriendApplyAccept); !ok {
		return false, nil
	}
	var peer_name string
	peer_ap := GetFriendApply(ap.PeerId(), hash)
	if peer_ap != nil {
		peer_name = peer_ap.Name
	}
	ok := AddFriendMutal(ap.FromId, ap.ToId, ap.Name, peer_name)
	if ok {
		SendBecomeFriends(ap)
	}
	return true, nil
}

// 删除好友申请(by hash)
func DeleteFriendApply(self_id, hash string) bool {
	ok := db.Primary.Exec(`delete from friend_apply where user_id=? and hash=?`,
		self_id, hash).OK()
	return ok
}

// 删除好友申请(by peer_id)
func DeleteFriendApply2(self_id, peer_id string) bool {
	ok := db.Primary.Exec(`delete from friend_apply where user_id=? and (to_id=? or from_id=?)`,
		self_id, peer_id, peer_id).OK()
	return ok
}

// 好友申请列表
func SearchTLFriendApplys(self_id string, offset, limit int) []*proto.FriendApply {
	var out = []*proto.FriendApply{}
	var applys []*FriendApplyRaw
	q := `select from_id, to_id, hash, greets, name, status, updated_at from friend_apply`
	p := db.Prepare()
	p.And("user_id=?", self_id)
	p.Slice(offset, limit)
	p.Sort("updated_at desc")
	q += p.Clause()
	rows := db.Replica.Query(q, p.Args()...)
	defer rows.Close()
	var peer_ids []string
	for rows.Next() {
		var a = &FriendApplyRaw{}
		var gs []byte
		rows.Scan(&a.FromId, &a.ToId, &a.Hash, &gs, &a.Name, &a.Status, &a.UpdatedAt)
		json.Unmarshal(gs, &a.Greets)
		a.UserId = self_id
		if a.FromId != "" {
			peer_ids = append(peer_ids, a.PeerId())
		}
		applys = append(applys, a)
	}
	peerm := GetUserMap(peer_ids)
	for _, a := range applys {
		peer_user := peerm[a.PeerId()]
		tl := a.TL(peer_user)
		out = append(out, tl)
	}
	return out
}

// 修改好友申请昵称
func EditApply(self_id, hash string, name string) bool {
	ok := db.Primary.Exec(`update friend_apply set name=? where user_id=? and hash=?`,
		name, self_id, hash).OK()
	return ok
}
