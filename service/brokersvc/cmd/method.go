package main

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"zim.cn/base"

	"zim.cn/service"
)

func Node_statics(md *service.Meta, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	var b = strings.Builder{}
	b.WriteString(fmt.Sprintf("user_count: %d\n", len(hub.users)))
	b.WriteString(fmt.Sprintf("device_count: %d\n", len(hub.devices)))
	b.WriteString(fmt.Sprintf("goroutine num: %d\n", runtime.NumGoroutine()))
	b.WriteString("\n")
	hub.dev_mu.RLock()
	for user_id, ds := range hub.users {
		b.WriteString(fmt.Sprintf("user: %s\n", user_id))
		for _, c := range ds {
			b.WriteString(fmt.Sprintf("  token: %s, user_id: %s, conn_id:%d, ip:%s, platform:%s, channels:%s\n",
				c.token, c.user_id, c.conn_id, c.ip, c.platform, base.JsonString(c.channels)))
		}
	}
	hub.dev_mu.RUnlock()
	out := b.String()
	w.Write([]byte(out))
}
