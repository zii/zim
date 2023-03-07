package main

import (
	"encoding/json"
	"sync"
	"time"

	"zim.cn/base/log"

	"zim.cn/biz/proto"

	redigo "github.com/gomodule/redigo/redis"
	"zim.cn/base/redis"

	"zim.cn/biz/def"
)

type Devices map[string]*Client

type Hub struct {
	devices Devices // {token:client}
	dev_mu  sync.RWMutex
	users   map[string]Devices // {user_id:[client]}}
	usr_mu  sync.RWMutex
}

var hub *Hub

func init() {
	hub = &Hub{
		devices: make(Devices),
		users:   make(map[string]Devices),
	}
}

func (h *Hub) setDeviceNX(c *Client) bool {
	h.dev_mu.Lock()
	_, ok := h.devices[c.token]
	if ok {
		h.dev_mu.Unlock()
		return false
	}
	h.devices[c.token] = c
	h.dev_mu.Unlock()
	return true
}

func (h *Hub) register(c *Client) bool {
	// 连接策略: 阻止后来的相同设备连接
	ok := h.setDeviceNX(c)
	if !ok {
		c.Kick("DUPLICATED")
		return false
	}

	h.usr_mu.Lock()
	ds := h.users[c.user_id]
	if ds == nil {
		ds = make(Devices)
	}
	ds[c.token] = c
	h.users[c.user_id] = ds
	h.usr_mu.Unlock()
	return true
}

// 返回false表示该连接不存在
func (h *Hub) delDevice(c *Client) bool {
	h.dev_mu.Lock()
	defer h.dev_mu.Unlock()
	old := h.devices[c.token]
	if old != nil && old.conn_id != c.conn_id {
		return false
	}
	delete(h.devices, c.token)
	return true
}

func (h *Hub) delUserDevice(c *Client) {
	h.usr_mu.Lock()
	ds := h.users[c.user_id]
	delete(ds, c.token)
	if len(ds) == 0 {
		delete(h.users, c.user_id)
	}
	h.usr_mu.Unlock()
}

func (h *Hub) unregister(c *Client) {
	ok := h.delDevice(c)
	if !ok {
		return
	}
	h.delUserDevice(c)
}

func (h *Hub) getClientsOfChannel(chat_id string) []*Client {
	var out []*Client
	h.dev_mu.RLock()
	for _, c := range h.devices {
		if c.InChannel(chat_id) {
			out = append(out, c)
		}
	}
	h.dev_mu.RUnlock()
	return out
}

func (h *Hub) getClientsOfUser(user_ids ...string) []*Client {
	var out []*Client
	h.usr_mu.RLock()
	for _, user_id := range user_ids {
		ds := h.users[user_id]
		for _, c := range ds {
			out = append(out, c)
		}
	}
	h.usr_mu.RUnlock()
	return out
}

func (h *Hub) addUserChannel(user_id string, chat_id string) {
	if def.ToIdType(chat_id) != def.IdChannel {
		return
	}
	clients := h.getClientsOfUser(user_id)
	for _, c := range clients {
		c.AddChannel(chat_id)
	}
}

func (h *Hub) deleteUserChannel(user_id string, chat_id string) {
	if def.ToIdType(chat_id) != def.IdChannel {
		return
	}
	clients := h.getClientsOfUser(user_id)
	for _, c := range clients {
		c.DeleteChannel(chat_id)
	}
}

func (h *Hub) beforeSend(msg *proto.Message) {
	switch msg.Type {
	case def.TipChatCreated:
		if msg.Tip != nil && msg.Tip.ChatCreated != nil {
			tip := msg.Tip.ChatCreated
			h.addUserChannel(tip.Creator.Id, tip.Chat.Id)
			for _, u := range tip.InitMembers {
				h.addUserChannel(u.Id, tip.Chat.Id)
			}
		}
	case def.TipMemberEnter:
		if msg.Tip != nil && msg.Tip.MemberEnter != nil {
			tip := msg.Tip.MemberEnter
			for _, user := range tip.Users {
				h.addUserChannel(user.Id, tip.ChatId)
			}
		}
	case def.TipMemberQuit:
		if msg.Tip != nil && msg.Tip.MemberQuit != nil {
			tip := msg.Tip.MemberQuit
			h.deleteUserChannel(tip.User.Id, tip.ChatId)
		}
	}
}

func (h *Hub) afterSend(msg *proto.Message) {
	switch msg.Type {
	case def.TipMemberKicked:
		if msg.Tip != nil && msg.Tip.MemberKicked != nil {
			tip := msg.Tip.MemberKicked
			h.deleteUserChannel(tip.User.Id, tip.ChatId)
		}
	}
}

func (h *Hub) handleCommand(cmd *proto.Command) {
	if cmd.Op == def.OpSend {
		msg := cmd.Message
		if msg == nil {
			return
		}
		h.beforeSend(msg)
		tot := def.ToIdType(msg.ToId)
		var clients []*Client
		if tot == def.IdChannel {
			clients = h.getClientsOfChannel(msg.ToId)
		} else {
			clients = h.getClientsOfUser(cmd.UserIds...)
		}
		for _, c := range clients {
			c.SendMessage(msg)
		}
		h.afterSend(msg)
	} else if cmd.Op == def.OpDisconnect {
		if cmd.CmdDisconnect == nil {
			return
		}
		data := cmd.CmdDisconnect
		if data == nil {
			return
		}
		clients := h.getClientsOfUser(data.UserId)
		for _, c := range clients {
			if data.Token == "" || c.token == data.Token {
				c.Kick(data.Reason)
				c.Close()
			}
		}
	}
}

func (h *Hub) subscribe() {
	c := redis.GetRedisPoolClient("pubsub")
	sub := redigo.PubSubConn{Conn: c}
	err := sub.Subscribe(def.MESSAGE_CHANNEL)
	if err != nil {
		log.Error("Subscribe:", err)
		return
	}

	for {
		v := sub.Receive()
		switch r := v.(type) {
		case redigo.Message:
			log.Println("➡", string(r.Data))
			var cmd *proto.Command
			err := json.Unmarshal(r.Data, &cmd)
			if err != nil {
				log.Error("unmarshal cmd:", err)
			} else {
				h.handleCommand(cmd)
			}
		case redigo.Subscription:
			log.Println("subscr:", r.Channel, r.Kind, r.Count)
		case error:
			log.Error("receive:", err)
			return
		}
	}
}

func (h *Hub) run() {
	for {
		h.subscribe()
		time.Sleep(10 * time.Second)
	}
}
