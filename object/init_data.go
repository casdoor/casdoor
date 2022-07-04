package object

import (
	"github.com/casdoor/casdoor/util"
	"os"
)

type InitData struct {
	Organizations []*Organization `json:"organizations"`
	Applications  []*Application  `json:"applications"`
	Users         []*User         `json:"users"`
	Certs         []*Cert         `json:"certs"`
	Providers     []*Provider     `json:"providers"`
	Ldaps         []*Ldap         `json:"ldaps"`
}

func InitFromFile() {
	initData := readInitDataFromFile("./init_data.json")
	if initData != nil {
		for _, organization := range initData.Organizations {
			initDefinedOrganization(organization)
		}
		for _, provider := range initData.Providers {
			initDefinedProvider(provider)
		}
		for _, user := range initData.Users {
			initDefinedUser(user)
		}
		for _, application := range initData.Applications {
			initDefinedApplication(application)
		}
		for _, cert := range initData.Certs {
			initDefinedCert(cert)
		}
		for _, ldap := range initData.Ldaps {
			initDefinedLdap(ldap)
		}
	}
}

func readInitDataFromFile(filePath string) *InitData {
	_, err := os.Stat(filePath)
	if err != nil {
		return nil
	}

	s := util.ReadStringFromPath(filePath)

	data := &InitData{}
	err = util.JsonToStruct(s, data)
	if err != nil {
		return nil
	}
	return data
}

func initDefinedOrganization(organization *Organization) {
	existed := getOrganization(organization.Owner, organization.Name)
	if existed != nil {
		return
	}
	organization.CreatedTime = util.GetCurrentTime()
	organization.AccountItems = []*AccountItem{
		{Name: "Organization", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "ID", Visible: true, ViewRule: "Public", ModifyRule: "Immutable"},
		{Name: "Name", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "Display name", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Avatar", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "User type", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "Password", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
		{Name: "Email", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Phone", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Country/Region", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Location", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Affiliation", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Title", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Homepage", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Bio", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Tag", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "Signup application", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "3rd-party logins", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
		{Name: "Properties", Visible: false, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is admin", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is global admin", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is forbidden", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is deleted", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
	}

	AddOrganization(organization)
}

func initDefinedApplication(application *Application) {
	existed := getApplication(application.Owner, application.Name)
	if existed != nil {
		return
	}
	application.CreatedTime = util.GetCurrentTime()
	AddApplication(application)
}

func initDefinedUser(user *User) {
	existed := getUser(user.Owner, user.Name)
	if existed != nil {
		return
	}
	user.CreatedTime = util.GetCurrentTime()
	user.Id = util.GenerateId()
	user.Properties = make(map[string]string)
	AddUser(user)
}

func initDefinedCert(cert *Cert) {
	existed := getCert(cert.Owner, cert.Name)
	if existed != nil {
		return
	}
	cert.CreatedTime = util.GetCurrentTime()
	AddCert(cert)
}

func initDefinedLdap(ldap *Ldap) {
	existed := GetLdap(ldap.Id)
	if existed != nil {
		return
	}
	AddLdap(ldap)
}

func initDefinedProvider(provider *Provider) {
	existed := GetProvider(provider.GetId())
	if existed != nil {
		return
	}
	AddProvider(provider)
}
