package oss

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"

	"zim.cn/base/objstorage/common"

	"zim.cn/base"
	"zim.cn/base/uuid"

	"github.com/disintegration/imaging"

	_ "image/jpeg"
	_ "image/png"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var client *oss.Client
var stsClient *sts.Client

var AccessKeyID = "LTAI5t8bJGHXeVAXanzGz5QF"
var AccessKeySecret = ""
var EndPoint = "oss-cn-guangzhou.aliyuncs.com"
var Bucket = "openim1"
var BucketHost = "http://openim1.oss-accelerate.aliyuncs.com"
var RoleArn = "acs:ram::1148165249088458:role/ramosstest"

func init() {
	InitClient()
}

func InitClient() {
	var err error
	client, err = oss.New(EndPoint, AccessKeyID, AccessKeySecret)
	base.Raise(err)
	err = InitSts()
	base.Raise(err)
}

func Url(bucketName, objectName string) string {
	switch bucketName {
	case Bucket:
		return fmt.Sprintf("%s/%s", BucketHost, objectName)
	}
	return ""
}

type PutResult struct {
	BucketName string
	ObjectName string
	ETag       string
	Width      int
	Height     int
}

func (this *PutResult) Url() string {
	return Url(this.BucketName, this.ObjectName)
}

func PutObject(bucketName, objectName, mime_type string, reader io.Reader) (*PutResult, error) {
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	var opts []oss.Option
	if mime_type != "" {
		opts = append(opts, oss.ContentType(mime_type))
	}
	request := &oss.PutObjectRequest{
		ObjectKey: objectName,
		Reader:    reader,
	}
	resp, err := bucket.DoPutObject(request, opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	etag := resp.Headers.Get("ETag")
	etag = strings.Trim(etag, `"`)
	etag = strings.ToLower(etag)

	out := &PutResult{
		BucketName: bucketName,
		ObjectName: objectName,
		ETag:       etag,
	}
	return out, nil
}

func RemoveObjects(bucketName string, objectNames []string) error {
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return err
	}

	_, err = bucket.DeleteObjects(objectNames)
	return err
}

func GenerateObjectName(ext string) string {
	id := uuid.NextIDString("oss")
	prefix := rand.Int() % 256
	return fmt.Sprintf("v2-%d/%s%s", prefix, id, ext)
}

// fileName: ?????????????????????
func CreateObject(bucketName, fileName, mime_type string, reader io.Reader) (*PutResult, error) {
	ext := path.Ext(fileName)
	objectName := GenerateObjectName(ext)

	result, err := PutObject(bucketName, objectName, mime_type, reader)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ???????????????
// ext: ????????? .jpg
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
	// Response.Body??????????????????, ?????????????????????, ?????????io.ReadSeeker
	var reader io.Reader = rsp.Body

	out, err := CreateObject(buckName, ext, mime_type, reader)
	if err != nil {
		return nil, err
	}
	return out, nil
}

const (
	ThumbnailSize = 240
)

// ???????????????url
// ext2: ?????????????????????(.png)
func ConvertSmallUrl(url, ext2 string) string {
	if url == "" {
		return ""
	}
	ext := path.Ext(url)
	prefix := url[:len(url)-len(ext)]
	if ext2 != "" {
		ext = ext2
	}
	return fmt.Sprintf("%s_s%s", prefix, ext)
}

// ????????????????????????url
// ext2: ?????????????????????(.gif)
func ConvertPreviewUrl(url, ext2 string) string {
	if url == "" {
		return ""
	}
	ext := path.Ext(url)
	prefix := url[:len(url)-len(ext)]
	if ext2 != "" {
		ext = ext2
	}
	return fmt.Sprintf("%s_a%s", prefix, ext)
}

// ?????????????????????, ?????????, ???????????????
func Thumbnail(buckName, objectName string) (*PutResult, error) {
	bucket, err := client.Bucket(buckName)
	if err != nil {
		return nil, err
	}
	newObjectName := ConvertSmallUrl(objectName, ".png")
	newobject := base64.URLEncoding.EncodeToString([]byte(newObjectName))
	process := fmt.Sprintf("image/resize,w_%d,h_%d/format,png|sys/saveas,o_%s", ThumbnailSize, ThumbnailSize, newobject)
	_, err = bucket.ProcessObject(objectName, process)
	if err != nil {
		return nil, err
	}

	out := &PutResult{
		BucketName: buckName,
		ObjectName: newObjectName,
		ETag:       "",
	}
	return out, nil
}

// ????????????, ?????????
// t: ????????????(ms)
// buckName: ??????buck name
// objectName: ??????OSS?????????
func Snapshot(t int, buckName, objectName string, photoBuckName string) (*PutResult, error) {
	bucket, err := client.Bucket(buckName)
	if err != nil {
		return nil, err
	}
	newObjectName := GenerateObjectName(".jpg")
	newobject := base64.URLEncoding.EncodeToString([]byte(newObjectName))
	newbuck := base64.URLEncoding.EncodeToString([]byte(photoBuckName))
	process := fmt.Sprintf("video/snapshot,t_%d,f_jpg|sys/saveas,o_%s,b_%s", t, newobject, newbuck)
	_, err = bucket.ProcessObject(objectName, process)
	if err != nil {
		return nil, err
	}

	// ??????etag
	photoBucket, err := client.Bucket(photoBuckName)
	if err != nil {
		return nil, err
	}
	meta, err := photoBucket.GetObjectMeta(newObjectName)
	if err != nil {
		log.Println("oss panic:", err)
		return nil, err
	}
	etag := meta.Get("ETag")
	etag = strings.Trim(etag, `"`)
	etag = strings.ToLower(etag)

	out := &PutResult{
		BucketName: photoBuckName,
		ObjectName: newObjectName,
		ETag:       etag,
	}
	return out, nil
}

// ??????????????????
func GetImageDim(r io.Reader) (int, int) {
	c, _, err := image.DecodeConfig(r)
	if err != nil {
		log.Println("GetImageDim:", err)
	}
	return c.Width, c.Height
}

func GetImageDimUrl(url string) (w int, h int) {
	rsp, err := http.Get(url)
	if err != nil {
		return
	}
	defer rsp.Body.Close()

	w, h = GetImageDim(rsp.Body)
	return
}

// ???golang??????png?????????
func ThumbnailImage(f io.Reader) (image.Image, error) {
	src, err := imaging.Decode(f, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	im := imaging.Thumbnail(src, ThumbnailSize, ThumbnailSize, imaging.Lanczos)
	return im, nil
}

// ???????????????(png)
func CreatePNGThumbnail(bucketName, srcObjectName string, f io.Reader) (*PutResult, error) {
	im, err := ThumbnailImage(f)
	if err != nil {
		log.Println("CreateThumbnail err1:", err)
		return nil, err
	}
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		err := imaging.Encode(pw, im, imaging.PNG, imaging.PNGCompressionLevel(png.BestCompression))
		if err != nil {
			log.Println("CreateThumbnail err2:", err)
		}
	}()
	objectName := ConvertSmallUrl(srcObjectName, ".png")
	result, err := PutObject(bucketName, objectName, "image/png", pr)
	if err != nil {
		log.Println("CreateThumbnail err3:", err)
	}
	return result, err
}

// ??????????????????????????????
type CVOption struct {
	TimeOff  string // ????????????:????????????(???), ??????0
	Duration string // ????????????:??????(???), ??????1
	Rate     string // ??????, ??????5
	Scale    string // ??????, ??????????????????480
}

// ??????????????????(????????????ffmpeg)
// in: ?????????
// format: ???????????? mp4/gif/webp/apng...
// ???????????????
func ConvertVideo(in io.Reader, format string, option *CVOption) (io.Reader, error) {
	pid := os.Getpid()
	tmpfile := fmt.Sprintf("/tmp/ff%d.tmp", pid)
	w, err := os.OpenFile(tmpfile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	defer func() {
		w.Close()
		os.Remove(tmpfile)
	}()
	_, err = io.Copy(w, in)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	var errbuf bytes.Buffer
	if option == nil {
		option = &CVOption{
			TimeOff:  "0",
			Duration: "1",
			Rate:     "10",
			Scale:    "-2:480", // -2??????????????????????????????
		}
	}
	cmd := exec.Command("ffmpeg", "-i", tmpfile, "-an", "-ss", option.TimeOff, "-t", option.Duration,
		"-r", option.Rate, "-vf", "scale="+option.Scale, "-movflags", "frag_keyframe", "-f", format, "pipe:1")
	cmd.Stderr = &errbuf
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Println("err:", errbuf.String())
		return nil, err
	}
	return &out, nil
}

func ConvertVideoURL(url string, format string, option *CVOption) (io.Reader, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ConvertVideo(rsp.Body, format, option)
}

type CredentialResult struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	Token           string
	Timeout         int
	Bucket          string
	FinalHost       string
}

func InitSts() error {
	const regionID = "cn-fujian" // ????????????, ?????????????????????sts.aliyuncs.com
	client, err := sts.NewClientWithAccessKey(regionID, AccessKeyID, AccessKeySecret)
	if err != nil {
		return err
	}
	stsClient = client
	return nil
}

// ???????????????????????????????????????
// timeout: ???
func Credential(user_id string, timeout int) (*CredentialResult, error) {
	if stsClient == nil {
		err := InitSts()
		if err != nil {
			return nil, err
		}
	}

	request := sts.CreateAssumeRoleRequest()
	request.Scheme = "https"

	//????????????
	request.RoleArn = RoleArn
	request.RoleSessionName = fmt.Sprintf("%s-%d", user_id, time.Now().Unix())
	request.DurationSeconds = requests.NewInteger(timeout)

	rsp, err := stsClient.AssumeRole(request)
	if err != nil {
		return nil, err
	}
	out := &CredentialResult{
		Endpoint:        EndPoint,
		AccessKeyId:     rsp.Credentials.AccessKeyId,
		AccessKeySecret: rsp.Credentials.AccessKeySecret,
		Token:           rsp.Credentials.SecurityToken,
		Timeout:         timeout,
		Bucket:          Bucket,
		FinalHost:       BucketHost,
	}
	return out, nil
}
