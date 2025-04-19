package cred

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// password type: pbkdf2-django

type Pbkdf2DjangoCredManager struct{}

func NewPbkdf2DjangoCredManager() *Pbkdf2DjangoCredManager {
	cm := &Pbkdf2DjangoCredManager{}
	return cm
}

func (m *Pbkdf2DjangoCredManager) GetHashedPassword(password string, userSalt string, organizationSalt string) string {
	iterations := 260000
	salt := userSalt
	if salt == "" {
		salt = organizationSalt
	}

	saltBytes := []byte(salt)
	passwordBytes := []byte(password)
	computedHash := pbkdf2.Key(passwordBytes, saltBytes, iterations, sha256.Size, sha256.New)
	hashBase64 := base64.StdEncoding.EncodeToString(computedHash)
	return "pbkdf2_sha256$" + strconv.Itoa(iterations) + "$" + salt + "$" + hashBase64
}

func (m *Pbkdf2DjangoCredManager) IsPasswordCorrect(password string, passwordHash string, userSalt string, organizationSalt string) bool {
	parts := strings.Split(passwordHash, "$")
	if len(parts) != 4 {
		return false
	}

	algorithm, iterations, salt, hash := parts[0], parts[1], parts[2], parts[3]
	if algorithm != "pbkdf2_sha256" {
		return false
	}

	iter, err := strconv.Atoi(iterations)
	if err != nil {
		return false
	}

	saltBytes := []byte(salt)
	passwordBytes := []byte(password)
	computedHash := pbkdf2.Key(passwordBytes, saltBytes, iter, sha256.Size, sha256.New)
	computedHashBase64 := base64.StdEncoding.EncodeToString(computedHash)

	return computedHashBase64 == hash
}
