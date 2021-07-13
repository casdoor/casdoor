package util

import (
	"bytes"
	"github.com/pquerna/otp/totp"
	"image/png"
)

func GetTOTPLink(accountName string) ([]byte, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Example.com",
		AccountName: "alice@example.com",
	})
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	img, _ := key.Image(200, 200)
	_ = png.Encode(&buf, img)

	return buf.Bytes(), key.Secret(), nil
}

func ValidateTOTP(passcode, secret string) bool {
	return totp.Validate(passcode, secret)
}
