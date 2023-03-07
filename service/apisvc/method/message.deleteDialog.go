package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

func Message_deleteDialog(md *service.Meta) (any, error) {
	self_id := md.UserId
	peer_id := md.Get("peer_id").String()
	clear := md.Get("clear").Bool()
	if peer_id == "" {
		return nil, service.NewError(400, "PEER_ID_EMPTY")
	}
	if !biz.IsDialogExists(self_id, peer_id) {
		return false, nil
	}
	biz.TLDeleteDialog(self_id, peer_id, clear)
	return true, nil
}
