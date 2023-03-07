package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Chat_dismiss godoc
// @Summary      解散群
// @Description  失败响应:
// @Description  400 SELF_NOT_MEMBER 您不是群成员
// @Description  400 ACCESS_DENIED 您没有权限，需要管理员或群主
// @Description  400 CHAT_ID_INVALID 群不存在
// @Tags         群组管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param chat_id formData string true "群ID"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /chat/dismiss [post]
func Chat_dismiss(md *service.Meta) (any, error) {
	chat_id := md.Get("chat_id").String()
	self_id := md.UserId
	if chat_id == "" {
		return nil, service.NewError(400, "CHAT_ID_EMPTY")
	}
	selfm := biz.GetMember(chat_id, self_id)
	if selfm == nil {
		return nil, service.NewError(400, "SELF_NOT_MEMBER")
	}
	if selfm.Role != def.RoleOwner {
		return nil, service.NewError(400, "ACCESS_DENIED")
	}
	chat := biz.GetChat(chat_id)
	if chat == nil {
		return nil, service.NewError(400, "CHAT_ID_INVALID")
	}
	ok := biz.DismissChat(chat)
	return ok, nil
}
