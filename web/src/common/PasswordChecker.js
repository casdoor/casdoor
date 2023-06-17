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

import i18next from "i18next";

function isValidOption_AtLeast6(password) {
  if (password.length < 6) {
    return i18next.t("user:The password must have at least 6 characters");
  }
  return "";
}

function isValidOption_AtLeast8(password) {
  if (password.length < 8) {
    return i18next.t("user:The password must have at least 8 characters");
  }
  return "";
}

function isValidOption_Aa123(password) {
  const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9]).+$/;
  if (!regex.test(password)) {
    return i18next.t("user:The password must contain at least one uppercase letter, one lowercase letter and one digit");
  }
  return "";
}

function isValidOption_SpecialChar(password) {
  const regex = /^(?=.*[!@#$%^&*]).+$/;
  if (!regex.test(password)) {
    return i18next.t("user:The password must contain at least one special character");
  }
  return "";
}

function isValidOption_NoRepeat(password) {
  const regex = /(.)\1+/;
  if (regex.test(password)) {
    return i18next.t("user:The password must not contain any repeated characters");
  }
  return "";
}

export function checkPasswordComplexity(password, options) {
  if (password.length === 0) {
    return i18next.t("login:Please input your password!");
  }

  if (!options || options.length === 0) {
    options = ["AtLeast6"];
  }

  const checkers = {
    AtLeast6: isValidOption_AtLeast6,
    AtLeast8: isValidOption_AtLeast8,
    Aa123: isValidOption_Aa123,
    SpecialChar: isValidOption_SpecialChar,
    NoRepeat: isValidOption_NoRepeat,
  };

  for (const option of options) {
    const checkerFunc = checkers[option];
    if (checkerFunc) {
      const errorMsg = checkerFunc(password);
      if (errorMsg !== "") {
        return errorMsg;
      }
    }
  }
  return "";
}
