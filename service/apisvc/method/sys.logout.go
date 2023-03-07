package method

import (
	"zim.cn/biz"
	"zim.cn/biz/cookie"
	"zim.cn/service"
)

// Sys_logout godoc
// @Summary      强制登出用户
// @Tags         服务端集成
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param token formData string true "对方用户token"
// @Success 200 {object} proto.Success{data=bool} "返回是否有数据变更"
// @Router       /sys/logout [post]
func Sys_logout(md *service.Meta) (any, error) {
	token := md.Get("token").String()
	cv, err := cookie.Parse(token)
	if err != nil {
		return nil, service.NewError(400, err.Error())
	}
	cookie.DelUserToken(cv.UserId, token)
	biz.PushCmdDisconnect(cv.UserId, token, "LOGOUT")
	return true, nil
}
