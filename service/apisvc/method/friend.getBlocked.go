package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Friend_getBlocked godoc
// @Summary      黑名单列表
// @Description  失败响应:
// @Description  400 LIMIT_EXCEED 分页参数错误
// @Tags         好友(通讯录)管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Success 200 {object} proto.Success{data=[]proto.Friend} "返回黑名单好友列表"
// @Router       /friend/getBlocked [post]
func Friend_getBlocked(md *service.Meta) (any, error) {
	self_id := md.UserId
	out := biz.GetTLFriends(self_id, true)
	return out, nil
}
