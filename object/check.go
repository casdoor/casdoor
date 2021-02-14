package object

import "fmt"

func CheckUserRegister(userId string, password string) string {
	if len(userId) == 0 || len(password) == 0 {
		return "username and password cannot be blank"
	} else if HasUser(userId) {
		return "username already exists"
	} else {
		return ""
	}
}

func CheckUserLogin(userId string, password string) string {
	if !HasUser(userId) {
		return "username does not exist, please sign up first"
	}

	if !IsPasswordCorrect(userId, password) {
		return "password incorrect"
	}

	return ""
}

func (user *User) getId() string {
	return fmt.Sprintf("%s/%s", user.Owner, user.Name)
}

func HasMail(application *Application, email string) string {
	user := GetMail(application.Organization, email)
	if user != nil {
		return user.getId()
	}
	return ""
}

func HasGithub(application *Application, github string) string {
	user := GetGithub(application.Organization, github)
	if user != nil {
		return user.getId()
	}
	return ""
}
