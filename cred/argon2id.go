package cred

import "github.com/alexedwards/argon2id"

type Argon2idCredManager struct{}

func NewArgon2idCredManager() *Argon2idCredManager {
	cm := &Argon2idCredManager{}
	return cm
}

func (cm *Argon2idCredManager) GetHashedPassword(password string, userSalt string, organizationSalt string) string {

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return ""
	}
	return hash
}

func (cm *Argon2idCredManager) IsPasswordCorrect(plainPwd string, hashedPwd string, userSalt string, organizationSalt string) bool {
	match, _ := argon2id.ComparePasswordAndHash(plainPwd, hashedPwd)
	return match
}