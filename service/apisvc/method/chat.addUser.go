package method

import (
	"zim.cn/base/log"

	"zim.cn/biz"
	"zim.cn/service"
)

// Chat_addUser godoc
// @Summary      添加成员
// @Description  失败响应:
// @Description  400 CHAT_DISMISSED 群已解散
// @Description  400 LIMIT_EXCEEDED 群人数已达上限
// @Description  400 NOT_MEMBER 您不是群成员
// @Description  400 MEMBER_DUPLICATED 不可添加重复的成员
// @Tags         群组管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param chat_id formData string true "群ID"
// @Param user_id formData string true "用户ID列表" default(["u2"])
// @Success 200 {object} proto.Success{data=bool}
// @Router       /chat/addUser [post]
func Chat_addUser(md *service.Meta) (any, error) {
	chat_id := md.Get("chat_id").String()
	user_ids := md.Get("user_id").Strings()
	self_id := md.UserId
	if chat_id == "" {
		return nil, service.NewError(400, "CHAT_ID_EMPTY")
	}
	if len(user_ids) == 0 {
		return nil, service.NewError(400, "USER_ID_EMPTY")
	}
	chat := biz.GetChat(chat_id)
	if chat.Deleted {
		return nil, service.NewError(400, "CHAT_DISMISSED")
	}
	if chat.Maxp > 0 {
		n := biz.GetMemberCount(chat_id)
		if n >= chat.Maxp {
			return nil, service.NewError(400, "LIMIT_EXCEEDED")
		}
	}
	selfm := biz.GetMember(chat_id, self_id)
	if selfm == nil {
		return nil, service.NewError(400, "NOT_MEMBER")
	}
	for _, user_id := range user_ids {
		if biz.GetMember(chat_id, user_id) != nil {
			log.Warn("MEMBER_DUPLICATED:", chat_id, user_id)
			return nil, service.NewError(400, "MEMBER_DUPLICATED")
		}
	}
	err := biz.AddMember(self_id, chat_id, user_ids)
	if err != nil {
		return nil, service.NewError(400, err.Error())
	}
	return true, nil
}
