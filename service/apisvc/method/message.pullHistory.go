package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

func Message_pullHistory(md *service.Meta) (any, error) {
	peer_id := md.Get("peer_id").String()
	self_id := md.UserId
	max_id := md.Get("max_id").Int64()
	min_id := md.Get("min_id").Int64()
	offset := md.Get("offset").Int()
	limit := md.Get("limit").Int()
	if peer_id == "" {
		return nil, service.NewError(400, "PEER_ID_EMPTY")
	}
	if offset < 0 {
		offset = 0
	}
	if limit < 0 {
		limit = 0
	}
	msgs := biz.LoadHistory(self_id, peer_id, min_id, max_id, offset, limit)
	// 拉取第一页自动标记已读
	if offset == 0 {
		if max_id <= 0 {
			pts := biz.GetDialogPts(self_id, peer_id)
			max_id = pts
		}
		biz.ReadHistory(self_id, peer_id, max_id)
	}
	return msgs, nil
}
