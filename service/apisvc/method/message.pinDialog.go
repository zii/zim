package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

func Message_pinDialog(md *service.Meta) (any, error) {
	self_id := md.UserId
	peer_id := md.Get("peer_id").String()
	pinned := md.Get("pinned").Bool()
	if peer_id == "" {
		return nil, service.NewError(400, "PEER_ID_EMPTY")
	}
	ok := biz.TLPinDialog(self_id, peer_id, pinned)
	return ok, nil
}
