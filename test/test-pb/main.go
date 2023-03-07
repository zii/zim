// Package pb contains generated protobuf types
//
//go:generate protoc -I=. --gogofaster_out=. api.proto
package main

// jsonpb 不堪使用:
// 1. 速度慢, 10x slow than json
// 2. int64居然会编码为字符串

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/tidwall/gjson"

	"zim.cn/base"

	"github.com/golang/protobuf/proto"
)

func test1() {
	m := &jsonpb.Marshaler{}
	_ = m
	d := &Message{
		Id:     1,
		FromId: "u1",
		ToId:   "u2",
		Type:   1,
		Elem: &Message_Photo{
			Photo: &PhotoElem{
				Big: "xxxx",
			},
		},
	}
	var s string
	var err error
	// 每秒10w次
	s, err = m.MarshalToString(d)
	base.Raise(err)
	fmt.Println("s:", s)
	var n Message
	// 每秒10w次
	err = jsonpb.UnmarshalString(s, &n)
	base.Raise(err)
	fmt.Println("n:", &n, n.GetPhoto())
}

type Elem interface {
	Elem()
}

type ElemText struct {
	Text string `json:"text"`
}

func (e *ElemText) Elem() {}

type ElemPhoto struct {
	Big   string `json:"big"`
	Small string `json:"small"`
}

func (e *ElemPhoto) Elem() {}

type Msg struct {
	Elem Elem `json:"elem,omitempty"`
}

func test2() {
	m := &Msg{
		Elem: &ElemPhoto{
			Big: "xxxx",
		},
	}
	s := base.JsonString(m)
	fmt.Println("encode:", s)
	//var out *Msg
	st := time.Now()
	var o gjson.Result
	var out string
	for i := 0; i < 1000000; i++ {
		o = gjson.Parse(s)
		out = o.Get("elem").Raw
		var n *Msg
		err := json.Unmarshal([]byte(o.Raw), &n)
		base.Raise(err)
	}
	fmt.Println("took:", time.Since(st))
	fmt.Println("o:", out)
}

func test3() {
	m := &Message{
		Id:     1,
		FromId: "u1",
		Type:   1,
	}
	var b []byte
	var err error
	st := time.Now()
	for i := 0; i < 1000000; i++ {
		b, err = proto.Marshal(m)
		base.Raise(err)
		err = proto.Unmarshal(b, m)
		base.Raise(err)
	}
	fmt.Println("took:", time.Since(st))
	fmt.Println("b:", b)
}

func test4() {
	m := &Message{
		Id:     1,
		FromId: "u1",
		Type:   1,
	}
	var s string
	st := time.Now()
	for i := 0; i < 1000000; i++ {
		s = base.JsonString(m)
		//err := json.Unmarshal([]byte(s), m)
		//base.Raise(err)
	}
	fmt.Println("took:", time.Since(st))
	fmt.Println("s:", s)
}

func main() {
	test1()
}
