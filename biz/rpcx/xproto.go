package rpcx

import "zim.cn/biz/proto"

type RegisterArgs struct {
	Name  string
	Photo string
	Ex    string
}

type RegisterReply struct {
	UserId string
}

type AuthTokenArgs struct {
	UserId   string
	Platform int
	DeviceId string
}

type AuthTokenReply struct {
	Token string
}

type SendMessageArgs struct {
	Message *proto.Message
}

type SendMessageReply struct {
	Id int64
}

type DisconnectArgs struct {
	UserId string
	Token  string
	Reason string
}

type DisconnectReply struct {
	Ok bool
}

type LogoutArgs struct {
	Token string
}

type LogoutReply struct {
	Ok bool
}

type EditUserArgs struct {
	UserId string
	Name   string
	Photo  string
	Ex     string
}

type EditUserReply struct {
	Ok bool
}

type BanArgs struct {
	UserId string
}

type BanReply struct {
	Ok bool
}
