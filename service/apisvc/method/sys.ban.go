package method

import (
	"zim.cn/biz"
	"zim.cn/biz/cookie"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Sys_ban godoc
// @Summary      禁用用户账号
// @Description  强制登出用户, 删除所有token; 下次授权时自动FORBIDDEN
// @Description  失败响应:
// @Description  400 USER_STATUS_INVALID 用户账号异常
// @Tags         服务端集成
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param user_id formData string true "用户ID"
// @Success 200 {object} proto.Success{data=bool} "返回是否有数据变更"
// @Router       /sys/ban [post]
func Sys_ban(md *service.Meta) (any, error) {
	user_id := md.Get("user_id").String()
	if user_id == "" {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	user := biz.GetUser(user_id)
	if user == nil {
		return nil, service.NewError(400, "USER_ID_INVALID")
	}
	if user.Banned() {
		return false, nil
	}
	if !user.OK() {
		return nil, service.NewError(400, "USER_STATUS_INVALID")
	}
	ok := biz.SetUserStatus(user_id, def.UserBanned)
	if !ok {
		return false, nil
	}
	cookie.ClearUserToken(user_id)
	biz.PushCmdDisconnect(user_id, "", "BANNED")
	return true, nil
}
