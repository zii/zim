package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Chat_unmute godoc
// @Summary      解除全员禁言
// @Description  失败响应:
// @Description  400 NOT_MEMBER 您不是群成员
// @Description  400 ACCESS_DENIED 您没有权限，需要管理员或群主
// @Tags         群组管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param chat_id formData string true "群ID"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /chat/unmute [post]
func Chat_unmute(md *service.Meta) (any, error) {
	chat_id := md.Get("chat_id").String()
	self_id := md.UserId
	if chat_id == "" {
		return nil, service.NewError(400, "CHAT_ID_EMPTY")
	}
	selfm := biz.GetMember(chat_id, self_id)
	if selfm == nil {
		return nil, service.NewError(400, "NOT_MEMBER")
	}
	if selfm.Role != def.RoleOwner {
		return nil, service.NewError(400, "ACCESS_DENIED")
	}
	ok := biz.MuteChat(self_id, chat_id, false, 0)
	return ok, nil
}
