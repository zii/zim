package main

import (
	"time"

	"zim.cn/base/aes"

	"zim.cn/base/sdk"

	"zim.cn/base/log"

	"zim.cn/base"

	"zim.cn/biz/proto"

	"zim.cn/biz/def"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 8192
)

type Client struct {
	conn_id  int64
	conn     *websocket.Conn
	ip       string
	send     chan []byte
	user_id  string
	appinfo  *sdk.AppInfo
	token    string // 既是token也是设备唯一ID
	platform def.Platform
	channels []string // 所在超级群 [chat_id]
}

func (c *Client) Json() string {
	type TLClient struct {
		ConnId   int64        `json:"conn_id"`
		Ip       string       `json:"ip"`
		UserId   string       `json:"user_id"`
		Token    string       `json:"token"`
		Platform def.Platform `json:"platform"`
		Channels []string     `json:"channels"`
	}
	out := &TLClient{
		ConnId:   c.conn_id,
		Ip:       c.ip,
		UserId:   c.user_id,
		Token:    c.token,
		Platform: c.platform,
		Channels: c.channels,
	}
	return base.JsonString(out)
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		close(c.send)
		hub.unregister(c)
		log.Println("close read", c.conn.RemoteAddr(), c.user_id, c.platform)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("unexpected error: %v", err)
			}
			break
		}
		_ = message
	}
}

func (c *Client) encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := aes.GcmEncrypt(plaintext, []byte(c.appinfo.Secret), []byte(c.token[:12]))
	return ciphertext, err
}

func (c *Client) Write(plaintext []byte) error {
	var data []byte
	if c.appinfo != nil && c.appinfo.Secret != "" {
		ciphertext, err := c.encrypt(plaintext)
		if err != nil {
			log.Error("encrypt:", err)
			return err
		}
		data = ciphertext
	} else {
		data = plaintext
	}
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		//log.Println("close write", c.conn.RemoteAddr())
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.Write(message)
			if err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			// 定时检查是否还在连接池
		}
	}
}

func (c *Client) Kick(reason string) {
	msg := &proto.Message{
		Type: def.EvDisconnect,
		Event: &proto.Event{
			Disconnect: &proto.EvDisconnect{
				Reason: reason,
			},
		},
	}
	b := msg.Blob()
	//c.send <- b
	c.Write(b)
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
	log.Println("kick conn:", c.user_id, c.token, reason)
}

func (c *Client) InChannel(chat_id string) bool {
	return base.InStrings(chat_id, c.channels)
}

func (c *Client) AddChannel(chat_id string) {
	c.channels = base.ArrayAdd(c.channels, chat_id)
}

func (c *Client) DeleteChannel(chat_id string) bool {
	c.channels = base.ArrayRemove(c.channels, chat_id)
	return true
}

func (c *Client) SendMessage(msg *proto.Message) {
	blob := msg.Blob()
	c.send <- blob
}

func (c *Client) SendBlob(b []byte) {
	c.send <- b
}

func (c *Client) Close() {
	c.conn.Close()
}
