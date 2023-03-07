package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Friend_editApply godoc
// @Summary      修改好友申请信息
// @Description  失败响应:
// @Description  400 NAME_LONG 备注名过长
// @Description  400 HASH_INVALID 申请记录不存在
// @Description  400 STATUS_INVALID 申请状态错误
// @Tags         好友(通讯录)管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param hash formData string true "申请记录唯一标识"
// @Param name formData string false "备注名"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /friend/editApply [post]
func Friend_editApply(md *service.Meta) (any, error) {
	self_id := md.UserId
	hash := md.Get("hash").String()
	if hash == "" {
		return nil, service.NewError(400, "HASH_EMPTY")
	}
	name := md.Get("name").String()
	if len(name) > 50 {
		return nil, service.NewError(400, "NAME_LONG")
	}
	ap := biz.GetFriendApply(self_id, hash)
	if ap == nil {
		return nil, service.NewError(400, "HASH_INVALID")
	}
	if ap.Status != def.FriendApplyWait {
		return nil, service.NewError(400, "STATUS_INVALID")
	}
	ok := biz.EditApply(self_id, hash, name)
	return ok, nil
}
