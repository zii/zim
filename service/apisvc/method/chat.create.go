package method

import (
	"zim.cn/base"
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Chat_create godoc
// @Summary      创建新群
// @Description  失败响应:
// @Description  400 MEMBER_ID_INVALID 初始成员不存在
// @Tags         群组管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param type formData int true "群类型(1普通群 2超级群)"
// @Param title formData string true "群标题"
// @Param about formData string false "群公告"
// @Param init_members formData string true "初始成员用户ID列表,最大100人" default(["u2"])
// @Success 200 {object} proto.Success{data=string} "返回群ID"
// @Router       /chat/create [post]
func Chat_create(md *service.Meta) (any, error) {
	typ := md.Get("type").Int()
	title := md.Get("title").String()
	about := md.Get("about").String()
	init_members := md.Get("init_members").Strings()
	if !base.InInts(typ, []int{def.TypeGroup, def.TypeChannel}) {
		return nil, service.NewError(400, "TYPE_INVALID")
	}
	self_id := md.UserId
	for _, member_id := range init_members {
		if member_id == "" {
			return nil, service.NewError(400, "MEMBER_ID_EMPTY")
		}
		if member_id == self_id {
			return nil, service.NewError(400, "MEMBER_ID_INVALID")
		}
		u := biz.GetUser(member_id)
		if u == nil {
			return nil, service.NewError(400, "MEMBER_ID_INVALID")
		}
	}

	maxp := biz.DefaultMemberLimit(typ)
	chat_id := biz.CreateChat(self_id, typ, title, about, init_members, maxp)
	return chat_id, nil
}
