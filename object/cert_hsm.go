// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/hsm"
	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v5"
	dsig "github.com/russellhaering/goxmldsig"
)

// CertKeyStorageHsm is the KeyStorage of a cert whose private key stays in the HSM
const CertKeyStorageHsm = "HSM"

func (p *Cert) IsHsmKey() bool {
	return p.KeyStorage == CertKeyStorageHsm
}

// getCertSigner returns the HSM signer of a cert whose private key is in the HSM
func getCertSigner(cert *Cert) (*hsm.Signer, error) {
	return getCertSignerEx(cert, false)
}

// getCertSignerEx looks the key up in the HSM again when reload is set
func getCertSignerEx(cert *Cert, reload bool) (*hsm.Signer, error) {
	if err := edition.Require(edition.Enterprise, edition.FeatureHsm); err != nil {
		return nil, err
	}
	getSigner := hsm.GetSigner
	if reload {
		getSigner = hsm.ReloadSigner
	}
	signer, err := getSigner(cert.KeyReference)
	if err != nil {
		return nil, fmt.Errorf("the private key of the cert: %s is in the HSM, %v", cert.GetId(), err)
	}
	return signer, nil
}

// checkHsmKeyAlgorithm checks that the HSM key fits the crypto algorithm of the cert
func checkHsmKeyAlgorithm(algorithm string, public crypto.PublicKey) error {
	switch key := public.(type) {
	case *rsa.PublicKey:
		if strings.HasPrefix(algorithm, "RS") || strings.HasPrefix(algorithm, "PS") {
			return nil
		}
		return fmt.Errorf("the HSM key is an RSA key, it can't be used for %s", algorithm)
	case *ecdsa.PublicKey:
		bits := key.Curve.Params().BitSize
		if (algorithm == "ES256" && bits == 256) || (algorithm == "ES384" && bits == 384) || (algorithm == "ES512" && bits == 521) {
			return nil
		}
		return fmt.Errorf("the HSM key is an EC key on %s, it can't be used for %s", key.Curve.Params().Name, algorithm)
	}
	return fmt.Errorf("unsupported HSM key type: %T", public)
}

var x509SignatureAlgorithms = map[string]x509.SignatureAlgorithm{
	"RS256": x509.SHA256WithRSA,
	"RS384": x509.SHA384WithRSA,
	"RS512": x509.SHA512WithRSA,
	"PS256": x509.SHA256WithRSAPSS,
	"PS384": x509.SHA384WithRSAPSS,
	"PS512": x509.SHA512WithRSAPSS,
	"ES256": x509.ECDSAWithSHA256,
	"ES384": x509.ECDSAWithSHA384,
	"ES512": x509.ECDSAWithSHA512,
}

// GenerateHsmCertificate creates the self-signed certificate in PEM of the cert's key in the HSM
func GenerateHsmCertificate(cert *Cert) (string, error) {
	if cert.Type != "x509" {
		return "", fmt.Errorf("only the x509 certs that sign tokens can keep the private key in the HSM, the cert: %s is %s", cert.GetId(), cert.Type)
	}
	signer, err := getCertSignerEx(cert, true)
	if err != nil {
		return "", err
	}
	if err = checkHsmKeyAlgorithm(cert.CryptoAlgorithm, signer.Public()); err != nil {
		return "", err
	}

	expireInYears := cert.ExpireInYears
	if expireInYears <= 0 {
		expireInYears = 1
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", err
	}
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: cert.Name, Organization: []string{cert.Owner}},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().AddDate(expireInYears, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509SignatureAlgorithms[cert.CryptoAlgorithm],
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, signer.Public(), signer)
	if err != nil {
		return "", fmt.Errorf("failed to sign the certificate in the HSM: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})), nil
}

// getCertificatePublicKey reads the public key of a PEM certificate
func getCertificatePublicKey(certificate string) (crypto.PublicKey, error) {
	block, _ := pem.Decode([]byte(certificate))
	if block == nil {
		return nil, errors.New("the certificate is not in PEM")
	}
	parsed, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return parsed.PublicKey, nil
}

// populateHsmContent checks the HSM key of a cert being saved and creates its certificate when there is none; the
// private key field stays empty
func (p *Cert) populateHsmContent() error {
	if err := edition.Require(edition.Enterprise, edition.FeatureHsm); err != nil {
		return err
	}
	if p.Type != "x509" {
		return fmt.Errorf("only the x509 certs that sign tokens can keep the private key in the HSM, the cert: %s is %s", p.GetId(), p.Type)
	}
	p.KeyReference = strings.TrimSpace(p.KeyReference)
	if p.KeyReference == "" {
		return fmt.Errorf("the key reference of the cert: %s should not be empty", p.GetId())
	}
	p.PrivateKey = ""

	if p.Certificate == "" {
		certificate, err := GenerateHsmCertificate(p)
		if err != nil {
			return err
		}
		p.Certificate = certificate
	} else {
		signer, err := getCertSigner(p)
		if err != nil {
			return err
		}
		if err = checkHsmKeyAlgorithm(p.CryptoAlgorithm, signer.Public()); err != nil {
			return err
		}
		public, err := getCertificatePublicKey(p.Certificate)
		if err != nil {
			return err
		}
		key, ok := public.(interface{ Equal(crypto.PublicKey) bool })
		if !ok || !key.Equal(signer.Public()) {
			return fmt.Errorf("the certificate of the cert: %s doesn't match the HSM key %q, generate the certificate again", p.GetId(), p.KeyReference)
		}
	}

	if expireTime, err := util.GetCertExpireTime(p.Certificate); err == nil {
		p.ExpireTime = expireTime
	}
	return nil
}

type HsmTestResult struct {
	Algorithm  string  `json:"algorithm"`
	KeyType    string  `json:"keyType"`
	KeyId      string  `json:"keyId"`
	Token      string  `json:"token"`
	ElapsedMs  float64 `json:"elapsedMs"`
	VerifiedBy string  `json:"verifiedBy"`
}

// TestHsmCert signs a test token with the cert's HSM key and verifies it with the cert's certificate, as a relying
// party does with the JWKS
func TestHsmCert(cert *Cert) (*HsmTestResult, error) {
	if !cert.IsHsmKey() {
		return nil, fmt.Errorf("the private key of the cert: %s is not in the HSM", cert.GetId())
	}
	signer, err := getCertSigner(cert)
	if err != nil {
		return nil, err
	}
	if err = checkHsmKeyAlgorithm(cert.CryptoAlgorithm, signer.Public()); err != nil {
		return nil, err
	}
	method := jwt.GetSigningMethod(cert.CryptoAlgorithm)
	if method == nil {
		return nil, fmt.Errorf("unsupported crypto algorithm: %s", cert.CryptoAlgorithm)
	}
	token := jwt.NewWithClaims(method, jwt.MapClaims{"sub": "hsm-signature-test", "iat": time.Now().Unix()})
	token.Header["kid"] = cert.GetKeyId()
	start := time.Now()
	tokenString, err := signJwtToken(token, signer)
	if err != nil {
		return nil, err
	}
	elapsed := time.Since(start)

	public, err := getCertificatePublicKey(cert.Certificate)
	if err != nil {
		return nil, err
	}
	_, err = jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return public, nil
	}, jwt.WithValidMethods([]string{method.Alg()}))
	if err != nil {
		return nil, fmt.Errorf("the HSM signed the test token, but the certificate doesn't verify it: %v", err)
	}

	keyType := "EC"
	switch key := signer.Public().(type) {
	case *rsa.PublicKey:
		keyType = fmt.Sprintf("RSA %d", key.N.BitLen())
	case *ecdsa.PublicKey:
		keyType = "EC " + key.Curve.Params().Name
	}
	return &HsmTestResult{
		Algorithm:  method.Alg(),
		KeyType:    keyType,
		KeyId:      cert.GetKeyId(),
		Token:      tokenString,
		ElapsedMs:  math.Round(float64(elapsed.Microseconds())/10) / 100,
		VerifiedBy: "certificate",
	}, nil
}

// signJwtToken signs a token with a private key, or with an HSM signer through a signing method of the same alg
func signJwtToken(token *jwt.Token, key interface{}) (string, error) {
	if hsm.IsSigner(key) {
		token.Method = hsm.JwtSigningMethod(token.Method)
	}
	return token.SignedString(key)
}

// newCertXmlSigningContext returns the XML signing context of a cert for SAML, with an HSM key only for RSA keys
func newCertXmlSigningContext(cert *Cert, certificate string) (*dsig.SigningContext, error) {
	if !cert.IsHsmKey() {
		return dsig.NewDefaultSigningContext(&X509Key{PrivateKey: cert.PrivateKey, X509Certificate: certificate}), nil
	}

	signer, err := getCertSigner(cert)
	if err != nil {
		return nil, err
	}
	if _, ok := signer.Public().(*rsa.PublicKey); !ok {
		return nil, fmt.Errorf("the HSM key of the cert: %s is not an RSA key, SAML signing needs an RSA key", cert.GetId())
	}
	der, err := base64.StdEncoding.DecodeString(certificate)
	if err != nil {
		return nil, err
	}
	return dsig.NewSigningContext(signer, [][]byte{der})
}

// GetCertTlsCertificate returns the TLS certificate of a cert, signing the handshakes in the HSM when its key is there
func GetCertTlsCertificate(cert *Cert) (tls.Certificate, error) {
	if !cert.IsHsmKey() {
		return tls.X509KeyPair([]byte(cert.Certificate), []byte(cert.PrivateKey))
	}

	signer, err := getCertSigner(cert)
	if err != nil {
		return tls.Certificate{}, err
	}
	block, _ := pem.Decode([]byte(cert.Certificate))
	if block == nil {
		return tls.Certificate{}, fmt.Errorf("the certificate of the cert: %s is not in PEM", cert.GetId())
	}
	return tls.Certificate{Certificate: [][]byte{block.Bytes}, PrivateKey: signer}, nil
}

var hsmSecrets sync.Map

// getHsmSecret derives a secret for a purpose from the cert's HSM key, for what used the private key itself as a
// secret: an RSA PKCS #1 v1.5 signature is deterministic, an EC one isn't, so that falls back to the HSM PIN in
// app.conf
func getHsmSecret(cert *Cert, purpose string) (string, error) {
	cacheKey := cert.KeyReference + "|" + cert.Certificate + "|" + purpose
	if secret, ok := hsmSecrets.Load(cacheKey); ok {
		return secret.(string), nil
	}

	signer, err := getCertSigner(cert)
	if err != nil {
		return "", err
	}

	var secret string
	if _, ok := signer.Public().(*rsa.PublicKey); ok {
		digest := sha256.Sum256([]byte("casdoor-hsm-secret:" + purpose))
		signature, err := signer.Sign(rand.Reader, digest[:], crypto.SHA256)
		if err != nil {
			return "", err
		}
		secret = hex.EncodeToString(signature)
	} else {
		config := hsm.GetConfig()
		if config.Pin == "" {
			return "", fmt.Errorf("the HSM key of the cert: %s can't derive a secret for %s, set hsmPin in app.conf", cert.GetId(), purpose)
		}
		sum := sha256.Sum256([]byte(purpose + "|" + config.Pin + "|" + cert.Certificate))
		secret = hex.EncodeToString(sum[:])
	}

	hsmSecrets.Store(cacheKey, secret)
	return secret, nil
}
