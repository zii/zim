package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Sys_unban godoc
// @Summary      解除禁用用户账号
// @Description  失败响应:
// @Description  400 USER_STATUS_INVALID 用户状态不正确
// @Tags         服务端集成
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param user_id formData string false "用户ID"
// @Success 200 {object} proto.Success{data=bool} "返回是否有数据变更"
// @Router       /sys/unban [post]
func Sys_unban(md *service.Meta) (any, error) {
	user_id := md.Get("user_id").String()
	if user_id == "" {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	user := biz.GetUser(user_id)
	if user == nil {
		return nil, service.NewError(400, "USER_ID_INVALID")
	}
	if user.OK() {
		return false, nil
	}
	if !user.Banned() {
		return nil, service.NewError(400, "USER_STATUS_INVALID")
	}
	ok := biz.SetUserStatus(user_id, def.UserOK)
	if !ok {
		return false, nil
	}
	return true, nil
}
