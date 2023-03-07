package method

import (
	"fmt"
	"math/rand"

	"zim.cn/biz/cache"

	"zim.cn/base"
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/biz/kafka"
	"zim.cn/biz/proto"
	"zim.cn/service"
)

func Test_test(md *service.Meta) (interface{}, error) {
	t := md.Get("t").Int()
	switch t {
	case 1:
		id := biz.NextUserId()
		n := cache.ChannelCounter("c6").LuaSegincr(id, def.SegCounterLimit).Int()
		return n, nil
	case 2:
		m := biz.MGetMemberCount([]string{"g56", "c55"})
		return m, nil
	case 3:
		n := md.Get("n").Int()
		mode := md.Get("mode").Int()
		biz.TestPipeline(mode, n)
	case 4:
		err := kafka.P1.SendMessage(base.RandDigitCode(4), []byte("hello"))
		return "", err
	case 5:
		// benchmark
		from_id := fmt.Sprintf("u%d", 1+rand.Intn(20))
		to_id := fmt.Sprintf("u%d", 2+rand.Intn(20))
		msg := &proto.Message{
			Type:   def.MsgText,
			FromId: from_id,
			ToId:   to_id,
			Elem: &proto.Elem{
				Text: "hello",
			},
		}
		err := biz.SendMessage(msg)
		if err != nil {
			return nil, err
		}
	case 6:
		biz.InitChannelCounter("c6")
	}
	return "hi", nil
}
