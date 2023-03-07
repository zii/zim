package obs

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"path"
	"strings"

	"zim.cn/base/objstorage/common"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"zim.cn/base/uuid"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"zim.cn/base"
)

var client *obs.ObsClient

var AccessKey = "V6ED0G06OAUBLSU2C5F5"
var SecretKey = ""
var Endpoint = "obs.cn-north-4.myhuaweicloud.com"
var Bucket = "lzf2"
var BucketHost = "https://lzf2.obs.cn-north-4.myhuaweicloud.com"

func init() {
	InitClient()
}

func InitClient() {
	var err error
	client, err = obs.New(AccessKey, SecretKey, Endpoint)
	base.Raise(err)
}

type PutResult struct {
	BucketName string
	ObjectName string
	ETag       string
	Width      int
	Height     int
}

func Url(bucketName, objectName string) string {
	switch bucketName {
	case Bucket:
		return fmt.Sprintf("%s/%s", BucketHost, objectName)
	}
	return ""
}

func (this *PutResult) Url() string {
	return Url(this.BucketName, this.ObjectName)
}

func PutObject(bucketName, objectName, mime_type string, reader io.Reader) (*PutResult, error) {
	input := &obs.PutObjectInput{}
	input.Bucket = bucketName
	input.Key = objectName
	input.ContentType = mime_type
	input.Body = reader

	r, err := client.PutObject(input)
	if err != nil {
		return nil, err
	}
	etag := strings.Trim(r.ETag, `"`)
	etag = strings.ToLower(etag)
	out := &PutResult{
		BucketName: bucketName,
		ObjectName: objectName,
		ETag:       etag,
	}
	return out, nil
}

func GenerateObjectName(ext string) string {
	id := uuid.NextIDString("obs")
	prefix := rand.Int() % 256
	return fmt.Sprintf("v2-%d/%s%s", prefix, id, ext)
}

// 用网址上传
// ext: 扩展名 .jpg
func CreateObjectWithURL(buckName, ext, url string) (*PutResult, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	mime_type := rsp.Header.Get("Content-Type")
	if mime_type == "" {
		mime_type = oss.TypeByExtension(ext)
	} else if ext == "" {
		ext = common.ExtByMime(mime_type)
	}
	// Response.Body不能重复使用, 必须先读到内存, 再转成io.ReadSeeker
	var reader io.Reader = rsp.Body

	out, err := CreateObject(buckName, ext, mime_type, reader)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// fileName: 上传的原文件名
func CreateObject(bucketName, fileName, mime_type string, reader io.Reader) (*PutResult, error) {
	ext := path.Ext(fileName)
	objectName := GenerateObjectName(ext)

	result, err := PutObject(bucketName, objectName, mime_type, reader)
	if err != nil {
		return nil, err
	}
	return result, nil
}

//func main() {
//	uuid.InitUUID()
//	r, err := CreateObjectWithURL("lzf2", "", "https://www.baidu.com/img/PCtm_d9c8750bed0b3c7d089fa7d55720d6cf.png")
//	base.Raise(err)
//	fmt.Println("r:", base.JsonString(r))
//}

type CredResult struct {
	SignedUrl string
	Timeout   int
}

// 生成预签名URL
func CreateSignedUrl(timeout int) (*CredResult, error) {
	input := &obs.CreateSignedUrlInput{}
	input.Bucket = Bucket
	input.Method = obs.HttpMethodPut
	input.Expires = timeout
	r, err := client.CreateSignedUrl(input)
	if err != nil {
		return nil, err
	}
	out := &CredResult{
		SignedUrl: r.SignedUrl,
		Timeout:   timeout,
	}
	return out, nil
}
