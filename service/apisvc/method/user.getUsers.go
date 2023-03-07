package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// User_getUsers godoc
// @Summary      批量获取多个用户详情
// @Description  失败响应:
// @Description  400 LIMIT_EXCEEDED 超过单次最大查询人数上限
// @Tags         其他
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param id formData string true "用户ID的JSON数组" default(["u2"])
// @Success 200 {object} proto.Success{data=[]proto.User} "返回用户列表"
// @Router       /user/getUsers [post]
func User_getUsers(md *service.Meta) (any, error) {
	user_ids := md.Get("id").Strings()
	if len(user_ids) == 0 {
		return nil, service.NewError(400, "ID_EMPTY")
	}
	if len(user_ids) > 100 {
		return nil, service.NewError(400, "LIMIT_EXCEEDED")
	}
	out := biz.GetTLUsers(user_ids)
	return out, nil
}
