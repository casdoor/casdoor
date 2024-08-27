package storage

import (
	"github.com/casdoor/oss"
	"github.com/casdoor/oss/casdoor"
)

func NewCasdoorStorageProvider(providerType string, clientId string, clientSecret string, region string, bucket string, endpoint string, cert string, content string) oss.StorageInterface {
	sp := casdoor.New(&casdoor.Config{
		clientId,
		clientSecret,
		endpoint,
		cert,
		region,
		content,
		bucket,
	})
	return sp
}
