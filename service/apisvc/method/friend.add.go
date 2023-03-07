package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Friend_add godoc
// @Summary      直接添加好友
// @Description  双向添加成功后发送TipBecomeFriends消息
// @Description  失败响应:
// @Description  400 NAME_LONG 备注名过长
// @Description  400 USER_ID_INVALID 对方账号不存在
// @Description  400 USER_FORBIDDEN 对方账号异常，已被禁止登录
// @Description  400 LIMIT_EXCEEDED 您的好友数量已达上限
// @Description  400 PEER_LIMIT_EXCEEDED 对方好友数量已达上限
// @Tags         好友(通讯录)管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param user_id formData string true "对方用户ID"
// @Param name formData string false "好友备注名"
// @Param mutal formData bool false "是否添加双向好友"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /friend/add [post]
func Friend_add(md *service.Meta) (any, error) {
	self_id := md.UserId
	peer_id := md.Get("user_id").String()
	if peer_id == "" {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	name := md.Get("name").String()
	if len(name) > 50 {
		return nil, service.NewError(400, "NAME_LONG")
	}
	mutal := md.Get("mutal").Bool()
	to := biz.GetUser(peer_id)
	if to == nil {
		return nil, service.NewError(400, "USER_ID_INVALID")
	}
	if to.Banned() {
		return nil, service.NewError(400, "USER_FORBIDDEN")
	}
	fn := biz.GetFriendCount(self_id)
	if fn >= def.FriendLimit {
		return nil, service.NewError(400, "LIMIT_EXCEEDED")
	}
	pn := biz.GetFriendCount(peer_id)
	if pn >= def.FriendLimit {
		return nil, service.NewError(400, "PEER_LIMIT_EXCEEDED")
	}
	ok := biz.TLAddFriend(self_id, peer_id, name, mutal)
	return ok, nil
}
