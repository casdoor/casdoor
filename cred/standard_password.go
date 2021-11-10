package cred

import (
	"bytes"
	"errors"
	"strings"
)

const (
	StandardPasswordInvalidError = "stand password format invalid"
)

type StandardPassword struct {
	Type string
	UserSalt string
	OrganizationSalt string
	PasswordHash string
	Chunk []string
}

func ParseStandardPassword(input string) (*StandardPassword, error) {
	data := strings.Split(input, "$")
	dataLen := len(data)

	// the standard password should at least have 3 parts
	// $type$plain
	// the first $ will be parsed as an empty string
	if dataLen < 3 {
		return nil, errors.New(StandardPasswordInvalidError)
	}

	ret := new(StandardPassword)
	ret.Chunk = data[2:]
	if dataLen == 3 {
		ret.Type = data[1]
		ret.OrganizationSalt = ""
		ret.UserSalt = ""
		ret.PasswordHash = data[2]
	} else if dataLen == 4 {
		ret.Type = data[1]
		ret.OrganizationSalt = data[2]
		ret.UserSalt = ""
		ret.PasswordHash = data[3]
	} else if dataLen >= 5 {
		ret.Type = data[1]
		ret.OrganizationSalt = data[2]
		ret.UserSalt = data[3]
		ret.PasswordHash = data[4]
	}

	return ret, nil
}

func (sp *StandardPassword) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("$")
	buffer.WriteString(sp.Type)
	buffer.WriteString("$")
	buffer.WriteString(sp.OrganizationSalt)
	buffer.WriteString("$")
	if sp.UserSalt != "" {
		buffer.WriteString(sp.UserSalt)
		buffer.WriteString("$")
	}
	buffer.WriteString(sp.PasswordHash)

	return buffer.String()
}