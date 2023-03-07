package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Friend_getFriends godoc
// @Summary      全部好友的列表
// @Description  失败响应:
// @Description  400 LIMIT_EXCEED 分页参数错误
// @Tags         好友(通讯录)管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Success 200 {object} proto.Success{data=[]proto.Friend} "返回全部好友列表"
// @Router       /friend/getFriends [post]
func Friend_getFriends(md *service.Meta) (any, error) {
	self_id := md.UserId
	out := biz.GetTLFriends(self_id, false)
	return out, nil
}
