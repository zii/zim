package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Chat_deleteUser godoc
// @Summary      删除成员/退出群
// @Description  失败响应:
// @Description  400 NOT_MEMBER 您不是群成员
// @Description  400 ACCESS_DENIED 您没有权限，需要管理员或群主
// @Tags         群组管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param chat_id formData string true "群ID"
// @Param user_id formData string true "群成员ID"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /chat/deleteUser [post]
func Chat_deleteUser(md *service.Meta) (any, error) {
	chat_id := md.Get("chat_id").String()
	user_id := md.Get("user_id").String()
	self_id := md.UserId
	if chat_id == "" {
		return nil, service.NewError(400, "CHAT_ID_EMPTY")
	}
	if user_id == "" {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	selfm := biz.GetMember(chat_id, self_id)
	if selfm == nil {
		return nil, service.NewError(400, "NOT_MEMBER")
	}
	userm := biz.GetMember(chat_id, user_id)
	if userm == nil {
		return false, nil
	}
	if userm.Role == def.RoleOwner {
		return nil, service.NewError(400, "ACCESS_DENIED")
	}
	ok := biz.DeleteMember(chat_id, user_id, self_id != user_id)
	return ok, nil
}
