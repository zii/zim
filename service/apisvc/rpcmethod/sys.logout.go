package rpcmethod

import (
	"golang.org/x/net/context"
	"zim.cn/biz"
	"zim.cn/biz/cookie"
	"zim.cn/biz/rpcx"
)

func Sys_logout(ctx context.Context, args *rpcx.LogoutArgs, reply *rpcx.LogoutReply) error {
	token := args.Token
	cv, err := cookie.Parse(token)
	if err != nil {
		return err
	}
	cookie.DelUserToken(cv.UserId, token)
	biz.PushCmdDisconnect(cv.UserId, token, "LOGOUT")
	reply.Ok = true
	return nil
}
