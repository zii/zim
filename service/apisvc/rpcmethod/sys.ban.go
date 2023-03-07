package rpcmethod

import (
	"errors"

	"golang.org/x/net/context"
	"zim.cn/biz"
	"zim.cn/biz/cookie"
	"zim.cn/biz/def"
	"zim.cn/biz/rpcx"
)

func Sys_ban(ctx context.Context, args *rpcx.BanArgs, reply *rpcx.BanReply) error {
	user_id := args.UserId
	if user_id == "" {
		return errors.New("USER_ID_EMPTY")
	}
	user := biz.GetUser(user_id)
	if user == nil {
		return errors.New("USER_ID_INVALID")
	}
	if user.Banned() {
		return nil
	}
	if !user.OK() {
		return errors.New("USER_STATUS_INVALID")
	}
	ok := biz.SetUserStatus(user_id, def.UserBanned)
	if !ok {
		return nil
	}
	cookie.ClearUserToken(user_id)
	biz.PushCmdDisconnect(user_id, "", "BANNED")
	reply.Ok = true
	return nil
}
