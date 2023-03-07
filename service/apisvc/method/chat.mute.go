package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Chat_mute godoc
// @Summary      全员禁言
// @Description  失败响应:
// @Description  400 NOT_MEMBER 您不是群成员
// @Description  400 ACCESS_DENIED 您没有权限，需要管理员或群主
// @Tags         群组管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param chat_id formData string true "群ID"
// @Param duration formData int true "时长(0无限)"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /chat/mute [post]
func Chat_mute(md *service.Meta) (any, error) {
	chat_id := md.Get("chat_id").String()
	duration := md.Get("duration").Int()
	self_id := md.UserId
	if chat_id == "" {
		return nil, service.NewError(400, "CHAT_ID_EMPTY")
	}
	if duration < 0 {
		duration = 0
	}
	selfm := biz.GetMember(chat_id, self_id)
	if selfm == nil {
		return nil, service.NewError(400, "NOT_MEMBER")
	}
	if selfm.Role != def.RoleOwner {
		return nil, service.NewError(400, "ACCESS_DENIED")
	}
	ok := biz.MuteChat(self_id, chat_id, true, duration)
	return ok, nil
}
