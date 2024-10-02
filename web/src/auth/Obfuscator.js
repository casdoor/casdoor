// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

import CryptoJS from "crypto-js";
import i18next from "i18next";
import * as Setting from "../Setting";

export function getRandomKeyForObfuscator(obfuscatorType) {
  if (obfuscatorType === "DES") {
    return Setting.getRandomHexKey(16);
  } else if (obfuscatorType === "AES") {
    return Setting.getRandomHexKey(32);
  } else {
    return "";
  }
}

export function checkPasswordObfuscatorKey(passwordObfuscatorType, passwordObfuscatorKey) {
  if (passwordObfuscatorType === "Plain" && passwordObfuscatorKey !== "") {
    return [false, i18next.t("organization:The password obfuscator key should be empty")];
  } else if (passwordObfuscatorType === "AES") {
    const regex = /^[1-9a-f]{32}$/;
    if (!regex.test(passwordObfuscatorKey)) {
      return [false, i18next.t("organization:The password obfuscator key doesn't match the AES regex") + " ^[1-9a-f]{32}$"];
    }
  } else if (passwordObfuscatorType === "DES") {
    const regex = /^[1-9a-f]{16}$/;
    if (!regex.test(passwordObfuscatorKey)) {
      return [false, i18next.t("organization:The password obfuscator key doesn't match the DES regex") + " ^[1-9a-f]{16}$"];
    }
  }
  return [true, ""];
}

function encrypt(cipher, key, iv, password) {
  const encrypted = cipher.encrypt(
    CryptoJS.enc.Hex.parse(Buffer.from(password, "utf-8").toString("hex")),
    CryptoJS.enc.Hex.parse(key),
    {
      iv: iv,
      mode: CryptoJS.mode.CBC,
      pad: CryptoJS.pad.Pkcs7,
    }
  );
  return iv.concat(encrypted.ciphertext).toString(CryptoJS.enc.Hex);
}

export function encryptByDes(key, password) {
  const iv = CryptoJS.lib.WordArray.random(8);
  return encrypt(CryptoJS.DES, key, iv, password);
}

export function encryptByAes(key, password) {
  const iv = CryptoJS.lib.WordArray.random(16);
  return encrypt(CryptoJS.AES, key, iv, password);
}
