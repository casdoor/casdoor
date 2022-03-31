package object

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/beevik/etree"
	"github.com/golang-jwt/jwt/v4"
	uuid "github.com/satori/go.uuid"
)

func NewSamlResponse(user *User, host string, publicKey string, destination string, iss string, redirectUri []string) (*etree.Element, error) {
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}
	now := time.Now().UTC().Format(time.RFC3339)
	expireTime := time.Now().UTC().Add(time.Hour * 24).Format(time.RFC3339)
	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	samlResponse.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	arId := uuid.NewV4()

	samlResponse.CreateAttr("ID", fmt.Sprintf("_%s", arId))
	samlResponse.CreateAttr("Version", "2.0")
	samlResponse.CreateAttr("IssueInstant", now)
	samlResponse.CreateAttr("Destination", destination)
	samlResponse.CreateAttr("InResponseTo", fmt.Sprintf("Casdoor_%s", arId))
	samlResponse.CreateElement("saml:Issuer").SetText(host)

	samlResponse.CreateElement("samlp:Status").CreateElement("samlp:StatusCode").CreateAttr("Value", "urn:oasis:names:tc:SAML:2.0:status:Success")

	assertion := samlResponse.CreateElement("saml:Assertion")
	assertion.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	assertion.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	assertion.CreateAttr("ID", fmt.Sprintf("_%s", uuid.NewV4()))
	assertion.CreateAttr("Version", "2.0")
	assertion.CreateAttr("IssueInstant", now)
	assertion.CreateElement("saml:Issuer").SetText(host)
	subject := assertion.CreateElement("saml:Subject")
	subject.CreateElement("saml:NameID").SetText(user.Email)
	subjectConfirmation := subject.CreateElement("saml:SubjectConfirmation")
	subjectConfirmation.CreateAttr("Method", "urn:oasis:names:tc:SAML:2.0:cm:bearer")
	subjectConfirmationData := subjectConfirmation.CreateElement("saml:SubjectConfirmationData")
	subjectConfirmationData.CreateAttr("InResponseTo", fmt.Sprintf("_%s", arId))
	subjectConfirmationData.CreateAttr("Recipient", destination)
	subjectConfirmationData.CreateAttr("NotOnOrAfter", expireTime)
	condition := assertion.CreateElement("saml:Conditions")
	condition.CreateAttr("NotBefore", now)
	condition.CreateAttr("NotOnOrAfter", expireTime)
	audience := condition.CreateElement("saml:AudienceRestriction")
	audience.CreateElement("saml:Audience").SetText(iss)
	for _, value := range redirectUri {
		audience.CreateElement("saml:Audience").SetText(value)
	}
	authnStatement := assertion.CreateElement("saml:AuthnStatement")
	authnStatement.CreateAttr("AuthnInstant", now)
	authnStatement.CreateAttr("SessionIndex", fmt.Sprintf("_%s", uuid.NewV4()))
	authnStatement.CreateAttr("SessionNotOnOrAfter", expireTime)
	authnStatement.CreateElement("saml:AuthnContext").CreateElement("saml:AuthnContextClassRef").SetText("urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport")

	attributes := assertion.CreateElement("saml:AttributeStatement")
	email := attributes.CreateElement("saml:Attribute")
	email.CreateAttr("Name", "email")
	email.CreateAttr("NameFormat", "urn:oasis:names:tc:SAML:2.0:attrname-format:basic")
	email.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string").Element().SetText(user.Email)

	return samlResponse, nil

}

type X509Key struct {
	X509Certificate string
	PrivateKey      string
}

func (x X509Key) GetKeyPair() (privateKey *rsa.PrivateKey, cert []byte, err error) {
	cert, _ = base64.StdEncoding.DecodeString(x.X509Certificate)
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(x.PrivateKey))
	return privateKey, cert, err
}
