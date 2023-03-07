package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"zim.cn/base"
)

func decodeGuid(content string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}
	for i := range data {
		data[i] = data[i] >> 1
	}
	return string(data), nil
}

func decodeFile() string {
	content, err := ioutil.ReadFile("/Users/mac/code/lzf.com/zim-server/test/test-licence/lincence")
	base.Raise(err)
	guid, err := decodeGuid(string(content))
	base.Raise(err)
	fmt.Println("guid:", guid)
	return guid
}

func main() {
	decodeFile()
}
