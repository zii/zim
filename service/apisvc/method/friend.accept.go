package method

import (
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/service"
)

// Friend_accept godoc
// @Summary      接受好友申请
// @Description  失败响应:
// @Description  400 LIMIT_EXCEEDED 分页参数错误
// @Tags         好友(通讯录)管理
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param hash formData string true "好友申请记录唯一标识"
// @Success 200 {object} proto.Success{data=bool} "返回是否有变更"
// @Router       /friend/accept [post]
func Friend_accept(md *service.Meta) (any, error) {
	self_id := md.UserId
	hash := md.Get("hash").String()
	if hash == "" {
		return nil, service.NewError(400, "HASH_EMPTY")
	}
	fn := biz.GetFriendCount(self_id)
	if fn >= def.FriendLimit {
		return nil, service.NewError(400, "LIMIT_EXCEEDED")
	}
	ok, err := biz.AcceptFriend(self_id, hash)
	if err != nil {
		return nil, service.NewError(400, err.Error())
	}
	return ok, nil
}
