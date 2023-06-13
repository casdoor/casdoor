package object

import (
	"testing"
)

func TestSaltedPasswordsGeneration(t *testing.T) {
	password := "casdoor"
	user1 := User{Password: password}
	user2 := User{Password: password}
	tests := []struct {
		name string
	}{
		{
			name: "md5-salt",
		},
		{
			name: "salt",
		},
		{
			name: "pbkdf2-salt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			organization := Organization{PasswordType: tt.name}
			user1.UpdateUserPassword(&organization)
			user2.UpdateUserPassword(&organization)
			if user1.Password == user2.Password {
				t.Error("Password hashes should be different but they are the same")
			}
			if user1.Password == password || user2.Password == password {
				t.Error("Password should be hashed but it wasn't")
			}
		})
	}
}
