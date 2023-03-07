package method

import (
	"zim.cn/biz"
	"zim.cn/biz/proto"
	"zim.cn/service"
)

func Message_readHistory(md *service.Meta) (any, error) {
	peer_id := md.Get("peer_id").String()
	self_id := md.UserId
	max_id := md.Get("max_id").Int64()
	if peer_id == "" {
		return nil, service.NewError(400, "PEER_ID_EMPTY")
	}
	if !biz.IsDialogExists(self_id, peer_id) {
		return nil, service.NewError(400, "DIALOG_NOT_FOUND")
	}
	pts, unread := biz.ReadHistory(self_id, peer_id, max_id)
	out := &proto.AffectedHistory{
		Pts:    pts,
		Unread: unread,
	}
	return out, nil
}
