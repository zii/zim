package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zim.cn/base/aes"

	"zim.cn/base/log"

	"zim.cn/base"

	"zim.cn/base/sdk"
	"zim.cn/biz/cache"

	"github.com/tidwall/gjson"
)

type Result struct {
	s string
	v interface{}
}

func (this Result) Exists() bool {
	return this.s != "" || this.v != nil
}

func (this Result) String() string {
	return this.s
}

func (this Result) Bool() bool {
	s := this.s
	if s != "" {
		return s == "1" || s == "true" || s == "True"
	}
	if this.v != nil {
		v, _ := this.v.(bool)
		return v
	}
	return false
}

func (this Result) Int() int {
	s := this.s
	if s != "" {
		i, _ := strconv.Atoi(s)
		return i
	}
	if this.v != nil {
		v, _ := this.v.(int)
		return v
	}
	return 0
}

func (this Result) Int32() int32 {
	s := this.s
	if s != "" {
		i, _ := strconv.Atoi(s)
		return int32(i)
	}
	if this.v != nil {
		v, _ := this.v.(int32)
		return v
	}
	return 0
}

func (this Result) Int64() int64 {
	s := this.s
	if s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		return i
	}
	if this.v != nil {
		v, _ := this.v.(int64)
		return v
	}
	return 0
}

func (this Result) Float64() float64 {
	s := this.s
	if s != "" {
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	if this.v != nil {
		f, _ := this.v.(float64)
		return f
	}
	return 0
}

func (this Result) Json() gjson.Result {
	return gjson.Parse(this.s)
}

func (this Result) Unmarshal(v interface{}) error {
	err := json.Unmarshal([]byte(this.s), v)
	return err
}

func (this Result) Strings() []string {
	var out = []string{}
	if this.s != "" {
		this.Unmarshal(&out)
		return out
	}
	if this.v != nil {
		rows, _ := this.v.([]interface{})
		for _, v := range rows {
			s, _ := v.(string)
			out = append(out, s)
		}
	}
	return out
}

func (this Result) Ints() []int {
	var out = []int{}
	if this.s != "" {
		this.Unmarshal(&out)
		return out
	}
	if this.v != nil {
		rows, _ := this.v.([]interface{})
		for _, v := range rows {
			n, _ := v.(json.Number)
			s := string(n)
			f, _ := strconv.Atoi(s)
			out = append(out, f)
		}
	}
	return out
}

func (this Result) Int64s() []int64 {
	var out = []int64{}
	if this.s != "" {
		this.Unmarshal(&out)
		return out
	}
	if this.v != nil {
		rows, _ := this.v.([]interface{})
		for _, v := range rows {
			n, _ := v.(json.Number)
			s := string(n)
			f, _ := strconv.ParseInt(s, 10, 64)
			out = append(out, f)
		}
	}
	return out
}

func (this Result) Date() *time.Time {
	t, err := time.Parse("2006-01-02", this.s)
	if err != nil {
		return nil
	}
	return &t
}

type Meta struct {
	UserId   string
	Token    string
	IP       string
	ct       string // Content-Type
	appInfo  *sdk.AppInfo
	sign     string
	Request  *http.Request
	jsonform map[string]interface{}
	cookie   *cache.CookieValue
	st       time.Time
	ap       map[string]string // accessed URL parameters
}

func NewMeta(r *http.Request) *Meta {
	m := &Meta{
		Request: r,
		st:      time.Now(),
		ap:      make(map[string]string),
	}
	return m
}

func (this *Meta) JsonForm() map[string]interface{} {
	return this.jsonform
}

func (this *Meta) Get(key string) Result {
	s := this.Request.Form.Get(key)
	v := this.jsonform[key]

	switch s2 := v.(type) {
	case string:
		s = s2
		v = nil
	case json.Number:
		s = string(s2)
		v = nil
	}
	this.ap[key] = s
	return Result{s: s, v: v}
}

func (this *Meta) AppInfo() *sdk.AppInfo {
	if this.appInfo != nil {
		return this.appInfo
	}
	appkey := this.Request.Header.Get("appkey")
	appinfo := sdk.GetAppInfo(appkey)
	if appinfo != nil {
		this.appInfo = appinfo
	}
	return appinfo
}

func (this *Meta) ContentType() string {
	if this.ct != "" {
		return this.ct
	}
	this.ct = this.Request.Header.Get("Content-Type")
	return this.ct
}

func (this *Meta) Sign() string {
	if this.sign != "" {
		return this.sign
	}
	this.sign = this.Request.Header.Get("sign")
	return this.sign
}

// 返回马甲包ID
func (this *Meta) AppId() int {
	appinfo := this.AppInfo()
	if appinfo == nil {
		return 0
	}
	return appinfo.AppId
}

// 返回平台类型
func (this *Meta) Platform() string {
	appinfo := this.AppInfo()
	if appinfo == nil {
		return ""
	}
	return appinfo.Platform
}

func (this *Meta) Cookie() *cache.CookieValue {
	if this.cookie == nil {
		return &cache.CookieValue{}
	}
	return this.cookie
}

// 截取字符串
func trunc(s string, prelen, suflen int) string {
	n := len(s)
	if n <= prelen+suflen {
		return s
	}
	o := s[:prelen]
	o += "..."
	o += s[n-suflen:]
	return o
}

func (this *Meta) APStr() string {
	s := "{"
	for k, v := range this.ap {
		s += k + ":" + v + " "
	}
	if len(this.ap) > 0 {
		s = s[:len(s)-1]
	}
	s += "}"
	return s
}

func (this *Meta) Took() time.Duration {
	return time.Since(this.st)
}

func (this *Meta) Log() {
	aps := this.APStr()
	aps = trunc(aps, 200, 70)
	log.Println("⬅", this.Request.RequestURI, this.IP, this.UserId, aps, this.Took())
}

func (this *Meta) LogRaw() {
	aps := this.APStr()
	aps = trunc(aps, 200, 70)
	log.Println("<=", this.Request.RequestURI, this.IP, aps, this.Took())
}

func equalMimeType(t1, wish string) bool {
	/* 判断是否符合希望的文档类型, 比如equalMimeType("image/gif", "image/*|video/mp4")  */
	if t1 == wish {
		return true
	}
	r1 := strings.SplitN(t1, "/", 2)
	if len(r1) != 2 {
		return false
	}
	ts := strings.Split(wish, "|")
	for _, t2 := range ts {
		r2 := strings.SplitN(t2, "/", 2)
		if len(r2) != 2 {
			continue
		}
		if r1[0] == r2[0] && (r1[1] == r2[1] || r1[1] == "*" || r2[1] == "*") {
			return true
		}
	}

	return false
}

func (this *Meta) FileMd5(key string) (string, error) {
	f, _, err := this.Request.FormFile(key)
	if err != nil {
		log.Error(err)
		return "", NewError(400, "READFILE_FAIL")
	}
	defer f.Close()
	s := base.FileMd5String(f)
	return s, nil
}

// 请求是否加密
func (this *Meta) Encrypted() bool {
	appinfo := this.AppInfo()
	if appinfo == nil {
		return false
	}
	return appinfo.Secret != ""
}

func (this *Meta) Encrypt(plaintext string) string {
	appinfo := this.AppInfo()
	if appinfo == nil {
		log.Error("md.Encrypt: APPINFO_EMPTY")
		return ""
	}
	sign := this.Sign()
	if len(sign) < 12 {
		log.Error("md.Encrypt: SIGN_SHORT")
		return ""
	}
	out, err := aes.GcmEncryptString(plaintext, appinfo.Secret, sign[:12])
	if err != nil {
		log.Error("md.Encrypt:", err)
		return ""
	}
	return out
}

func (this *Meta) Decrypt(cipherdata []byte) []byte {
	appinfo := this.AppInfo()
	if appinfo == nil {
		log.Error("md.Decrypt: APPINFO_EMPTY")
		return nil
	}
	sign := this.Sign()
	if len(sign) < 12 {
		log.Error("md.Decrypt: SIGN_SHORT")
		return nil
	}
	out, err := aes.GcmDecrypt(cipherdata, []byte(appinfo.Secret), []byte(sign[:12]))
	if err != nil {
		log.Error("md.Decrypt:", err)
		return nil
	}
	return out
}

type Method func(*Meta) (interface{}, error)

type Error struct {
	Code        int
	Description string
	Data        interface{}
}

func (this *Error) Error() string {
	return fmt.Sprintf("%d:%s:%v", this.Code, this.Description, this.Data)
}

func NewError(code int, desc string, data ...interface{}) *Error {
	out := &Error{Code: code, Description: desc}
	if len(data) > 0 {
		out.Data = data[0]
	}
	return out
}
