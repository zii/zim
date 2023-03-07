package objstorage

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"path"

	"zim.cn/base/objstorage/minio"

	"zim.cn/base/objstorage/common"

	"zim.cn/base/uuid"

	"github.com/jinzhu/copier"
	"zim.cn/base/objstorage/cos"
	"zim.cn/base/objstorage/obs"
	oss2 "zim.cn/base/objstorage/oss"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OssConfig struct {
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`
	EndPoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
	BucketHost      string `json:"bucketHost"`
}

type ObsConfig struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	Endpoint   string `json:"endpoint"`
	Bucket     string `json:"bucket"`
	BucketHost string `json:"bucketHost"`
}

type CosConfig struct {
	SecretID   string `json:"secretID"`
	SecretKey  string `json:"secretKey"`
	Region     string `json:"region"`
	Bucket     string `json:"bucket"`
	BucketHost string `json:"bucketHost"`
}

type Config struct {
	Platform string     `json:"platform"` // 平台类型 oss/obs/...
	Oss      *OssConfig `json:"oss"`
	Obs      *ObsConfig `json:"obs"`
	Cos      *CosConfig `json:"cos"`
}

// oss/obs/cos/minio
var Platform string

type PutResult struct {
	BucketName string
	ObjectName string
	ETag       string
	Width      int
	Height     int
}

func Url(objectName string) string {
	switch Platform {
	case "oss":
		return oss2.Url(oss2.Bucket, objectName)
	case "obs":
		return obs.Url(obs.Bucket, objectName)
	case "cos":
		return cos.Url(cos.Bucket, objectName)
	case "minio":
		return minio.Url(minio.Bucket, objectName)
	}
	panic(fmt.Errorf("Objstorage: Invalid Platform: %s", Platform))
}

func (this *PutResult) Url() string {
	return Url(this.ObjectName)
}

func PutObject(objectName, mime_type string, reader io.Reader, size int64) (*PutResult, error) {
	var out = &PutResult{}
	var r any
	var err error
	switch Platform {
	case "oss":
		r, err = oss2.PutObject(oss2.Bucket, objectName, mime_type, reader)
	case "obs":
		r, err = obs.PutObject(obs.Bucket, objectName, mime_type, reader)
	case "cos":
		r, err = cos.PutObject(cos.Bucket, objectName, mime_type, reader)
	case "minio":
		r, err = minio.PutObject(minio.Bucket, objectName, mime_type, reader, size)
	default:
		panic(fmt.Errorf("Objstorage: Invalid Platform: %s", Platform))
	}
	if err != nil {
		return nil, err
	}
	copier.Copy(out, r)
	return out, nil
}

func GenerateObjectName(ext string) string {
	id := uuid.NextIDString("ostorage")
	prefix := rand.Int() % 256
	return fmt.Sprintf("v2-%d/%s%s", prefix, id, ext)
}

// 用网址上传
// ext: 扩展名 .jpg
func CreateObjectWithURL(ext, url string) (*PutResult, error) {
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
	out, err := CreateObject(ext, mime_type, reader, rsp.ContentLength)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// fileName: 上传的原文件名
func CreateObject(fileName, mime_type string, reader io.Reader, size int64) (*PutResult, error) {
	ext := path.Ext(fileName)
	objectName := GenerateObjectName(ext)

	result, err := PutObject(objectName, mime_type, reader, size)
	if err != nil {
		return nil, err
	}
	return result, nil
}
