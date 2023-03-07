package method

import (
	"zim.cn/base/log"
	"zim.cn/biz"
	"zim.cn/service"
)

// Account_updateNotifySetting godoc
// @Summary      更新对话免打扰设置
// @Description  失败响应:
// @Description  400 PEER_ID_INVALID 对话不存在
// @Description  400 SETTING_INVALID 参数解析错误
// @Tags         用户体系
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Param peer_id formData string true "对话ID"
// @Param setting formData string true "免打扰设置" default({"badge":true})
// @Success 200 {object} proto.Success{data=bool}
// @Router       /account/updateNotifySetting [post]
func Account_updateNotifySetting(md *service.Meta) (any, error) {
	self_id := md.UserId
	peer_id := md.Get("peer_id").String()
	if peer_id == "" {
		return nil, service.NewError(400, "PEER_ID_EMPTY")
	}
	if !biz.IsDialogExists(self_id, peer_id) {
		return nil, service.NewError(400, "PEER_ID_INVALID")
	}
	var input *biz.PeerNotifySetting
	err := md.Get("setting").Unmarshal(&input)
	if err != nil {
		log.Error(err)
		return nil, service.NewError(400, "SETTING_INVALID")
	}
	ok := biz.UpdatePeerNotifySetting(self_id, peer_id, input)
	return ok, nil
}
