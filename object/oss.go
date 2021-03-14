package object

import (
	"bytes"
	"github.com/qor/oss"
	"github.com/qor/oss/aliyun"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	//"github.com/qor/oss/qiniu"
	"github.com/qor/oss/s3"
	"gopkg.in/ini.v1"
)

var storage oss.StorageInterface

func AliyunInit(section *ini.Section) string {
	accessID := section.Key("accessid").String()
	accessKey := section.Key("accesskey").String()
	bucket := section.Key("bucket").String()
	endpoint := section.Key("endpoint").String()
	if accessID == "" || accessKey == "" || bucket == "" || endpoint == "" {
		return "Config oss.conf wrong"
	}
	storage = aliyun.New(&aliyun.Config{
		AccessID: accessID,
		AccessKey: accessKey,
		Bucket: bucket,
		Endpoint: endpoint,
	})
	return ""
}

//func QiniuInit(section *ini.Section) string {
//	accessID := section.Key("accessid").String()
//	accessKey := section.Key("accesskey").String()
//	bucket := section.Key("bucket").String()
//	region := section.Key("region").String()
//	endpoint := section.Key("endpoint").String()
//	if accessID == "" || accessKey == "" || bucket == "" || endpoint == "" || region == "" {
//		return "Config oss.conf wrong"
//	}
//	storage = qiniu.New(&qiniu.Config{
//		AccessID: accessID,
//		AccessKey: accessKey,
//		Bucket: bucket,
//		Region: region,
//		Endpoint: endpoint,
//	})
//	return ""
//}

func Awss3Init(section *ini.Section) string {
	accessID := section.Key("accessid").String()
	accessKey := section.Key("accesskey").String()
	bucket := section.Key("bucket").String()
	region := section.Key("region").String()
	endpoint := section.Key("endpoint").String()
	if accessID == "" || accessKey == "" || bucket == "" || endpoint == "" || region == "" {
		return "Config oss.conf wrong"
	}
	storage = s3.New(&s3.Config{
		AccessID: accessID,
		AccessKey: accessKey,
		Region: region,
		Bucket: bucket,
		Endpoint: endpoint,
		ACL: awss3.BucketCannedACLPublicRead,
	})
	return ""
}

func InitOssClient() {
	if storage != nil {
		return
	}
	ossConf, err := ini.Load("./conf/oss.conf")
	if err != nil {
		panic(err)
		return
	}
	aliyunSection, _ := ossConf.GetSection("aliyun")
	qiniuSection, _ := ossConf.GetSection("qiniu")
	awss3Section, _ := ossConf.GetSection("s3")
	if aliyunSection != nil {
		AliyunInit(aliyunSection)
	} else if qiniuSection != nil {
		//QiniuInit(qiniuSection)
	} else {
		Awss3Init(awss3Section)
	}
}

func UploadAvatar(username string, avatar []byte) string {
	if storage == nil {
		InitOssClient()
		if storage == nil {
			return "oss error"
		}
	}
	_, err := storage.Put("/casdoor/avatar/" + username + ".png", bytes.NewReader(avatar))
	if err != nil {
		panic(err)
		return "oss error"
	}
	return ""
}

func GetAvatarPath() string {
	return "https://" + storage.GetEndpoint() + "/casdoor/avatar/"
}
