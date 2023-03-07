package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Sys_disconnect godoc
// @Summary      强制用户断线
// @Description  失败响应:
// @Description  400 USER_ID_EMPTY 参数为空
// @Tags         服务端集成
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param user_id formData string true "用户ID"
// @Param token formData string true "(可选)具体到设备"
// @Param reason formData string true "参考*断线原因*"
// @Success 200 {object} proto.Success{data=bool} "返回是否有数据变更"
// @Router       /sys/disconnect [post]
func Sys_disconnect(md *service.Meta) (any, error) {
	user_id := md.Get("user_id").String()
	token := md.Get("token").String()
	reason := md.Get("reason").String()
	if user_id == "" {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	biz.PushCmdDisconnect(user_id, token, reason)
	return true, nil
}
