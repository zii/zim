package service

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"zim.cn/base/log"

	"zim.cn/biz/cookie"

	"zim.cn/biz/cache"

	"zim.cn/base"
	"zim.cn/base/sdk"

	"zim.cn/biz/proto"

	"zim.cn/base/net"

	"github.com/gorilla/mux"
)

var ErrToken = &Error{Code: 401, Description: "登录已过期"}                           // 401
var ErrOtherDeviceOnline = &Error{Code: 410, Description: "OTHER_DEVICE_ONLINE"} // 410
var ErrAppKey = &Error{Code: 403, Description: "APPKEY_INVALID"}

const MIME_ENCRYPTED = "application/encrypted"

func (md *Meta) MergeForm() error {
	var err error
	// 20 MB max for upload files
	ct := md.ContentType()
	multipart := strings.HasPrefix(ct, "multipart/form-data")
	if multipart {
		err = md.Request.ParseMultipartForm(40 * 1024 * 1024)
		if err != nil {
			log.Error("MergeForm:", err)
			return err
		}
	} else {
		err = md.Request.ParseForm()
		if err != nil {
			log.Error("MergeForm:", err)
			return err
		}
	}

	// 解密
	if md.Encrypted() {
		cipherdata, err := ioutil.ReadAll(md.Request.Body)
		if err != nil {
			log.Error("readall:", err)
			return err
		}
		data := md.Decrypt(cipherdata)
		jsonform := make(map[string]interface{})
		json.Unmarshal(data, &jsonform)
		md.jsonform = jsonform
		return nil
	}

	//将json参数合到Form里, 云信回调和抄送用的是json body
	if !multipart {
		jsonform := make(map[string]interface{})
		d := json.NewDecoder(md.Request.Body)
		d.UseNumber()
		_ = d.Decode(&jsonform)
		md.jsonform = jsonform
	}

	return nil
}

func (md *Meta) auth(authentication bool) error {
	var err error
	var token string
	var cv *cache.CookieValue

	r := md.Request
	if authentication {
		token = r.Header.Get("token")
		cv, err = cookie.Parse(token)
	}
	ip := net.GetIP(r)

	_ = md.MergeForm()
	md.cookie = cv
	if cv != nil {
		md.UserId = cv.UserId
	}
	md.Token = token
	md.IP = ip

	if authentication {
		if err != nil {
			if err == cookie.ErrOtherDeviceOnline {
				return ErrOtherDeviceOnline
			}
			return ErrToken
		}
		if md.UserId == "" {
			return ErrToken
		}
	}

	return nil
}

func (md *Meta) WriteValue(w http.ResponseWriter, response interface{}) error {
	if md.Encrypted() {
		w.Header().Set("Content-Type", MIME_ENCRYPTED)
		d, _ := json.Marshal(response)
		ciphertext := md.Encrypt(string(d))
		w.Write([]byte(ciphertext))
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func (md *Meta) WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var fail *proto.Fail
	switch e := err.(type) {
	case *Error:
		fail = &proto.Fail{Code: e.Code, Msg: e.Description, Data: e.Data}
	default:
		fail = &proto.Fail{Code: 500, Msg: "INTERNAL_ERROR"}
	}
	md.WriteValue(w, fail)
}

func toSuccess(result interface{}) *proto.Success {
	/* 构造成功json */
	var out = &proto.Success{Code: 200}
	out.Data = result
	out.Msg = "success"
	return out
}

func toFail(err error) *proto.Fail {
	var code = 500
	var description = "INTERAL_ERROR"
	var data interface{}

	var out = &proto.Fail{}

	switch err.(type) {
	case *Error:
		e := err.(*Error)
		code = e.Code
		description = e.Description
		data = e.Data
	default:
		panic(err)
	}
	log.Println("✖", code, description)
	out.Code = code
	out.Msg = description
	out.Data = data
	return out
}

type Decorator func(*Meta) error

type ErrorHandler struct {
}

func (h *ErrorHandler) Handle(ctx context.Context, err error) {
	log.Error("Service:", err)
}

// 当前活动请求数
var activeRequests int32

func ActiveRequests() int {
	return int(activeRequests)
}

func WaitQuiescent(timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	const pollInterval = 500 * time.Millisecond
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		if activeRequests <= 0 {
			return nil
		}
		select {
		case <-ticker.C:
		case <-timer.C:
			return context.DeadlineExceeded
		}
	}
}

func (md *Meta) checkHeaderSign() error {
	// 检查请求头
	open := false
	//open := base.RELEASE && req.Header.Get("pwd") != "XpeZaK7TCR"
	if open {
		req := md.Request
		appkey := req.Header.Get("appkey")
		appinfo := sdk.GetAppInfo(appkey)
		if appinfo == nil {
			return NewError(403, "APPKEY_INVALID")
		}
		appsec := appinfo.Secret
		if appsec == "" {
			return NewError(403, "APPSECRET_INVALID")
		}
		ts := req.Header.Get("timestamp")
		timestamp, _ := strconv.ParseInt(ts, 10, 64)
		if timestamp == 0 {
			return NewError(403, "TIMESTAMP_INVALID")
		}
		now := time.Now().Unix()
		if base.AbsInt64(now-timestamp) > 5*60 {
			return NewError(403, "TIMESTAMP_INVALID")
		}
		sign := req.Header.Get("sign")
		if sign == "" {
			return NewError(403, "SIGN_EMPTY")
		}
		path := req.URL.Path
		sign2 := base.Sha1String([]byte(path + ts + appsec))
		if sign != sign2 {
			return NewError(403, "SIGN_INVALID")
		}
	}
	return nil
}

// authentication: 是否需要登录
func RegisterMethod(router *mux.Router, path string, method Method, authentication bool, decorators ...Decorator) {
	endpoint := func(md *Meta) (response interface{}, err error) {
		atomic.AddInt32(&activeRequests, 1)
		defer func() {
			atomic.AddInt32(&activeRequests, -1)
			md.Log()
		}()
		r, err := method(md)
		if err != nil {
			return toFail(err), nil
		} else {
			return toSuccess(r), nil
		}
	}
	dec := func(md *Meta) error {
		for _, dec := range decorators {
			err := dec(md)
			if err != nil {
				return err
			}
		}
		return nil
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")  //允许访问所有域
		w.Header().Add("Access-Control-Allow-Headers", "*") //header的类型
		md := NewMeta(r)
		err := md.checkHeaderSign()
		if err != nil {
			md.WriteError(w, err)
			return
		}
		err = md.auth(authentication)
		if err != nil {
			md.Log()
			md.WriteError(w, err)
			return
		}
		err = dec(md)
		if err != nil {
			md.WriteError(w, err)
			return
		}
		rsp, err := endpoint(md)
		if err != nil {
			md.WriteError(w, err)
			return
		}
		err = md.WriteValue(w, rsp)
		if err != nil {
			md.WriteError(w, err)
			return
		}
	}

	//handler := kithttp.NewServer(endpoint, dec, onsuccess, opts...)
	router.HandleFunc(path, handler)
}

// 半原生回调，用于第三方平台与服务器的表单回调, 不适用body格式的请求(比如XML)
func RegisterMethodRaw(router *mux.Router, path string, f func(*Meta, http.ResponseWriter)) {
	dec := func(w http.ResponseWriter, r *http.Request) {
		md := NewMeta(r)
		err := md.auth(false)
		if err != nil {
			log.Error("500 MD_ERROR:", err)
			http.Error(w, "MD_ERROR", 500)
			return
		}
		atomic.AddInt32(&activeRequests, 1)
		defer func() {
			atomic.AddInt32(&activeRequests, -1)
			md.LogRaw()
		}()
		f(md, w)
	}
	router.HandleFunc(path, dec)
}

// 原生回调，用于第三方平台与服务器的回调
func RegisterHandler(router *mux.Router, path string, f func(*http.Request, http.ResponseWriter)) {
	dec := func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&activeRequests, 1)
		st := time.Now()
		defer func() {
			atomic.AddInt32(&activeRequests, -1)
			ip := net.GetIP(r)
			log.Println("<=", r.RequestURI, ip, time.Since(st))
		}()
		f(r, w)
	}
	router.HandleFunc(path, dec)
}
