package biz

import (
	"encoding/json"
	"time"

	"zim.cn/biz/def"

	"zim.cn/base/db"
	"zim.cn/biz/proto"
)

func InsertEvent(self_id, peer_id string, seq int64, msg *proto.Message) {
	msg.Event.Seq = seq
	b := msg.Blob()
	now := int(time.Now().Unix())
	db.Primary.Exec(`insert into event(self_id, peer_id, seq, msg_blob, created_at) values(?,?,?,?,?)`,
		self_id, peer_id, seq, b, now)
}

// 加载历史事件
func LoadEvents(user_id, peer_id string, min_id, max_id int64, offset, limit int) []*proto.Message {
	var out = []*proto.Message{}
	pt := def.ToIdType(peer_id)
	q := `select msg_blob from event`
	p := db.Prepare()
	if pt == def.IdChannel {
		p.And("self_id=?", peer_id)
	} else {
		p.And("self_id=? and peer_id=?", user_id, peer_id)
	}
	if min_id > 0 {
		p.And("seq>?", min_id)
	}
	if max_id > 0 {
		p.And("seq<=?", max_id)
	}
	p.Sort("seq desc")
	p.Slice(offset, limit)
	q += p.Clause()
	rows := db.Replica.Query(q, p.Args()...)
	defer rows.Close()
	for rows.Next() {
		var blob []byte
		rows.Scan(&blob)
		var m *proto.Message
		json.Unmarshal(blob, &m)
		out = append(out, m)
	}
	return out
}
