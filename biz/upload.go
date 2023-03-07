package biz

import (
	"errors"

	"zim.cn/base/objstorage/minio"

	"zim.cn/base/objstorage/cos"

	"zim.cn/base/objstorage/obs"

	"zim.cn/base/objstorage/oss"
	"zim.cn/biz/proto"
)

// 获取直连凭证
func GetTLCredential(platform string, user_id string) (*proto.Credential, error) {
	out := &proto.Credential{
		Platform: platform,
	}
	switch platform {
	case "oss":
		r, err := oss.Credential(user_id, 3600)
		if err != nil {
			return nil, err
		}
		c := &proto.OssCred{
			Endpoint:        r.Endpoint,
			AccessKeyId:     r.AccessKeyId,
			AccessKeySecret: r.AccessKeySecret,
			Token:           r.Token,
			Timeout:         r.Timeout,
			Bucket:          r.Bucket,
			FinalHost:       r.FinalHost,
		}
		out.Oss = c
		return out, nil
	case "obs":
		r, err := obs.CreateSignedUrl(3600)
		if err != nil {
			return nil, err
		}
		c := &proto.ObsCred{
			SignedUrl: r.SignedUrl,
			Timeout:   r.Timeout,
		}
		out.Obs = c
		return out, nil
	case "cos":
		r, err := cos.Credential(3600)
		if err != nil {
			return nil, err
		}
		c := &proto.CosCred{
			Bucket:       r.Bucket,
			Region:       r.Region,
			TmpSecretId:  r.TmpSecretId,
			TmpSecretKey: r.TmpSecretKey,
			SessionToken: r.SessionToken,
			Timeout:      r.Timeout,
		}
		out.Cos = c
		return out, nil
	case "minio":
		r, err := minio.Credential(3600)
		if err != nil {
			return nil, err
		}
		c := &proto.MinioCred{
			Endpoint:  r.Endpoint,
			Bucket:    r.Bucket,
			FinalHost: r.FinalHost,
			AccessId:  r.AccessId,
			AccessKey: r.AccessKey,
			Token:     r.Token,
			Timeout:   r.Timeout,
		}
		out.Minio = c
		return out, nil
	}
	return nil, errors.New("INVALID_PLATFORM")
}
