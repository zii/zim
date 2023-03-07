package method

import (
	"zim.cn/base/log"

	"zim.cn/biz"
	"zim.cn/service"
)

// Sys_authToken godoc
// @Summary      后台授权新token
// @Description  失败响应:
// @Description  400 USER_ID_INVALID 用户账号不存在
// @Description  400 FORBIDDEN 用户账号异常，已被禁止登录
// @Tags         服务端集成
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param user_id formData string true "用户ID"
// @Param platform formData int true "平台类型(1:IOS 2:ANDROID 3:WEB 4:DESKTOP)"
// @Param device_id formData string true "设备ID, 用来做推送通知"
// @Success 200 {object} proto.Success{data=proto.Authorization} "返回授权结果"
// @Router       /sys/authToken [post]
func Sys_authToken(md *service.Meta) (any, error) {
	user_id := md.Get("user_id").String()
	if user_id == "" {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	platform := md.Get("platform").Int()
	if platform == 0 {
		return nil, service.NewError(400, "PLATFORM_EMPTY")
	}
	device_id := md.Get("device_id").String()
	if device_id == "" {
		return nil, service.NewError(400, "DEVICE_ID_EMPTY")
	}
	user := biz.GetUser(user_id)
	if user == nil {
		return nil, service.NewError(400, "USER_ID_INVALID")
	}
	if user.Banned() {
		return nil, service.NewError(403, "FORBIDDEN")
	}
	out, err := biz.Authorize(user_id, platform, device_id)
	if err != nil {
		log.Error("Authorize:", err)
		return nil, service.NewError(400, err.Error())
	}
	return out, nil
}
