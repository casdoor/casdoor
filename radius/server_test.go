// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

//go:build !skipCi
// +build !skipCi

package radius

import (
	"context"
	"testing"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestAccessRequestRejected(t *testing.T) {
	packet := radius.New(radius.CodeAccessRequest, []byte(`secret`))
	rfc2865.UserName_SetString(packet, "admin")
	rfc2865.UserPassword_SetString(packet, "12345")
	rfc2865.Class_SetString(packet, "built-in")
	response, err := radius.Exchange(context.Background(), packet, "localhost:1812")
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != radius.CodeAccessReject {
		t.Fatalf("Expected %v, got %v", radius.CodeAccessReject, response.Code)
	}
}

func TestAccessRequestAccepted(t *testing.T) {
	packet := radius.New(radius.CodeAccessRequest, []byte(`secret`))
	rfc2865.UserName_SetString(packet, "admin")
	rfc2865.UserPassword_SetString(packet, "123")
	rfc2865.Class_SetString(packet, "built-in")
	response, err := radius.Exchange(context.Background(), packet, "localhost:1812")
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != radius.CodeAccessAccept {
		t.Fatalf("Expected %v, got %v", radius.CodeAccessAccept, response.Code)
	}
}
