// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

"github.com/casdoor/casdoor/i18n"
"github.com/casdoor/casdoor/idv"
)

func VerifyIdentification(user *User, provider *Provider, realName string, lang string) (bool, error) {
if provider == nil {
return false, fmt.Errorf(i18n.Translate(lang, "provider:No ID verification provider configured"))
}

if provider.Category != "ID Verification" {
return false, fmt.Errorf(i18n.Translate(lang, "provider:Provider is not an ID verification provider"))
}

if user.IdCard == "" {
return false, fmt.Errorf(i18n.Translate(lang, "user:User ID card is not set"))
}

if user.IdCardType == "" {
return false, fmt.Errorf(i18n.Translate(lang, "user:User ID card type is not set"))
}

if realName == "" {
return false, fmt.Errorf(i18n.Translate(lang, "user:Real name cannot be empty"))
}

idvProvider := idv.GetIdvProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.Endpoint)
if idvProvider == nil {
return false, fmt.Errorf(i18n.Translate(lang, "provider:Unsupported ID verification provider type: %s"), provider.Type)
}

verified, err := idvProvider.VerifyIdentification(user.IdCardType, user.IdCard, realName)
if err != nil {
return false, err
}

if verified {
user.RealName = realName
_, err = UpdateUser(user.GetId(), user, []string{"real_name"}, false)
if err != nil {
return false, err
}
}

return verified, nil
}
