// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"testing"

	"github.com/casdoor/casdoor/util"
)

func TestGenerateRsaKeys(t *testing.T) {
	fileId := "token_jwt_key"
	publicKey, privateKey := generateRsaKeys(4096, 20, "Casdoor Cert", "Casdoor Organization")

	// Write certificate (aka public key) to file.
	util.WriteStringToPath(publicKey, fmt.Sprintf("%s.pem", fileId))

	// Write private key to file.
	util.WriteStringToPath(privateKey, fmt.Sprintf("%s.key", fileId))
}
