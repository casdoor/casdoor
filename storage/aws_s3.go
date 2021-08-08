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

package storage

import (
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/qor/oss"
	"github.com/qor/oss/s3"
)

func NewAwsS3StorageProvider(clientId string, clientSecret string, region string, bucket string, endpoint string) oss.StorageInterface {
	sp := s3.New(&s3.Config{
		AccessID:  clientId,
		AccessKey: clientSecret,
		Region:    region,
		Bucket:    bucket,
		Endpoint:  endpoint,
		ACL:       awss3.BucketCannedACLPublicRead,
	})

	return sp
}
