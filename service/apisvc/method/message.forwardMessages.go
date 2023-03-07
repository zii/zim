package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

func Message_forwardMessages(md *service.Meta) (any, error) {
	self_id := md.UserId
	to_ids := md.Get("to_id").Strings()
	peer_id := md.Get("peer_id").String()
	msg_ids := md.Get("msg_id").Int64s()
	merge := md.Get("merge").Bool()
	if len(to_ids) == 0 {
		return nil, service.NewError(400, "TO_ID_EMPTY")
	}
	if len(to_ids) > 100 {
		return nil, service.NewError(400, "最大发给100人")
	}
	if peer_id == "" {
		return nil, service.NewError(400, "PEER_ID_EMPTY")
	}
	if len(msg_ids) == 0 {
		return nil, service.NewError(400, "MSG_ID_EMPTY")
	}
	if len(msg_ids) > 100 {
		return nil, service.NewError(400, "消息条数不能超过100条")
	}
	if !merge {
		err := biz.ForwardMessages(self_id, to_ids, peer_id, msg_ids)
		if err != nil {
			return nil, service.NewError(400, err.Error())
		}
	} else {
		err := biz.SendChatlog(self_id, to_ids, peer_id, msg_ids)
		if err != nil {
			return nil, service.NewError(400, err.Error())
		}
	}
	return true, nil
}
