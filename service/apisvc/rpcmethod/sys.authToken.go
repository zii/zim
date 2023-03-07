package rpcmethod

import (
	"errors"

	"zim.cn/base/log"

	"golang.org/x/net/context"
	"zim.cn/biz"
	"zim.cn/biz/rpcx"
)

func Sys_authToken(ctx context.Context, args *rpcx.AuthTokenArgs, reply *rpcx.AuthTokenReply) error {
	user_id := args.UserId
	if user_id == "" {
		return errors.New("USER_ID_EMPTY")
	}
	platform := args.Platform
	if platform == 0 {
		return errors.New("PLATFORM_EMPTY")
	}
	device_id := args.DeviceId
	if device_id == "" {
		return errors.New("DEVICE_ID_EMPTY")
	}
	user := biz.GetUser(user_id)
	if user == nil {
		return errors.New("USER_ID_INVALID")
	}
	if user.Banned() {
		return errors.New("FORBIDDEN")
	}
	out, err := biz.Authorize(user_id, platform, device_id)
	if err != nil {
		log.Error("Authorize:", err)
		return err
	}
	reply.Token = out.Token
	return nil
}
