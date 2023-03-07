package rpcx

import (
	"context"

	"zim.cn/biz/proto"

	"github.com/smallnest/rpcx/client"
)

var ServiceAddr = "127.0.0.1:1850"
var ServicePath = "sys"
var xclient client.XClient

func init() {
	Init()
}

func Init() {
	d, _ := client.NewPeer2PeerDiscovery("tcp@"+ServiceAddr, "")
	xclient = client.NewXClient(ServicePath, client.Failtry, client.RandomSelect, d, client.DefaultOption)
}

// 注册新用户
func Register(name, photo, ex string) (string, error) {
	args := &RegisterArgs{
		Name:  name,
		Photo: photo,
		Ex:    ex,
	}
	reply := &RegisterReply{}
	err := xclient.Call(context.Background(), "register", args, reply)
	if err != nil {
		return "", err
	}
	return reply.UserId, nil
}

// 生成token
func AuthToken(user_id string, platform int, device_id string) (*proto.Authorization, error) {
	args := &AuthTokenArgs{
		UserId:   user_id,
		Platform: platform,
		DeviceId: device_id,
	}
	reply := &AuthTokenReply{}
	err := xclient.Call(context.Background(), "authToken", args, reply)
	if err != nil {
		return nil, err
	}
	out := &proto.Authorization{
		Token: reply.Token,
	}
	return out, nil
}

// 后台发送消息
func SendMessage(msg *proto.Message) (int64, error) {
	args := &SendMessageArgs{
		Message: msg,
	}
	reply := &SendMessageReply{}
	err := xclient.Call(context.Background(), "sendMessage", args, reply)
	if err != nil {
		return 0, err
	}
	return reply.Id, nil
}

// 强制短线
func Disconnect(user_id string, token string, reason string) (bool, error) {
	args := &DisconnectArgs{
		UserId: user_id,
		Token:  token,
		Reason: reason,
	}
	reply := &DisconnectReply{}
	err := xclient.Call(context.Background(), "disconnect", args, reply)
	if err != nil {
		return false, err
	}
	return reply.Ok, nil
}

// 强制退出登录
func Logout(token string) (bool, error) {
	args := &LogoutArgs{
		Token: token,
	}
	reply := &LogoutReply{}
	err := xclient.Call(context.Background(), "logout", args, reply)
	if err != nil {
		return false, err
	}
	return reply.Ok, nil
}

// 调用修改用户资料
func EditUser(user_id, name, photo, ex string) (bool, error) {
	args := &EditUserArgs{
		UserId: user_id,
		Name:   name,
		Photo:  photo,
		Ex:     ex,
	}
	reply := &EditUserReply{}
	err := xclient.Call(context.Background(), "editUser", args, reply)
	if err != nil {
		return false, err
	}
	return reply.Ok, nil
}

// 封禁用户
func Ban(user_id string) (bool, error) {
	args := &BanArgs{
		UserId: user_id,
	}
	reply := &BanReply{}
	err := xclient.Call(context.Background(), "ban", args, reply)
	if err != nil {
		return false, err
	}
	return reply.Ok, nil
}

// 解封用户
func Unban(user_id string) (bool, error) {
	args := &BanArgs{
		UserId: user_id,
	}
	reply := &BanReply{}
	err := xclient.Call(context.Background(), "unban", args, reply)
	if err != nil {
		return false, err
	}
	return reply.Ok, nil
}
