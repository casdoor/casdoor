// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"bytes"
	"fmt"

	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/qor/oss"
	"github.com/qor/oss/aliyun"
	//"github.com/qor/oss/qiniu"
	"github.com/qor/oss/s3"
	"gopkg.in/ini.v1"
)

var storage oss.StorageInterface
var domain string

func AliyunInit(section *ini.Section) string {
	endpoint := section.Key("endpoint").String()
	accessId := section.Key("accessId").String()
	accessKey := section.Key("accessKey").String()
	domain = section.Key("domain").String()
	bucket := section.Key("bucket").String()
	if accessId == "" || accessKey == "" || bucket == "" || endpoint == "" {
		return "Config oss.conf wrong"
	}
	storage = aliyun.New(&aliyun.Config{
		AccessID:  accessId,
		AccessKey: accessKey,
		Bucket:    bucket,
		Endpoint:  endpoint,
	})
	return ""
}

//func QiniuInit(section *ini.Section) string {
//	endpoint := section.Key("endpoint").String()
//	accessId := section.Key("accessId").String()
//	accessKey := section.Key("accessKey").String()
//  domain = section.Key("domain").String()
//	bucket := section.Key("bucket").String()
//	region := section.Key("region").String()
//	if accessId == "" || accessKey == "" || bucket == "" || endpoint == "" || region == "" {
//		return "Config oss.conf wrong"
//	}
//	storage = qiniu.New(&qiniu.Config{
//		AccessID: accessId,
//		AccessKey: accessKey,
//		Bucket: bucket,
//		Region: region,
//		Endpoint: endpoint,
//	})
//	return ""
//}

func Awss3Init(section *ini.Section) string {
	endpoint := section.Key("endpoint").String()
	accessId := section.Key("accessId").String()
	accessKey := section.Key("accessKey").String()
	domain = section.Key("domain").String()
	bucket := section.Key("bucket").String()
	region := section.Key("region").String()
	if accessId == "" || accessKey == "" || bucket == "" || endpoint == "" || region == "" {
		return "Config oss.conf wrong"
	}
	storage = s3.New(&s3.Config{
		AccessID:  accessId,
		AccessKey: accessKey,
		Region:    region,
		Bucket:    bucket,
		Endpoint:  endpoint,
		ACL:       awss3.BucketCannedACLPublicRead,
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

	path := fmt.Sprintf("/casdoor/avatar/%s.png", username)
	_, err := storage.Put(path, bytes.NewReader(avatar))
	if err != nil {
		panic(err)
	}
	return ""
}

func GetAvatarPath() string {
	return fmt.Sprintf("https://%s/casdoor/avatar/", domain)
}
