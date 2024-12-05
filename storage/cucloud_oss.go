package storage

import (
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/casdoor/oss"
	"github.com/casdoor/oss/s3"
)

func NewCUCloudOssStorageProvider(clientId string, clientSecret string, region string, bucket string, endpoint string) oss.StorageInterface {
	sp := s3.New(&s3.Config{
		AccessID:   clientId,
		AccessKey:  clientSecret,
		Region:     region,
		Bucket:     bucket,
		Endpoint:   endpoint,
		S3Endpoint: endpoint,
		ACL:        awss3.BucketCannedACLPublicRead,
	})

	return sp
}
