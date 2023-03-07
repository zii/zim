package rpcmethod

import (
	"errors"

	"golang.org/x/net/context"
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/biz/rpcx"
)

func Sys_sendMessage(ctx context.Context, args *rpcx.SendMessageArgs, reply *rpcx.SendMessageReply) error {
	msg := args.Message
	if msg == nil {
		return errors.New("MESSAGE_INVALID")
	}
	if msg.Type.Class() != def.Msg {
		return errors.New("TYPE_INVALID")
	}
	if msg.FromId == "" {
		return errors.New("FROM_ID_EMPTY")
	}
	if msg.ToId == "" {
		return errors.New("TO_ID_EMPTY")
	}
	err := biz.SendMessage(msg)
	if err != nil {
		return err
	}
	reply.Id = msg.Id
	return nil
}
