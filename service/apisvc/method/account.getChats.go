package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Account_getChats godoc
// @Summary      用户所在群列表
// @Description  失败响应:
// @Description  400 LIMIT_EXCEED 分页参数错误
// @Tags         用户体系
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param offset formData int false "分页偏移量, 默认0"
// @Param limit  formData int false "每页条数, 最大100"
// @Success 200 {object} proto.Success{data=[]proto.Chat} "群列表"
// @Router       /account/getChats [post]
func Account_getChats(md *service.Meta) (any, error) {
	self_id := md.UserId
	offset := md.Get("offset").Int()
	limit := md.Get("limit").Int()
	if err := biz.PagingVerify(&offset, &limit, 100); err != nil {
		return nil, service.NewError(400, err.Error())
	}
	out := biz.SearchUserTLChats(self_id, offset, limit)
	return out, nil
}
