package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

func Message_pullEvents(md *service.Meta) (any, error) {
	peer_id := md.Get("peer_id").String()
	self_id := md.UserId
	min_id := md.Get("min").Int64()
	max_id := md.Get("max").Int64()
	offset := md.Get("offset").Int()
	limit := md.Get("limit").Int()
	if peer_id == "" {
		return nil, service.NewError(400, "PEER_ID_EMPTY")
	}
	if limit < 0 {
		limit = 0
	}
	if limit > 100 {
		return nil, service.NewError(400, "LIMIT_EXCEED")
	}
	out := biz.LoadEvents(self_id, peer_id, min_id, max_id, offset, limit)
	return out, nil
}
