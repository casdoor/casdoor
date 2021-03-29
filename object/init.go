package object

import "github.com/casdoor/casdoor/util"

func InitDb() {
	initBuiltInOrganization()
	initBuiltInUser()
	initBuiltInApplication()
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
		Owner:         "built-in",
		Name:          "admin",
		CreatedTime:   util.GetCurrentTime(),
		Id:            util.GenerateId(),
		Password:      "123",
		PasswordType:  "plain",
		DisplayName:   "Admin",
		Avatar:        "https://casbin.org/img/casbin.svg",
		Email:         "admin@example.com",
		Phone:         "1-12345678",
		Affiliation:   "Example Inc.",
		Tag:           "staff",
		IsAdmin:       true,
		IsGlobalAdmin: true,
	}
	AddUser(user)
}

func initBuiltInApplication() {
	application := getApplication("admin", "app-built-in")
	if application != nil {
		return
	}

	application = &Application{
		Owner:          "admin",
		Name:           "app-built-in",
		CreatedTime:    util.GetCurrentTime(),
		DisplayName:    "Casdoor",
		Logo:           "https://cdn.casbin.com/logo/logo_1024x256.png",
		Organization:   "built-in",
		EnablePassword: true,
		Providers:      []string{},
	}
	AddApplication(application)
}
