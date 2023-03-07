package method

import (
	"zim.cn/biz"
	"zim.cn/service"
)

func Message_getDialogs(md *service.Meta) (any, error) {
	user_id := md.UserId
	offset := md.Get("offset").Int()
	limit := md.Get("limit").Int()
	if err := biz.PagingVerify(&offset, &limit, 100); err != nil {
		return nil, service.NewError(400, err.Error())
	}
	out := biz.SearchTLDialogs(user_id, offset, limit)
	return out, nil
}
