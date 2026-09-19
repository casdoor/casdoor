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

// Package authenticator implements the RADIUS Message-Authenticator attribute
// (RFC 2869 section 5.14), the mitigation for Blast-RADIUS (CVE-2024-3596).
package authenticator

import (
	"crypto/hmac"
	"crypto/md5"
	"errors"
	"fmt"

	"layeh.com/radius"
	"layeh.com/radius/rfc2869"
)

var (
	ErrMissing = errors.New("RADIUS Message-Authenticator is missing")
	ErrInvalid = errors.New("RADIUS Message-Authenticator is invalid")
)

func Present(packet *radius.Packet) bool {
	return packet != nil && packet.Get(rfc2869.MessageAuthenticator_Type) != nil
}

// Sign sets the Message-Authenticator of a request, or of a response whose
// Authenticator field still holds the request authenticator (as produced by
// Packet.Response()).
func Sign(packet *radius.Packet) error {
	if len(packet.Secret) == 0 {
		return errors.New("RADIUS shared secret is empty")
	}
	if err := rfc2869.MessageAuthenticator_Set(packet, make([]byte, md5.Size)); err != nil {
		return err
	}
	wire, err := packet.MarshalBinary()
	if err != nil {
		return err
	}
	return rfc2869.MessageAuthenticator_Set(packet, compute(wire, packet.Secret))
}

func VerifyRequest(request *radius.Packet) error {
	return verify(request, request.Authenticator)
}

func VerifyResponse(response, request *radius.Packet) error {
	if response.Identifier != request.Identifier {
		return errors.New("RADIUS response identifier does not match the request")
	}
	return verify(response, request.Authenticator)
}

func verify(packet *radius.Packet, requestAuthenticator [16]byte) error {
	wire, err := packet.MarshalBinary()
	if err != nil {
		return err
	}
	copy(wire[4:20], requestAuthenticator[:])

	var got []byte
	for i := 20; i < len(wire); {
		if i+2 > len(wire) {
			return fmt.Errorf("RADIUS packet has a truncated attribute")
		}
		attrType, attrLen := wire[i], int(wire[i+1])
		if attrLen < 2 || i+attrLen > len(wire) {
			return fmt.Errorf("RADIUS packet has an attribute with invalid length %d", attrLen)
		}
		if attrType == byte(rfc2869.MessageAuthenticator_Type) {
			if got != nil || attrLen != 2+md5.Size {
				return ErrInvalid
			}
			value := wire[i+2 : i+attrLen]
			got = append([]byte(nil), value...)
			clear(value)
		}
		i += attrLen
	}
	if got == nil {
		return ErrMissing
	}
	if !hmac.Equal(got, compute(wire, packet.Secret)) {
		return ErrInvalid
	}
	return nil
}

func compute(wire, secret []byte) []byte {
	mac := hmac.New(md5.New, secret)
	mac.Write(wire)
	return mac.Sum(nil)
}

type responseWriter struct {
	radius.ResponseWriter
}

// NewResponseWriter wraps w so that every response it writes carries a
// Message-Authenticator.
func NewResponseWriter(w radius.ResponseWriter) radius.ResponseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (w *responseWriter) Write(packet *radius.Packet) error {
	if err := Sign(packet); err != nil {
		return err
	}
	return w.ResponseWriter.Write(packet)
}
