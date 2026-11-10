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
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/action"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
)

// CertLoginPolicy lets the users of an organization sign in with a digital certificate, usually kept in a UKey: the
// certificate must be issued by one of the trusted CAs, and one of its fields is matched against a user attribute
type CertLoginPolicy struct {
	Enabled bool `json:"enabled"`
	// the PEM certificates of the trusted CAs, RSA or ECDSA
	TrustedCas string `json:"trustedCas"`
	// the certificate field that identifies the user, one of CertLoginFields
	CertField string `json:"certField"`
	// the user attribute it is matched against, one of CertLoginUserFields
	UserField string `json:"userField"`
	// optional, the CRLs that the certificate is checked against, separated by commas
	CrlUrl string `json:"crlUrl"`
	// the full URL of /api/login/cert on the port that asks for client certificates, the login page's button goes
	// there, the same origin's /api/login/cert when empty (a reverse proxy passing the client certificate)
	LoginUrl string `json:"loginUrl"`
	// optional, an action of kind "cert" installed from the Marketplace that checks the certificates instead of
	// Casdoor, for the algorithms Casdoor doesn't verify itself
	Verifier string `json:"verifier"`
}

const (
	CertFieldCommonName   = "CN"
	CertFieldSerialNumber = "SERIALNUMBER"
	CertFieldEmail        = "Email"
	CertFieldUid          = "UID"
	CertFieldFingerprint  = "Fingerprint"
)

var CertLoginFields = []string{CertFieldCommonName, CertFieldSerialNumber, CertFieldEmail, CertFieldUid, CertFieldFingerprint}

// CertFingerprintProperty is the user property that binds certificates to a user, by their SHA-256 fingerprints in
// hex separated by commas
const CertFingerprintProperty = "certFingerprint"

var CertLoginUserFields = []string{"name", "idCard", "email", "phone", CertFingerprintProperty}

const SigninMethodCertificate = "Certificate"

var (
	oidEmailAddress = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}
	oidUid          = asn1.ObjectIdentifier{0, 9, 2342, 19200300, 100, 1, 1}
)

func (application *Application) IsCertLoginEnabled() bool {
	return application.HasSigninMethod(SigninMethodCertificate)
}

func (policy *CertLoginPolicy) isEnabled() bool {
	return policy != nil && policy.Enabled
}

func parseTrustedCas(text string) (*x509.CertPool, int, error) {
	pool := x509.NewCertPool()
	count := 0
	rest := []byte(text)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, 0, err
		}
		pool.AddCert(certificate)
		count++
	}
	return pool, count, nil
}

func countPemCertificates(text string) int {
	count := 0
	rest := []byte(text)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			return count
		}
		if block.Type == "CERTIFICATE" {
			count++
		}
	}
}

func validateCertLoginPolicy(policy *CertLoginPolicy, lang string) error {
	if !policy.isEnabled() {
		return nil
	}
	count := countPemCertificates(policy.TrustedCas)
	if policy.Verifier == "" {
		_, _, err := parseTrustedCas(policy.TrustedCas)
		if err != nil {
			return fmt.Errorf(i18n.Translate(lang, "enterprise:The trusted CA certificates are invalid: %s"), err.Error())
		}
	}
	if count == 0 {
		return errors.New(i18n.Translate(lang, "enterprise:Add at least one trusted CA certificate in PEM"))
	}
	if !util.InSlice(CertLoginFields, policy.CertField) || !util.InSlice(CertLoginUserFields, policy.UserField) {
		return errors.New(i18n.Translate(lang, "enterprise:Choose the certificate field and the user attribute it matches"))
	}
	values := strings.Fields(strings.ReplaceAll(policy.CrlUrl, ",", " "))
	for _, value := range append(values, policy.LoginUrl) {
		if value == "" {
			continue
		}
		parsed, err := url.Parse(value)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf(i18n.Translate(lang, "enterprise:The URL: %s is invalid"), value)
		}
	}
	return nil
}

// ReadCertificateDers reads the DER certificates in PEM blocks, or a certificate in base64 DER, the first being the
// user's own one
func ReadCertificateDers(text string) ([][]byte, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("the certificate is empty")
	}
	if strings.Contains(text, "-----BEGIN") {
		res := [][]byte{}
		rest := []byte(text)
		for {
			var block *pem.Block
			block, rest = pem.Decode(rest)
			if block == nil {
				break
			}
			if block.Type == "CERTIFICATE" {
				res = append(res, block.Bytes)
			}
		}
		if len(res) == 0 {
			return nil, errors.New("no certificate is found in the PEM")
		}
		return res, nil
	}

	text = strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' || r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, text)
	der, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		der, err = base64.RawURLEncoding.DecodeString(strings.TrimRight(text, "="))
		if err != nil {
			return nil, errors.New("the certificate is neither PEM nor base64 DER")
		}
	}
	return [][]byte{der}, nil
}

func parseCertificateDers(ders [][]byte) ([]*x509.Certificate, error) {
	res := []*x509.Certificate{}
	for _, der := range ders {
		certificates, err := x509.ParseCertificates(der)
		if err != nil {
			return nil, err
		}
		res = append(res, certificates...)
	}
	if len(res) == 0 {
		return nil, errors.New("no certificate is found")
	}
	return res, nil
}

func getCertFingerprint(der []byte) string {
	sum := sha256.Sum256(der)
	return hex.EncodeToString(sum[:])
}

func getSubjectAttribute(certificate *x509.Certificate, oid asn1.ObjectIdentifier) string {
	for _, name := range certificate.Subject.Names {
		if name.Type.Equal(oid) {
			if value, ok := name.Value.(string); ok {
				return value
			}
		}
	}
	return ""
}

// getCertIdentity returns the certificate field that identifies the user
func getCertIdentity(certificate *x509.Certificate, field string) string {
	switch field {
	case CertFieldCommonName:
		return strings.TrimSpace(certificate.Subject.CommonName)
	case CertFieldSerialNumber:
		return strings.TrimSpace(certificate.Subject.SerialNumber)
	case CertFieldEmail:
		if len(certificate.EmailAddresses) > 0 {
			return strings.TrimSpace(certificate.EmailAddresses[0])
		}
		return strings.TrimSpace(getSubjectAttribute(certificate, oidEmailAddress))
	case CertFieldUid:
		return strings.TrimSpace(getSubjectAttribute(certificate, oidUid))
	case CertFieldFingerprint:
		return getCertFingerprint(certificate.Raw)
	}
	return ""
}

var crlCache sync.Map

type cachedCrl struct {
	crl       *x509.RevocationList
	fetchTime time.Time
}

var crlDataCache sync.Map

type cachedCrlData struct {
	data      []byte
	fetchTime time.Time
}

// getCrlData fetches a CRL, DER or PEM, kept for 10 minutes
func getCrlData(crlUrl string) ([]byte, error) {
	if value, ok := crlDataCache.Load(crlUrl); ok {
		cached := value.(*cachedCrlData)
		if time.Since(cached.fetchTime) < 10*time.Minute {
			return cached.data, nil
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(crlUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	crlDataCache.Store(crlUrl, &cachedCrlData{data: data, fetchTime: time.Now()})
	return data, nil
}

func getCrl(crlUrl string) (*x509.RevocationList, error) {
	if value, ok := crlCache.Load(crlUrl); ok {
		cached := value.(*cachedCrl)
		if time.Since(cached.fetchTime) < 10*time.Minute && (cached.crl.NextUpdate.IsZero() || time.Now().Before(cached.crl.NextUpdate)) {
			return cached.crl, nil
		}
	}

	data, err := getCrlData(crlUrl)
	if err != nil {
		return nil, err
	}
	if block, _ := pem.Decode(data); block != nil {
		data = block.Bytes
	}
	crl, err := x509.ParseRevocationList(data)
	if err != nil {
		return nil, err
	}
	crlCache.Store(crlUrl, &cachedCrl{crl: crl, fetchTime: time.Now()})
	return crl, nil
}

// checkCrl checks the certificate against the CRLs, separated by commas or spaces, that its issuer signed: a CRL
// of another CA is skipped, and one that can't be fetched fails the sign-in
func checkCrl(crlUrls string, chain []*x509.Certificate, lang string) error {
	certificate := chain[0]
	for _, crlUrl := range strings.Fields(strings.ReplaceAll(crlUrls, ",", " ")) {
		crl, err := getCrl(crlUrl)
		if err != nil {
			return fmt.Errorf(i18n.Translate(lang, "enterprise:Failed to check the certificate revocation list: %s"), err.Error())
		}
		signedByIssuer := false
		for _, issuer := range chain[1:] {
			if crl.CheckSignatureFrom(issuer) == nil {
				signedByIssuer = true
				break
			}
		}
		if !signedByIssuer {
			continue
		}
		for _, entry := range crl.RevokedCertificateEntries {
			if entry.SerialNumber != nil && entry.SerialNumber.Cmp(certificate.SerialNumber) == 0 {
				return errors.New(i18n.Translate(lang, "enterprise:The certificate has been revoked"))
			}
		}
	}
	return nil
}

func findUniqueUser(owner string, column string, value string, lang string) (*User, error) {
	users := []*User{}
	err := ormer.Engine.Where("owner = ?", owner).And(fmt.Sprintf("%s = ?", column), value).Limit(2).Find(&users)
	if err != nil {
		return nil, err
	}
	if len(users) > 1 {
		return nil, fmt.Errorf(i18n.Translate(lang, "enterprise:More than one user matches the certificate: %s"), value)
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

func hasFingerprint(user *User, fingerprint string) bool {
	for _, item := range strings.FieldsFunc(user.Properties[CertFingerprintProperty], func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		if strings.EqualFold(strings.ReplaceAll(item, ":", ""), fingerprint) {
			return true
		}
	}
	return false
}

// findCertUser finds the user whose attribute matches the certificate's identity
func findCertUser(organization *Organization, userField string, value string, lang string) (*User, error) {
	owner := organization.Name
	switch userField {
	case "name":
		return findUniqueUser(owner, "name", value, lang)
	case "email":
		return findUniqueUser(owner, "email", strings.ToLower(value), lang)
	case "phone":
		return GetUserByPhone(owner, value)
	case "idCard":
		policy := GetEnterprisePolicy(organization)
		if policy == nil || !util.InSlice(policy.EncryptedFields, "idCard") {
			return findUniqueUser(owner, "id_card", value, lang)
		}
		// the encrypted ID card numbers can't be queried, they are compared once decrypted
		users := []*User{}
		err := ormer.Engine.Where("owner = ? and id_card <> ''", owner).Find(&users)
		if err != nil {
			return nil, err
		}
		var res *User
		for _, user := range users {
			if strings.EqualFold(user.IdCard, value) {
				if res != nil {
					return nil, fmt.Errorf(i18n.Translate(lang, "enterprise:More than one user matches the certificate: %s"), value)
				}
				res = user
			}
		}
		return res, nil
	case CertFingerprintProperty:
		value = strings.ToLower(value)
		users := []*User{}
		err := ormer.Engine.Where("owner = ?", owner).And("properties like ?", "%"+value+"%").Find(&users)
		if err != nil {
			return nil, err
		}
		var res *User
		for _, user := range users {
			if hasFingerprint(user, value) {
				if res != nil {
					return nil, fmt.Errorf(i18n.Translate(lang, "enterprise:More than one user matches the certificate: %s"), value)
				}
				res = user
			}
		}
		return res, nil
	}
	return nil, nil
}

// CertLoginInput is what a certificate sign-in presents: the DER certificates, the user's own one first, and for a
// UKey plug-in the challenge and the signature the UKey made of it
type CertLoginInput struct {
	Ders      [][]byte
	Message   []byte
	Signature []byte
}

// VerifyCertLogin checks a certificate sign-in against the organization's trusted CAs, the validity period, the CRL
// and the signature of the challenge if any, by Casdoor itself or by the organization's verifier action, and returns
// the user it identifies and the fingerprint of the user's certificate
func VerifyCertLogin(organizationName string, input *CertLoginInput, lang string) (*User, string, error) {
	if err := edition.Require(edition.Enterprise, edition.FeatureCertLogin); err != nil {
		return nil, "", err
	}
	if input == nil || len(input.Ders) == 0 {
		return nil, "", errors.New(i18n.Translate(lang, "enterprise:No client certificate was presented"))
	}

	organization, err := GetOrganization(util.GetId("admin", organizationName))
	if err != nil {
		return nil, "", err
	}
	if organization == nil {
		return nil, "", fmt.Errorf(i18n.Translate(lang, "auth:The organization: %s does not exist"), organizationName)
	}
	policy := GetEnterprisePolicy(organization)
	if policy == nil || !policy.CertLogin.isEnabled() {
		return nil, "", fmt.Errorf(i18n.Translate(lang, "enterprise:Certificate login is not enabled for the organization: %s"), organizationName)
	}
	certLogin := policy.CertLogin

	var fields map[string]string
	if certLogin.Verifier != "" {
		fields, err = verifyCertByAction(organization.Name, certLogin, input, lang)
	} else {
		fields, err = verifyCert(certLogin, input, lang)
	}
	if err != nil {
		return nil, "", err
	}

	identity := strings.TrimSpace(fields[certLogin.CertField])
	if identity == "" {
		return nil, "", fmt.Errorf(i18n.Translate(lang, "enterprise:The certificate has no %s"), certLogin.CertField)
	}

	user, err := findCertUser(organization, certLogin.UserField, identity, lang)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", fmt.Errorf(i18n.Translate(lang, "enterprise:No user matches the certificate: %s"), identity)
	}
	if user.IsDeleted || user.IsForbidden {
		return nil, "", errors.New(i18n.Translate(lang, "check:The user is forbidden to sign in, please contact the administrator"))
	}
	return user, fields[CertFieldFingerprint], nil
}

// verifyCert checks an RSA or ECDSA certificate sign-in and returns the fields of the user's certificate
func verifyCert(certLogin *CertLoginPolicy, input *CertLoginInput, lang string) (map[string]string, error) {
	chain, err := parseCertificateDers(input.Ders)
	if err != nil {
		return nil, fmt.Errorf(i18n.Translate(lang, "enterprise:The certificate is invalid: %s"), err.Error())
	}
	certificate := chain[0]
	if input.Signature != nil && !verifyCertSignature(certificate, input.Message, input.Signature) {
		return nil, errors.New(i18n.Translate(lang, "enterprise:The signature of the challenge is invalid"))
	}

	now := time.Now()
	if now.Before(certificate.NotBefore) || now.After(certificate.NotAfter) {
		return nil, errors.New(i18n.Translate(lang, "enterprise:The certificate has expired or is not valid yet"))
	}

	roots, _, err := parseTrustedCas(certLogin.TrustedCas)
	if err != nil {
		return nil, err
	}
	intermediates := x509.NewCertPool()
	for _, item := range chain[1:] {
		intermediates.AddCert(item)
	}
	chains, err := certificate.Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   now,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	if err != nil {
		var invalidErr x509.CertificateInvalidError
		if errors.As(err, &invalidErr) && invalidErr.Reason == x509.Expired {
			return nil, errors.New(i18n.Translate(lang, "enterprise:The certificate has expired or is not valid yet"))
		}
		return nil, fmt.Errorf(i18n.Translate(lang, "enterprise:The certificate is not issued by a trusted CA: %s"), err.Error())
	}

	if certLogin.CrlUrl != "" {
		if err = checkCrl(certLogin.CrlUrl, chains[0], lang); err != nil {
			return nil, err
		}
	}

	fields := map[string]string{}
	for _, field := range CertLoginFields {
		fields[field] = getCertIdentity(certificate, field)
	}
	return fields, nil
}

// verifyCertByAction has the organization's verifier action check the sign-in, with the CRLs that Casdoor fetched
func verifyCertByAction(owner string, certLogin *CertLoginPolicy, input *CertLoginInput, lang string) (map[string]string, error) {
	m, _, err := getInstalledAction(owner, certLogin.Verifier, "cert")
	if err != nil {
		return nil, err
	}

	req := &action.CertVerifyRequest{
		Certificates: []string{},
		TrustedCas:   certLogin.TrustedCas,
		Crls:         []string{},
		Time:         time.Now().UTC().Format(time.RFC3339),
	}
	for _, der := range input.Ders {
		req.Certificates = append(req.Certificates, base64.StdEncoding.EncodeToString(der))
	}
	for _, crlUrl := range strings.Fields(strings.ReplaceAll(certLogin.CrlUrl, ",", " ")) {
		data, err := getCrlData(crlUrl)
		if err != nil {
			return nil, fmt.Errorf(i18n.Translate(lang, "enterprise:Failed to check the certificate revocation list: %s"), err.Error())
		}
		req.Crls = append(req.Crls, base64.StdEncoding.EncodeToString(data))
	}
	if input.Signature != nil {
		req.Message = base64.StdEncoding.EncodeToString(input.Message)
		req.Signature = base64.StdEncoding.EncodeToString(input.Signature)
	}

	res, err := action.VerifyCertificate(m, req)
	if err != nil {
		return nil, err
	}
	if !res.Valid {
		switch res.Reason {
		case "expired":
			return nil, errors.New(i18n.Translate(lang, "enterprise:The certificate has expired or is not valid yet"))
		case "untrusted":
			return nil, fmt.Errorf(i18n.Translate(lang, "enterprise:The certificate is not issued by a trusted CA: %s"), res.Detail)
		case "revoked":
			return nil, errors.New(i18n.Translate(lang, "enterprise:The certificate has been revoked"))
		case "signature":
			return nil, errors.New(i18n.Translate(lang, "enterprise:The signature of the challenge is invalid"))
		}
		return nil, fmt.Errorf(i18n.Translate(lang, "enterprise:The certificate is invalid: %s"), res.Detail)
	}

	fields := res.Fields
	if fields == nil {
		fields = map[string]string{}
	}
	if fields[CertFieldFingerprint] == "" {
		fields[CertFieldFingerprint] = getCertFingerprint(input.Ders[0])
	}
	return fields, nil
}

// verifyCertSignature checks the signature that a UKey made of the message with the certificate's key: RSA PKCS #1
// v1.5 with SHA-256 or ECDSA with SHA-256, in ASN.1 DER or r and s
func verifyCertSignature(certificate *x509.Certificate, message []byte, signature []byte) bool {
	switch publicKey := certificate.PublicKey.(type) {
	case *ecdsa.PublicKey:
		digest := sha256.Sum256(message)
		size := (publicKey.Curve.Params().BitSize + 7) / 8
		if len(signature) == 2*size {
			r := new(big.Int).SetBytes(signature[:size])
			s := new(big.Int).SetBytes(signature[size:])
			if ecdsa.Verify(publicKey, digest[:], r, s) {
				return true
			}
		}
		return ecdsa.VerifyASN1(publicKey, digest[:], signature)
	case *rsa.PublicKey:
		digest := sha256.Sum256(message)
		return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature) == nil
	}
	return false
}

// DecodeSignature reads a signature in base64, standard or URL-safe, or in hex
func DecodeSignature(text string) ([]byte, error) {
	text = strings.TrimSpace(text)
	if res, err := base64.StdEncoding.DecodeString(text); err == nil {
		return res, nil
	}
	if res, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(text, "=")); err == nil {
		return res, nil
	}
	if res, err := hex.DecodeString(text); err == nil {
		return res, nil
	}
	return nil, errors.New("the signature is not in base64")
}

// CertLoginTicket is a one-time secret of a certificate sign-in, only its hash is kept: a challenge that a UKey
// signs, or a ticket that the port asking for client certificates hands to the login page after checking the
// certificate
type CertLoginTicket struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Kind        string `xorm:"varchar(20)" json:"kind"`
	Application string `xorm:"varchar(100)" json:"application"`
	User        string `xorm:"varchar(100)" json:"user"`
	Fingerprint string `xorm:"varchar(100)" json:"fingerprint"`
	SecretHash  string `xorm:"varchar(100) index" json:"-"`
	ExpireTime  int64  `json:"expireTime"`
	IsUsed      bool   `json:"isUsed"`
}

const (
	CertLoginKindChallenge = "challenge"
	CertLoginKindTicket    = "ticket"

	CertChallengeTimeout = 5 * time.Minute
	certTicketTimeout    = 2 * time.Minute
)

func addCertLoginTicket(kind string, application *Application, user string, fingerprint string, timeout time.Duration) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	secret := base64.RawURLEncoding.EncodeToString(bytes)

	now := time.Now()
	_, _ = ormer.Engine.Where("expire_time < ?", now.Add(-time.Hour).Unix()).Delete(&CertLoginTicket{})

	ticket := &CertLoginTicket{
		Owner:       application.Organization,
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		Kind:        kind,
		Application: application.Name,
		User:        user,
		Fingerprint: fingerprint,
		SecretHash:  HashMagicLinkSecret(secret),
		ExpireTime:  now.Add(timeout).Unix(),
	}
	_, err := ormer.Engine.Insert(ticket)
	if err != nil {
		return "", err
	}
	return secret, nil
}

// AddCertChallenge returns a one-time nonce for a UKey to sign
func AddCertChallenge(application *Application) (string, error) {
	return addCertLoginTicket(CertLoginKindChallenge, application, "", "", CertChallengeTimeout)
}

// AddCertTicket returns a one-time ticket that signs the user in on the login page
func AddCertTicket(application *Application, user *User, fingerprint string) (string, error) {
	return addCertLoginTicket(CertLoginKindTicket, application, user.GetId(), fingerprint, certTicketTimeout)
}

// ConsumeCertLoginTicket claims a challenge or a ticket of the application for a single sign-in
func ConsumeCertLoginTicket(kind string, secret string, application *Application, lang string) (*CertLoginTicket, error) {
	invalidErr := errors.New(i18n.Translate(lang, "enterprise:The certificate sign-in has expired or was already used, please try again"))
	if secret == "" {
		return nil, invalidErr
	}

	ticket := &CertLoginTicket{SecretHash: HashMagicLinkSecret(secret)}
	existed, err := ormer.Engine.Get(ticket)
	if err != nil {
		return nil, err
	}
	if !existed || ticket.Kind != kind || ticket.Application != application.Name || ticket.IsUsed || time.Now().Unix() > ticket.ExpireTime {
		return nil, invalidErr
	}

	affected, err := ormer.Engine.Where("secret_hash = ?", ticket.SecretHash).And("is_used = ?", false).Cols("is_used").Update(&CertLoginTicket{IsUsed: true})
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, invalidErr
	}
	ticket.IsUsed = true
	return ticket, nil
}
