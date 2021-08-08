package storage

import (
	"github.com/casbin/casdoor/util/filesystem"
	"github.com/qor/oss"
)

func NewLocalStorageProvider() oss.StorageInterface {
	return filesystem.New("storage")
}