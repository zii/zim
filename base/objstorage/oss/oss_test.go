package oss

import (
	"fmt"
	"testing"

	"zim.cn/base"
	"zim.cn/base/uuid"
)

func Test1(t *testing.T) {
	uuid.InitUUID()
	//bucket, err := client.Bucket("winkchat-pic1")
	//base.Raise(err)
	//fmt.Println("bucket:", bucket)
	//newobject := base64.URLEncoding.EncodeToString([]byte("2019/03/30/q2u2dyezwc_w50.jpg"))
	//process := fmt.Sprintf("image/resize,w_%d|sys/saveas,o_%s", 50, newobject)
	//r, err := bucket.ProcessObject("2019/03/30/q2u2dyezwc.jpg", process)
	//base.Raise(err)
	//fmt.Println("process:", r)

	//for i := 0; i < 5; i++ {
	//	r, err := Snapshot(i*1000, "winkchat-vid1", "v2-67/6u52r8y5ybv.mp4", "winkchat-pic1")
	//	base.Raise(err)
	//	fmt.Println("r:", r, err)
	//}

	//bucket, err := client.Bucket("winkchat-vid1")
	//base.Raise(err)
	//st := time.Now()
	////err = bucket.PutObjectFromFile("test1.mp4", "/Users/cat/Downloads/SampleVideo_1280x720_10mb.mp4")
	////base.Raise(err)
	//_, err = bucket.ProcessObject("v2-189/acxk9dnrfa4.mp4",
	//	"video/snapshot,t_0,f_jpg|sys/saveas,o_djItNzIvYWN4azlmdm82Y2MuanBn,b_d2lua2NoYXQtcGljMQ==")
	//base.Raise(err)
	//fmt.Println("took:", time.Since(st))

	//w, h := GetImageDimUrl("https://p1.winkchat.innonice.com/v2-176/dlrv61nuh0y.jpg")
	//fmt.Println("wh:", w, h)

	r, err := CreateObjectWithURL(Bucket, "",
		"http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTLz3icr3mGs5ib1cCskViccqepj7oJcSlJpFSBmwibheNbCsrlVSZLfVanK0GaQibNDMQIHWoxSmDibD80g/132")
	base.Raise(err)
	fmt.Println("r:", r)
}

func Test2(t *testing.T) {
	r, err := Credential("u1", 3600)
	base.Raise(err)
	fmt.Println(base.JsonPretty(r))
}
