package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Sys_editUser godoc
// @Summary      修改用户资料
// @Description  失败响应:
// @Description  400 USER_ID_EMPTY 参数为空
// @Tags         服务端集成
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param user_id formData string true "用户ID"
// @Param name formData string false "昵称"
// @Param photo formData string false "头像URL"
// @Param ex formData string false "自定义信息"
// @Success 200 {object} proto.Success{data=bool} "返回是否有数据变更"
// @Router       /sys/editUser [post]
func Sys_editUser(md *service.Meta) (any, error) {
	user_id := md.Get("user_id").String()
	if user_id == "" {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	name := md.Get("name").String()
	photo := md.Get("photo").String()
	ex := md.Get("ex").String()
	user := biz.GetUser(user_id)
	if user == nil {
		return nil, service.NewError(400, "USER_ID_INVALID")
	}
	ok := biz.EditUser(user, name, photo, ex)
	return ok, nil
}
