package method

import (
	"zim.cn/biz/def"

	"zim.cn/biz/proto"

	"zim.cn/biz"
	"zim.cn/service"
)

func Message_sendMessage(md *service.Meta) (any, error) {
	var msg *proto.Message
	err := md.Get("message").Unmarshal(&msg)
	if err != nil {
		return nil, service.NewError(400, "MESSAGE_INVALID")
	}
	if msg == nil {
		return nil, service.NewError(400, "MESSAGE_INVALID")
	}
	if msg.Type.Class() != def.Msg {
		return nil, service.NewError(400, "TYPE_INVALID")
	}
	msg.FromId = md.UserId
	err = biz.VerifyMessage(msg)
	if err != nil {
		return nil, service.NewError(400, err.Error())
	}
	err = biz.SendMessage(msg)
	if err != nil {
		return nil, service.NewError(400, err.Error())
	}
	return msg.Id, nil
}
