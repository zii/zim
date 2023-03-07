package kafka

import (
	"testing"

	"zim.cn/base"
)

func TestProducer(_ *testing.T) {
	p := NewProducer([]string{"10.10.10.86:9092"}, "top1")
	err := p.SendMessage("222", []byte("多对多"))
	base.Raise(err)
}
