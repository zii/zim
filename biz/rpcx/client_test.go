package rpcx

import (
	"fmt"
	"testing"

	"zim.cn/biz/def"

	"zim.cn/biz/proto"

	"zim.cn/base"
)

func TestRegister(_ *testing.T) {
	user_id, err := Register("cat", "", "xx")
	base.Raise(err)
	fmt.Println("register:", user_id)
}

func TestAuthToken(_ *testing.T) {
	a, err := AuthToken("u1", 1, "d1")
	base.Raise(err)
	fmt.Println("AuthToken:", a.Token)
}

func TestSendMessage(_ *testing.T) {
	msg := &proto.Message{
		FromId: "u1",
		ToId:   "u44",
		Type:   def.MsgText,
		Elem: &proto.Elem{
			Text: "hi",
		},
	}
	id, err := SendMessage(msg)
	base.Raise(err)
	fmt.Println("SendMessage:", id)
}

func TestDisconnect(_ *testing.T) {
	ok, err := Disconnect("u44", "", "hehe")
	base.Raise(err)
	fmt.Println("Disconnect:", ok)
}

func TestLogout(_ *testing.T) {
	ok, err := Logout("0164d75cae1e1d9972")
	base.Raise(err)
	fmt.Println("logout:", ok)
}

func TestEditUser(_ *testing.T) {
	ok, err := EditUser("u1", "用户1", "xx", "ee")
	base.Raise(err)
	fmt.Println("editUser:", ok)
}

func TestBan(_ *testing.T) {
	ok, err := Ban("u1")
	base.Raise(err)
	fmt.Println("ban:", ok)
}

func TestUnban(_ *testing.T) {
	ok, err := Unban("u1")
	base.Raise(err)
	fmt.Println("unban:", ok)
}
