package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"
	"zim.cn/biz/def"
	"zim.cn/biz/proto"

	"github.com/tidwall/gjson"

	"zim.cn/base"
)

var APIHOST = "https://10.10.10.86:1840"

var client *http.Client

func init() {
	client = newClient()
}

func newClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		IdleConnTimeout:     300 * time.Second,
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 0,
		MaxConnsPerHost:     0,
	}
	if err := http2.ConfigureTransport(transport); err != nil {
		panic(err)
	}
	c := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	return c
}

func post(path string, param url.Values, token string) gjson.Result {
	url2 := APIHOST + path
	payload := strings.NewReader(param.Encode())
	req, err := http.NewRequest("POST", url2, payload)
	base.Raise(err)
	req.Header.Set("token", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r, err := client.Do(req)
	base.Raise(err)
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	base.Raise(err)
	return gjson.ParseBytes(data)
}

func sendmessage(token, from_id, to_id, text string) {
	param := url.Values{}
	msg := &proto.Message{
		FromId: from_id,
		ToId:   to_id,
		Type:   def.MsgText,
		Elem: &proto.Elem{
			Text: text,
		},
	}
	msgs := base.JsonString(msg)
	param.Set("message", msgs)
	r := post("/v1/message/sendMessage", param, token)
	_ = r
	//fmt.Println("r:", r.String())
}

func testSendmessage(co, n int) {
	wg := &sync.WaitGroup{}
	wg.Add(co)
	for i := 0; i < co; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < n; j++ {
				to_id := fmt.Sprintf("u%d", 2+j%10)
				sendmessage("013b6d578f65171a82", "u1", to_id, "jajajaja")
				time.Sleep(1 * time.Second)
			}
		}()
	}
	wg.Wait()
}

func main() {
	var co int
	var n int
	flag.IntVar(&co, "c", 1000, "")
	flag.IntVar(&n, "n", 10, "")
	flag.Parse()
	fmt.Println("args:", co, n)
	st := time.Now()
	testSendmessage(co, n)
	fmt.Println("took:", time.Since(st))
}
