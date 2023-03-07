package cos

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"

	"golang.org/x/net/context"
	"zim.cn/base/objstorage/common"
	"zim.cn/base/uuid"
)

var AppId = "1254751699"
var SecretID = "AKID62j9xc3EBtMZWEjyxgAPO7cssd8PPAgy"
var SecretKey = ""
var Region = "ap-nanjing"
var Bucket = "cat1-1254751699"
var BucketHost = ""

var client *cos.Client
var stsClient *sts.Client

func init() {
	InitClient()
}

func InitClient() {
	// 将 examplebucket-1250000000 和 COS_REGION 修改为用户真实的信息
	// 存储桶名称，由bucketname-appid 组成，appid必须填入，可以在COS控制台查看存储桶名称。https://console.cloud.tencent.com/cos5/bucket
	// COS_REGION 可以在控制台查看，https://console.cloud.tencent.com/cos5/bucket, 关于地域的详情见 https://cloud.tencent.com/document/product/436/6224
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", Bucket, Region))
	// 用于Get Service 查询，默认全地域 service.cos.myqcloud.com
	su, _ := url.Parse(fmt.Sprintf("https://cos.%s.myqcloud.com", Region))
	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}
	client = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  SecretID,  // 替换为用户的 SecretId，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
			SecretKey: SecretKey, // 替换为用户的 SecretKey，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
		},
	})
	InitSts()
}

func InitSts() {
	cli := sts.NewClient(
		SecretID,
		SecretKey,
		nil,
	)
	stsClient = cli
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
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", bucketName, Region, objectName)
}

func PutObject(bucketName, objectName, mime_type string, reader io.Reader) (*PutResult, error) {
	opt := &cos.ObjectPutOptions{
		nil,
		&cos.ObjectPutHeaderOptions{
			ContentType: mime_type,
		},
	}
	r, err := client.Object.Put(context.Background(), objectName, reader, opt)
	if err != nil {
		return nil, err
	}
	etag := r.Header.Get("ETag")
	etag = strings.Trim(etag, `"`)
	etag = strings.ToLower(etag)
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
//	r, err := CreateObjectWithURL(Bucket, "", "https://www.baidu.com/img/PCtm_d9c8750bed0b3c7d089fa7d55720d6cf.png")
//	base.Raise(err)
//	fmt.Println("r:", base.JsonString(r))
//	fmt.Println("url:", r.Url())
//}

type CredResult struct {
	Bucket       string
	Region       string
	TmpSecretId  string
	TmpSecretKey string
	SessionToken string
	Timeout      int
}

// 获取直传凭证
func Credential(timeout int) (*CredResult, error) {
	opt := &sts.CredentialOptions{
		DurationSeconds: int64(timeout),
		Region:          Region,
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{
				{
					Action: []string{
						"name/cos:PostObject",
						"name/cos:PutObject",
					},
					Effect: "allow",
					Resource: []string{
						"qcs::cos:" + Region + ":uid/" + AppId + ":" + Bucket + "/*",
					},
				},
			},
		},
	}
	res, err := stsClient.GetCredential(opt)
	if err != nil {
		return nil, err
	}
	out := &CredResult{
		Bucket:       Bucket,
		Region:       Region,
		TmpSecretId:  res.Credentials.TmpSecretID,
		TmpSecretKey: res.Credentials.TmpSecretKey,
		SessionToken: res.Credentials.SessionToken,
		Timeout:      timeout,
	}
	return out, nil
}
