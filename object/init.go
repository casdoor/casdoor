package object

import "github.com/casdoor/casdoor/util"

func InitDb() {
	initBuiltInOrganization()
	initBuiltInUser()
}

func initBuiltInOrganization() {
	organization := getOrganization("admin", "built-in")
	if organization != nil {
		return
	}

	organization = &Organization{
		Owner:       "admin",
		Name:        "built-in",
		CreatedTime: util.GetCurrentTime(),
		DisplayName: "Built-in Organization",
		WebsiteUrl:  "https://example.com",
	}
	AddOrganization(organization)
}

func initBuiltInUser() {
	user := getUser("built-in", "admin")
	if user != nil {
		return
	}

	user = &User{
		Owner:        "built-in",
		Name:         "admin",
		CreatedTime:  util.GetCurrentTime(),
		Id:           util.GenerateId(),
		Password:     "123",
		PasswordType: "plain",
		DisplayName:  "Admin",
		Avatar:       "https://casbin.org/img/casbin.svg",
		Email:        "admin@example.com",
		Phone:        "1-12345678",
		Affiliation:  "Example Inc.",
		Tag:          "staff",
		IsAdmin:      true,
		Github:       "",
	}
	AddUser(user)
}
