package util

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

func EncodeSHA1AndBase64(message, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(message))
	sum := h.Sum(nil)
	sig := base64.StdEncoding.EncodeToString(sum)

	return sig
}