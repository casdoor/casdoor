package object

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
