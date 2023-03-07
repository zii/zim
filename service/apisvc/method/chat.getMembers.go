package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

// Chat_getMembers godoc
// @Summary      群成员列表
// @Description  失败响应:
// @Description  400 LIMIT_EXCEED 分页参数错误
// @Tags         群组管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param chat_id formData string true "群ID"
// @Param offset formData int true "分页偏移量"
// @Param limit formData int true "每页条数"
// @Success 200 {object} proto.Success{data=[]proto.Member} "返回群成员列表"
// @Router       /chat/getMembers [post]
func Chat_getMembers(md *service.Meta) (any, error) {
	chat_id := md.Get("chat_id").String()
	if chat_id == "" {
		return nil, service.NewError(400, "CHAT_ID_EMPTY")
	}
	offset := md.Get("offset").Int()
	limit := md.Get("limit").Int()
	if err := biz.PagingVerify(&offset, &limit, 100); err != nil {
		return nil, err
	}
	out := biz.GetTLMembers(chat_id, offset, limit)
	return out, nil
}
