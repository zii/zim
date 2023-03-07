package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Sys_register godoc
// @Summary      后台创建im用户
// @Description  失败响应:
// @Description  400 NAME_EMPTY 昵称不能为空
// @Tags         服务端集成
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param name formData string false "昵称"
// @Param photo formData string false "头像URL"
// @Param ex formData string false "自定义信息"
// @Success 200 {object} proto.Success{data=string} "返回用户ID"
// @Router       /sys/register [post]
func Sys_register(md *service.Meta) (any, error) {
	name := md.Get("name").String()
	photo := md.Get("photo").String()
	ex := md.Get("ex").String()
	if name == "" {
		return nil, service.NewError(400, "NAME_EMPTY")
	}
	user_id := biz.CreateUser(name, photo, ex)
	return user_id, nil
}
