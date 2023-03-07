package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Friend_getApplyList godoc
// @Summary      获取好友申请列表
// @Description  包括我邀请别人的记录和别人邀请我的记录, 按申请时间倒序排列翻页
// @Description  失败响应:
// @Description  400 LIMIT_EXCEED 分页参数错误
// @Tags         好友(通讯录)管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param offset formData int true "分页偏移"
// @Param limit formData int false "每页条数"
// @Success 200 {object} proto.Success{data=proto.FriendApply} "返回申请记录列表"
// @Router       /friend/getApplyList [post]
func Friend_getApplyList(md *service.Meta) (any, error) {
	self_id := md.UserId
	offset := md.Get("offset").Int()
	limit := md.Get("limit").Int()
	if err := biz.PagingVerify(&offset, &limit, 100); err != nil {
		return nil, service.NewError(400, err.Error())
	}
	out := biz.SearchTLFriendApplys(self_id, offset, limit)
	return out, nil
}
