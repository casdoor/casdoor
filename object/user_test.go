package object

import (
	"fmt"
	"testing"

	"xorm.io/core"
)

func updateUserAvatar(user *User) bool {
	affected, err := adapter.engine.ID(core.PK{user.Owner, user.Name}).Cols("avatar").Update(user)
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
		updateUserAvatar(user)
	}
}
