package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Chat_editInfo godoc
// @Summary      修改群信息
// @Description  失败响应:
// @Description  400 CHAT_INVALID 群不存在
// @Description  400 NOT_MEMBER 您不是群成员
// @Description  400 ACCESS_DENIED 您没有权限，需要管理员或群主
// @Tags         群组管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param chat_id formData string true "群ID"
// @Param title formData string true "群标题"
// @Param about formData string true "群公告"
// @Param photo formData string true "群图标URL"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /chat/editInfo [post]
func Chat_editInfo(md *service.Meta) (any, error) {
	chat_id := md.Get("chat_id").String()
	title := md.Get("title").String()
	about := md.Get("about").String()
	photo := md.Get("photo").String()
	if chat_id == "" {
		return nil, service.NewError(400, "CHAT_ID_EMPTY")
	}
	old := biz.GetChat(chat_id)
	if old == nil {
		return nil, service.NewError(400, "CHAT_INVALID")
	}
	self_id := md.UserId
	selfm := biz.GetMember(chat_id, self_id)
	if selfm == nil {
		return nil, service.NewError(400, "NOT_MEMBER")
	}
	if selfm.Role != def.RoleOwner {
		return nil, service.NewError(400, "ACCESS_DENIED")
	}
	nue := &biz.ChatRaw{
		ChatId: chat_id,
		Title:  title,
		About:  about,
		Photo:  photo,
	}
	ok := biz.EditChatInfo(old, nue)
	return ok, nil
}
