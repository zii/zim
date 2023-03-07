package main

import (
	"fmt"
	"testing"

	"zim.cn/base"
)

func Test1(_ *testing.T) {
	o, err := ParseFile("/Users/cat/code/lzf/csim/doc/API.txt")
	base.Raise(err)
	//fmt.Println("o:", FunctionToMd(o.Functions[2]))
	fmt.Println("doc:", DocToMd(o))
}

func Test2(_ *testing.T) {
	comment := `登录成功
    accid: aaaa
    im_token: xxxx
    im_timeout: yyyy
    token: zzzz`
	i := findCommentKey(comment, "token:")
	if i > 0 {
		fmt.Println("context:", comment[i:i+10])
	}
}
