package object

import (
	"time"

	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"
)

func CheckPasswordExpired(user *User) bool {
	organization, err := GetOrganizationByUser(user)
	if err != nil {
		return false
	}
	passwordExpireDays := organization.PasswordExpireDays
	if passwordExpireDays <= 0 {
		return false
	}
	lastChangeTime, err := getLastTimeOfAction(organization.Name, user.Name, "set-password")
	if err != nil || lastChangeTime.IsZero() {
		return false
	}
	expiryDate := lastChangeTime.AddDate(0, 0, passwordExpireDays)
	return time.Now().After(expiryDate)
}

func getLastTimeOfAction(organization, username, action string) (time.Time, error) {
	record := &casvisorsdk.Record{
		Organization: organization,
		User:         username,
		Action:       action,
	}
	records, err := GetRecordsByField(record)
	if err != nil || len(records) == 0 {
		return time.Time{}, err
	}

	lastRecord := records[len(records)-1]
	lastChangeTime, err := time.Parse(time.RFC3339, lastRecord.CreatedTime)
	if err != nil {
		return time.Time{}, err
	}
	return lastChangeTime, nil
}
