package cred

import "golang.org/x/crypto/bcrypt"

type BcryptCredManager struct{}

func NewBcryptCredManager() *BcryptCredManager {
	cm := &BcryptCredManager{}
	return cm
}

func (cm *BcryptCredManager) GetHashedPassword(password string, userSalt string, organizationSalt string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (cm *BcryptCredManager) IsPasswordCorrect(plainPwd string, hashedPwd string, userSalt string, organizationSalt string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
	return err == nil
}
