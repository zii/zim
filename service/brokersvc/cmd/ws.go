package main

import (
	"net/http"
	"time"

	"zim.cn/base/sdk"

	"zim.cn/base/log"

	"zim.cn/biz"

	"zim.cn/base/uuid"

	"zim.cn/base/net"

	"zim.cn/service"

	"zim.cn/biz/def"

	"zim.cn/biz/cookie"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true,
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ip := net.GetIP(r)
	log.Println("⬅", ip, r.RequestURI)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("init ws:", err)
		return
	}
	client := &Client{
		conn_id: uuid.NextID("conn"),
		conn:    conn,
		ip:      ip,
		send:    make(chan []byte, 256),
	}
	go client.writePump()
	go client.readPump()

	r.ParseForm()
	md := service.NewMeta(r)
	token := md.Get("token").String()
	ok := authorize(client, token)
	if !ok {
		time.Sleep(10 * time.Second)
		conn.Close()
		return
	}
	ok = hub.register(client)
	if !ok {
		time.Sleep(10 * time.Second)
		conn.Close()
		return
	}
	client.channels = biz.GetChannelIdsOfUser(client.user_id)
	log.Println("login success:", client.user_id, client.token, client.platform)
}

func authorize(c *Client, token string) bool {
	cv, err := cookie.Parse(token)
	if err != nil {
		log.Error("authorize:", err)
		c.Kick("TOKEN_INVALID")
		return false
	}
	c.token = token
	c.user_id = cv.UserId
	c.appinfo = sdk.GetAppInfo(cv.Appkey)
	c.platform = def.Platform(cv.Platform)
	return true
}
