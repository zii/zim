package method

import (
	"zim.cn/base"
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Chat_editRole godoc
// @Summary      修改成员角色
// @Description  失败响应:
// @Description  400 SELF_NOT_MEMBER 您不是群成员
// @Description  400 USER_NOT_MEMBER 对方不是群成员
// @Description  400 ACCESS_DENIED 您没有权限，需要管理员或群主
// @Tags         群组管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param chat_id formData string true "群ID"
// @Param user_id formData string true "群成员ID"
// @Param role formData int true "角色类型(0普通成员 1管理员 2群主)"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /chat/editRole [post]
func Chat_editRole(md *service.Meta) (any, error) {
	chat_id := md.Get("chat_id").String()
	user_id := md.Get("user_id").String()
	role := md.Get("role").Int()
	self_id := md.UserId
	if chat_id == "" {
		return nil, service.NewError(400, "CHAT_ID_EMPTY")
	}
	if user_id == "" {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	if !base.InArray(def.ChatRole(role), []def.ChatRole{def.RoleAdmin, def.RoleMember}) {
		return nil, service.NewError(400, "ROLE_INVALID")
	}
	selfm := biz.GetMember(chat_id, self_id)
	if selfm == nil {
		return nil, service.NewError(400, "SELF_NOT_MEMBER")
	}
	userm := biz.GetMember(chat_id, user_id)
	if userm == nil {
		return nil, service.NewError(400, "USER_NOT_MEMBER")
	}
	if self_id != user_id && selfm.Role <= userm.Role {
		return nil, service.NewError(400, "ACCESS_DENIED")
	}
	ok := biz.EditRole(self_id, chat_id, user_id, role)
	return ok, nil
}
