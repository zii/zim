package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Chat_getChat godoc
// @Summary      获取单个群详情
// @Tags         群组管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param id formData string true "群ID"
// @Success 200 {object} proto.Success{data=proto.Chat} "群信息"
// @Router       /chat/getChat [post]
func Chat_getChat(md *service.Meta) (any, error) {
	chat_id := md.Get("id").String()
	if chat_id == "" {
		return nil, service.NewError(400, "ID_EMPTY")
	}
	out := biz.GetTLChat(chat_id)
	return out, nil
}
