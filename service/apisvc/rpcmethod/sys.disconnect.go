package rpcmethod

import (
	"errors"

	"golang.org/x/net/context"
	"zim.cn/biz"
	"zim.cn/biz/rpcx"
)

func Sys_disconnect(ctx context.Context, args *rpcx.DisconnectArgs, reply *rpcx.DisconnectReply) error {
	if args.UserId == "" {
		return errors.New("USER_ID_EMPTY")
	}
	biz.PushCmdDisconnect(args.UserId, args.Token, args.Reason)
	reply.Ok = true
	return nil
}
