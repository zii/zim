package method

import (
	"zim.cn/biz/proto"

	"zim.cn/biz"
	"zim.cn/service"
)

// Sys_sendMessage godoc
// @Summary      后台发消息接口
// @Description  失败响应:
// @Description  400 MESSAGE_INVALID 消息无法解析
// @Description  400 FROM_ID_INVALID 发送者账号不存在
// @Description  400 FROM_USER_BANNED 发送者账号异常
// @Tags         服务端集成
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param message formData string true "消息结构" default({"type": 101, "from_id":"#friend", "to_id":"u4", "elem":{"text": "你好1"}})
// @Success 200 {object} proto.Success{data=string} "返回消息ID"
// @Router       /sys/sendMessage [post]
func Sys_sendMessage(md *service.Meta) (any, error) {
	var msg *proto.Message
	err := md.Get("message").Unmarshal(&msg)
	if err != nil {
		return nil, service.NewError(400, "MESSAGE_INVALID")
	}
	if msg == nil {
		return nil, service.NewError(400, "MESSAGE_INVALID")
	}
	if msg.FromId == "" {
		return nil, service.NewError(400, "FROM_ID_EMPTY")
	}
	if msg.ToId == "" {
		return nil, service.NewError(400, "TO_ID_EMPTY")
	}
	err = biz.SendMessage(msg)
	if err != nil {
		return nil, service.NewError(400, err.Error())
	}
	return msg.Id, nil
}
