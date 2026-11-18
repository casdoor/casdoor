// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/ee/gm"
	"github.com/casdoor/casdoor/object"
	"github.com/golang-jwt/jwt/v5"
)

// getJwtSigningKey returns the HSM signer of a cert whose key is in the HSM, which signJwtToken signs with, the SM2
// private key for an SM2 token, or the standard private key
func getJwtSigningKey(method jwt.SigningMethod, cert *object.Cert) (interface{}, error) {
	if cert.IsHsmKey() {
		return getCertSigner(cert)
	}
	if method == gm.SigningMethodSm2 {
		if err := edition.Require(edition.Enterprise, edition.FeatureGuomi); err != nil {
			return nil, err
		}
		return gm.ParsePrivateKey(cert.PrivateKey)
	}
	return object.ParseJwtSigningKey(method, cert)
}

func getJwtVerifyKey(token *jwt.Token, certificate string) (interface{}, error) {
	if token.Method == gm.SigningMethodSm2 {
		return gm.ParsePublicKey(certificate)
	}
	return object.ParseJwtVerifyKey(token, certificate)
}
