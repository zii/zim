package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Friend_deleteApply godoc
// @Summary      删除好友申请
// @Description  失败响应:
// @Description  400 HASH_EMPTY 参数为空
// @Tags         好友(通讯录)管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param hash formData string true "申请记录唯一标识"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /friend/deleteApply [post]
func Friend_deleteApply(md *service.Meta) (any, error) {
	self_id := md.UserId
	hash := md.Get("hash").String()
	if hash == "" {
		return nil, service.NewError(400, "HASH_EMPTY")
	}
	ok := biz.DeleteFriendApply(self_id, hash)
	return ok, nil
}
