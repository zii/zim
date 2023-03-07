package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

func Message_deleteMessages(md *service.Meta) (any, error) {
	self_id := md.UserId
	peer_id := md.Get("peer_id").String()
	msg_ids := md.Get("id").Int64s()
	if peer_id == "" {
		return nil, service.NewError(400, "PEER_ID_EMPTY")
	}
	if len(msg_ids) == 0 {
		return nil, service.NewError(400, "MSG_ID_EMPTY")
	}
	if len(msg_ids) > 100 {
		return nil, service.NewError(400, "LIMIT_EXCEED")
	}
	var ok bool
	pt := def.ToIdType(peer_id)
	if pt == def.IdChannel {
		ok = biz.DeleteChannelMessages(peer_id, msg_ids)
	} else {
		ok = biz.DeleteUserMessages(self_id, peer_id, msg_ids)
	}
	return ok, nil
}
