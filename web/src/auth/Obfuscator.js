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

const characters = "123456789abcdef";

export function getRandomHexKey(length) {
  let key = "";
  for (let i = 0; i < length; i++) {
    const randomIndex = Math.floor(Math.random() * characters.length);
    key += characters[randomIndex];
  }
  return key;
}

export function getRandomKeyForObfuscator(obfuscatorType) {
  if (obfuscatorType === "DES") {
    return this.getRandomHexKey(16);
  } else if (obfuscatorType === "AES") {
    return this.getRandomHexKey(32);
  } else {
    return "";
  }
}

export function checkObfuscatorKey(obfuscatorType, obfuscatorKey) {
  if (obfuscatorType === "Plain" && obfuscatorKey !== "") {
    return [false, i18next.t("organization:The key should be empty")];
  } else if (obfuscatorType === "DES") {
    const regex = /^[1-9a-f]{16}$/;
    if (!regex.test(obfuscatorKey)) {
      return [false, i18next.t("organization:The input key doesn't match the DES regex") + " ^[1-9a-f]{16}$"];
    }
  } else if (obfuscatorType === "AES") {
    const regex = /^[1-9a-f]{32}$/;
    if (!regex.test(obfuscatorKey)) {
      return [false, i18next.t("organization:The input key doesn't match the AES regex") + " ^[1-9a-f]{32}$"];
    }
  }
  return [true, ""];
}

export function encryptByDes(key, password) {
  const iv = CryptoJS.lib.WordArray.random(8);
  const encrypted = CryptoJS.DES.encrypt(
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

export function encryptByAes(key, password) {
  const iv = CryptoJS.lib.WordArray.random(16);
  const encrypted = CryptoJS.AES.encrypt(
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
