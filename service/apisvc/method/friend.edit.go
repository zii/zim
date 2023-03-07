package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Friend_edit godoc
// @Summary      修改好友信息
// @Description  失败响应:
// @Description  400 NOT_FRIEND 对方不是您的好友
// @Description  400 NAME_LONG 备注名过长
// @Tags         好友(通讯录)管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param user_id formData string true "好友ID"
// @Param name formData string false "备注名"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /friend/edit [post]
func Friend_edit(md *service.Meta) (any, error) {
	self_id := md.UserId
	user_id := md.Get("user_id").String()
	if user_id == "" {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	if !biz.IsFriend(self_id, user_id) {
		return nil, service.NewError(400, "NOT_FRIEND")
	}
	name := md.Get("name").String()
	if len(name) > 50 {
		return nil, service.NewError(400, "NAME_LONG")
	}
	ok := biz.EditFriend(self_id, user_id, name)
	return ok, nil
}
