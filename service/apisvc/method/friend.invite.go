package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Friend_invite godoc
// @Summary      发起好友申请
// @Description  如果对方已添加自己, 发送TipBecomeFriends, 直接成为好友; 否则#friend给to_id发送消息, 收到在通讯录按钮显示小红点
// @Description  失败响应:
// @Description  400 GREET_LONG 问候语过长
// @Description  400 NAME_LONG 备注名过长
// @Description  400 TO_ID_INVALID 对方账号不存在
// @Description  400 USER_FORBIDDEN 对方账号异常，已被禁止登录
// @Description  400 LIMIT_EXCEEDED 您的好友数已达上限
// @Description  400 DUPLICATED_APPLY 重复邀请
// @Tags         好友(通讯录)管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param to_id formData string true "被邀请用户ID"
// @Param greet formData string false "问候语"
// @Param name formData string false "备注名"
// @Success 200 {object} proto.Success{data=proto.FriendApply} "返回申请记录"
// @Router       /friend/invite [post]
func Friend_invite(md *service.Meta) (any, error) {
	self_id := md.UserId
	to_id := md.Get("to_id").String()
	greet := md.Get("greet").String()
	name := md.Get("name").String()
	if self_id == to_id {
		return nil, service.NewError(400, "TO_ID_INVALID")
	}
	if len(greet) > 255 {
		return nil, service.NewError(400, "GREET_LONG")
	}
	if len(name) > 50 {
		return nil, service.NewError(400, "NAME_LONG")
	}
	to := biz.GetUser(to_id)
	if to == nil {
		return nil, service.NewError(400, "TO_ID_INVALID")
	}
	if to.Banned() {
		return nil, service.NewError(400, "USER_FORBIDDEN")
	}
	fn := biz.GetFriendCount(self_id)
	if fn >= def.FriendLimit {
		return nil, service.NewError(400, "LIMIT_EXCEEDED")
	}
	out, err := biz.InviteFriend(self_id, to_id, greet, name)
	if err != nil {
		return nil, service.NewError(400, err.Error())
	}
	return out, nil
}
