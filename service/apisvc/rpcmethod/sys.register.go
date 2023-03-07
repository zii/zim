package rpcmethod

import (
	"errors"

	"golang.org/x/net/context"
	"zim.cn/biz"
	"zim.cn/biz/rpcx"
)

func Sys_register(ctx context.Context, args *rpcx.RegisterArgs, reply *rpcx.RegisterReply) error {
	if args.Name == "" {
		return errors.New("NAME_EMPTY")
	}
	user_id := biz.CreateUser(args.Name, args.Photo, args.Ex)
	reply.UserId = user_id
	return nil
}
