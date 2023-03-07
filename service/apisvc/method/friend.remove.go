package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Friend_remove godoc
// @Summary      移除好友
// @Description  同时删除好友申请记录, 删除对话, 聊天记录
// @Description  失败响应:
// @Description  400 USER_ID_EMPTY 参数为空
// @Tags         好友(通讯录)管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param user_id formData string true "好友ID"
// @Success 200 {object} proto.Success{data=bool} "返回是否发生数据变更"
// @Router       /friend/remove [post]
func Friend_remove(md *service.Meta) (any, error) {
	self_id := md.UserId
	peer_id := md.Get("user_id").String()
	if peer_id == "" {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	ok := biz.RemoveFriend(self_id, peer_id)
	return ok, nil
}
