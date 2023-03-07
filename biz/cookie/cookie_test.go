package cookie

import (
	"fmt"
	"strings"
	"testing"
)

func TestValid(t *testing.T) {
	// 测试自己签名自己验证
	//s, err := Sign(1, nil)
	//base.Raise(err)
	//fmt.Println("token:", s)
	//
	//user_id, err := Parse(s)
	//base.Raise(err)
	//fmt.Println("user_id:", user_id)
	v := "1.0.0-35"
	tp := strings.SplitN(v, "-", 2)
	fmt.Println("tp:", tp, len(tp))
}
