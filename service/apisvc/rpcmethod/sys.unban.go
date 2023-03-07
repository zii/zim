package rpcmethod

import (
	"errors"

	"golang.org/x/net/context"
	"zim.cn/biz"
	"zim.cn/biz/def"
	"zim.cn/biz/rpcx"
)

func Sys_unban(ctx context.Context, args *rpcx.BanArgs, reply *rpcx.BanReply) error {
	user_id := args.UserId
	if user_id == "" {
		return errors.New("USER_ID_EMPTY")
	}
	user := biz.GetUser(user_id)
	if user == nil {
		return errors.New("USER_ID_INVALID")
	}
	if user.OK() {
		return nil
	}
	if !user.Banned() {
		return errors.New("USER_STATUS_INVALID")
	}
	ok := biz.SetUserStatus(user_id, def.UserOK)
	if !ok {
		return nil
	}
	reply.Ok = true
	return nil
}
