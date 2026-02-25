package util

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/publicsuffix"
)

func GetCertExpireTime(s string) (string, error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		return "", errors.New("getCertExpireTime() error, block should not be nil")
	} else if block.Type != "CERTIFICATE" {
		return "", errors.New(fmt.Sprintf("getCertExpireTime() error, block.Type should be \"CERTIFICATE\" instead of %s", block.Type))
	}

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	t := certificate.NotAfter
	return t.Local().Format(time.RFC3339), nil
}

func GetBaseDomain(domain string) (string, error) {
	// abc.com -> abc.com
	// abc.com.it -> abc.com.it
	// subdomain.abc.io -> abc.io
	// subdomain.abc.org.us -> abc.org.us
	return publicsuffix.EffectiveTLDPlusOne(domain)
}
