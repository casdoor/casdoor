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
		Owner:         "admin",
		Name:          "built-in",
		CreatedTime:   util.GetCurrentTime(),
		DisplayName:   "Built-in Organization",
		WebsiteUrl:    "https://example.com",
		Favicon:       "https://cdn.casbin.com/static/favicon.ico",
		PhonePrefix:   "86",
		DefaultAvatar: "https://casbin.org/img/casbin.svg",
		PasswordType:  "plain",
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
		Type:          "normal-user",
		Password:      "123",
		DisplayName:   "Admin",
		Avatar:        "https://casbin.org/img/casbin.svg",
		Email:         "admin@example.com",
		Phone:         "12345678910",
		Address:       []string{},
		Affiliation:   "Example Inc.",
		Tag:           "staff",
		IsAdmin:       true,
		IsGlobalAdmin: true,
		IsForbidden:   false,
		Properties:    make(map[string]string),
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
		HomepageUrl:    "https://casdoor.org",
		Organization:   "built-in",
		EnablePassword: true,
		EnableSignUp:   true,
		Providers:      []*ProviderItem{},
		SignupItems:    []*SignupItem{},
		RedirectUris:   []string{},
		ExpireInHours:  168,
	}
	AddApplication(application)
}
