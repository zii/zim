package objstorage

import (
	"fmt"
	"testing"

	"zim.cn/base/uuid"

	"zim.cn/base"
)

func Test1(_ *testing.T) {
	Platform = "cos"
	uuid.InitUUID()
	r, err := CreateObjectWithURL("", "https://www.baidu.com/img/PCtm_d9c8750bed0b3c7d089fa7d55720d6cf.png")
	base.Raise(err)
	fmt.Println("r:", base.JsonString(r))
	fmt.Println("url:", r.Url())
}
