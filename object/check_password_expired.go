package object

import (
	"time"
)

func checkPasswordExpired(user *User) bool {
	organization, err := GetOrganizationByUser(user)
	if err != nil {
		return false
	}
	passwordExpireDays := organization.PasswordExpireDays
	if passwordExpireDays <= 0 {
		return false
	}
	lastChangePasswordTime := user.LastChangePasswordTime
	if lastChangePasswordTime == "" {
		return false
	}
	lastTime, err := time.Parse(time.RFC3339, lastChangePasswordTime)
	if err != nil {
		return false
	}
	expiryDate := lastTime.AddDate(0, 0, passwordExpireDays)
	return time.Now().After(expiryDate)
}
