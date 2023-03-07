package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// User_getUser godoc
// @Summary      获取单个用户详情
// @Description  失败响应:
// @Description  400 ID_EMPTY 参数为空
// @Tags         其他
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param id formData string true "用户ID"
// @Success 200 {object} proto.Success{data=proto.User} "返回单个用户模型"
// @Router       /user/getUser [post]
func User_getUser(md *service.Meta) (any, error) {
	user_id := md.Get("id").String()
	if user_id == "" {
		return nil, service.NewError(400, "ID_EMPTY")
	}
	out := biz.GetTLUser(user_id)
	return out, nil
}
