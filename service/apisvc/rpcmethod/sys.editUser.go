package rpcmethod

import (
	"errors"

	"golang.org/x/net/context"
	"zim.cn/biz"
	"zim.cn/biz/rpcx"
)

func Sys_editUser(ctx context.Context, args *rpcx.EditUserArgs, reply *rpcx.EditUserReply) error {
	user_id := args.UserId
	if user_id == "" {
		return errors.New("USER_ID_EMPTY")
	}
	user := biz.GetUser(user_id)
	if user == nil {
		return errors.New("USER_ID_INVALID")
	}
	ok := biz.EditUser(user, args.Name, args.Photo, args.Ex)
	reply.Ok = ok
	return nil
}
