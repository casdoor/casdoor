package object

import (
	"fmt"
	"testing"

	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

func updateUserColumn(column string, user *User) bool {
	affected, err := adapter.Engine.ID(core.PK{user.Owner, user.Name}).Cols(column).Update(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func TestSyncAvatarsFromGitHub(t *testing.T) {
	InitConfig()

	users := GetGlobalUsers()
	for _, user := range users {
		if user.Github == "" {
			continue
		}

		user.Avatar = fmt.Sprintf("https://avatars.githubusercontent.com/%s", user.Github)
		updateUserColumn("avatar", user)
	}
}

func TestSyncIds(t *testing.T) {
	InitConfig()

	users := GetGlobalUsers()
	for _, user := range users {
		if user.Id != "" {
			continue
		}

		user.Id = util.GenerateId()
		updateUserColumn("id", user)
	}
}

func TestGetSaltedPassword(t *testing.T) {
	password := "123456"
	fmt.Printf("%s -> %s\n", password, getSaltedPassword(password))
}
