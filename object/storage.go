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
	"github.com/casbin/casdoor/util"
	"github.com/qor/oss"
	"github.com/qor/oss/aliyun"
	//"github.com/qor/oss/qiniu"
	"github.com/qor/oss/s3"
)

func getAliyunClient(provider *Provider) oss.StorageInterface {
	if util.IsStrsEmpty(
		provider.Endpoint,
		provider.ClientId,
		provider.ClientSecret,
		provider.Bucket) {
		return nil
	}

	ret := aliyun.New(&aliyun.Config{
		AccessID:  provider.ClientId,
		AccessKey: provider.ClientSecret,
		Bucket:    provider.Bucket,
		Endpoint:  provider.Endpoint,
	})

	if len(provider.Domain) == 0 {
		provider.Domain = ret.GetEndpoint()
		UpdateProvider(provider.GetId(), provider)
	}
	return ret
}

func getQiniuClient(provider *Provider) oss.StorageInterface {
	fmt.Println("Casdoor does not support Qiniu now.")
	return nil
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
}

func getAwss3Client(provider *Provider) oss.StorageInterface {
	if util.IsStrsEmpty(
		provider.Endpoint,
		provider.ClientId,
		provider.ClientSecret,
		provider.Bucket,
		provider.RegionId) {
		return nil
	}

	ret := s3.New(&s3.Config{
		AccessID:  provider.ClientId,
		AccessKey: provider.ClientSecret,
		Region:    provider.RegionId,
		Bucket:    provider.Bucket,
		Endpoint:  provider.Endpoint,
		ACL:       awss3.BucketCannedACLPublicRead,
	})

	if len(provider.Domain) == 0 {
		provider.Domain = ret.GetEndpoint()
		UpdateProvider(provider.GetId(), provider)
	}
	return ret
}

func getStorageClient(provider *Provider) oss.StorageInterface {
	if provider == nil || provider.Category != "Storage" {
		return nil
	}

	switch provider.Type {
	case "Aliyun OSS":
		return getAliyunClient(provider)
	case "Qiniu":
		return getQiniuClient(provider)
	case "AWS S3":
		return getAwss3Client(provider)
	}

	return nil
}

func UploadAvatar(provider *Provider, username string, avatar []byte) string {
	storage := getStorageClient(provider)
	if storage == nil {
		return fmt.Sprintf("Provider type: %s is not supported", provider.Type)
	}

	path := fmt.Sprintf("%s/%s.png", util.UrlJoin(util.GetUrlPath(provider.Domain), "/avatar"), username)
	_, err := storage.Put(path, bytes.NewReader(avatar))
	if err != nil {
		return err.Error()
	}
	return ""
}
