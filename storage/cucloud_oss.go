package storage

import (
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
		ACL:        string(types.BucketCannedACLPublicRead),
	})

	return sp
}
