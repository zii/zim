package method

import (
	"zim.cn/base/objstorage"
	"zim.cn/biz"
	"zim.cn/service"
)

// Upload_getCredential godoc
// @Summary      获取直传凭证
// @Description  Token自颁发后将在一段时间内有效(timeout), 并在有效期内重复使用
// @Description  失败响应:
// @Description  400 异常信息
// @Tags         其他
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param token header string true "授权token"
// @Success 200 {object} proto.Success{data=proto.Credential} "返回直传凭证"
// @Router       /upload/getCredential [post]
func Upload_getCredential(md *service.Meta) (any, error) {
	p := md.Get("platform").String()
	if p == "" {
		p = objstorage.Platform
	}
	out, err := biz.GetTLCredential(p, md.UserId)
	if err != nil {
		return nil, service.NewError(400, err.Error())
	}
	return out, nil
}
