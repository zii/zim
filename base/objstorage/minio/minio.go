package minio

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	url2 "net/url"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"zim.cn/base"

	"zim.cn/base/objstorage/common"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"zim.cn/base/uuid"

	"golang.org/x/net/context"
)

var AccessKeyID = "user12345"
var SecretAccessKey = "key12345"
var Region = "us-east-1"
var Endpoint = "http://10.10.10.86:9000"
var Bucket = "buck1"
var BucketHost = ""

var client *minio.Client

func init() {
	InitClient()
}

func InitClient() {
	opts := &minio.Options{
		Creds: credentials.NewStaticV4(AccessKeyID, SecretAccessKey, ""),
	}
	opts.Secure = false
	//opts.Region = Region
	minioUrl, err := url2.Parse(Endpoint)
	base.Raise(err)
	c, err := minio.New(minioUrl.Host, opts)
	base.Raise(err)
	client = c
}

type PutResult struct {
	BucketName string
	ObjectName string
	ETag       string
	Width      int
	Height     int
}

func Url(bucketName, objectName string) string {
	if BucketHost != "" {
		return fmt.Sprintf("%s/%s", BucketHost, objectName)
	}
	return fmt.Sprintf("%s/%s/%s", Endpoint, Bucket, objectName)
}

func PutObject(bucketName, objectName, mime_type string, reader io.Reader, size int64) (*PutResult, error) {
	opt := minio.PutObjectOptions{
		ContentType:          mime_type,
		SendContentMd5:       true,
		DisableContentSha256: true,
	}
	r, err := client.PutObject(context.Background(), bucketName, objectName, reader, size, opt)
	if err != nil {
		return nil, err
	}
	etag := r.ETag
	out := &PutResult{
		BucketName: bucketName,
		ObjectName: objectName,
		ETag:       etag,
	}
	return out, nil
}

func (this *PutResult) Url() string {
	return Url(this.BucketName, this.ObjectName)
}

func GenerateObjectName(ext string) string {
	id := uuid.NextIDString("cos")
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

	out, err := CreateObject(buckName, ext, mime_type, reader, rsp.ContentLength)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// fileName: 上传的原文件名
func CreateObject(bucketName, fileName, mime_type string, reader io.Reader, size int64) (*PutResult, error) {
	ext := path.Ext(fileName)
	objectName := GenerateObjectName(ext)

	result, err := PutObject(bucketName, objectName, mime_type, reader, size)
	if err != nil {
		return nil, err
	}
	return result, nil
}

//func main() {
//	uuid.InitUUID()
//	r, err := CreateObjectWithURL(Bucket, "", "https://www.baidu.com/img/PCtm_d9c8750bed0b3c7d089fa7d55720d6cf.png")
//	base.Raise(err)
//	fmt.Println("r:", base.JsonString(r))
//	fmt.Println("url:", r.Url())
//}

type CredResult struct {
	Endpoint  string
	Bucket    string
	FinalHost string
	AccessId  string
	AccessKey string
	Token     string
	Timeout   int
}

// 生成直传凭证
func Credential(timeout int) (*CredResult, error) {
	var stsOpts credentials.STSAssumeRoleOptions
	stsOpts.AccessKey = AccessKeyID
	stsOpts.SecretKey = SecretAccessKey
	stsOpts.DurationSeconds = timeout
	stsOpts.Location = Region
	li, err := credentials.NewSTSAssumeRole(Endpoint, stsOpts)
	if err != nil {
		return nil, err
	}
	v, err := li.Get()
	if err != nil {
		return nil, err
	}

	out := &CredResult{
		Endpoint:  Endpoint,
		Bucket:    Bucket,
		AccessId:  v.AccessKeyID,
		AccessKey: v.SecretAccessKey,
		Token:     v.SessionToken,
		Timeout:   timeout,
	}
	if BucketHost != "" {
		out.FinalHost = BucketHost
	} else {
		out.FinalHost = fmt.Sprintf("%s/%s", Endpoint, Bucket)
	}
	return out, nil
}
