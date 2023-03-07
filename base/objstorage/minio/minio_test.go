package minio

import (
	"fmt"
	"testing"

	"zim.cn/base/uuid"

	"zim.cn/base"
)

func Test1(_ *testing.T) {
	uuid.InitUUID()
	r, err := CreateObjectWithURL(Bucket, "", "https://lzf-p1.oss-cn-guangzhou.aliyuncs.com/v2-177/1bmjxku4hckz.jpg")
	base.Raise(err)
	fmt.Println("r:", base.JsonString(r))
}
