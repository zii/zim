package biz

import (
	"zim.cn/service"
)

// 检查分页参数
func PagingVerify(offset, limit *int, limit_max int) error {
	if *offset < 0 {
		*offset = 0
	}
	if *limit < 0 {
		*limit = 0
	}
	if *limit > limit_max {
		return service.NewError(400, "LIMIT_EXCEED")
	}
	return nil
}
