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
import {Buffer} from "buffer";

export function getRandomKeyForObfuscator(obfuscatorType) {
  if (obfuscatorType === "DES") {
    return getRandomHexKey(16);
  } else if (obfuscatorType === "AES") {
    return getRandomHexKey(32);
  } else {
    return "";
  }
}

export const passwordObfuscatorKeyRegexes = {
  "DES": /^[1-9a-f]{16}$/,
  "AES": /^[1-9a-f]{32}$/,
};

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

export function checkPasswordObfuscator(passwordObfuscatorType, passwordObfuscatorKey) {
  if (passwordObfuscatorType === undefined) {
    return "passwordObfuscatorType should not be undefined";
  } else if (passwordObfuscatorType === "Plain" || passwordObfuscatorType === "") {
    return "";
  } else if (passwordObfuscatorType === "AES" || passwordObfuscatorType === "DES") {
    if (passwordObfuscatorKeyRegexes[passwordObfuscatorType].test(passwordObfuscatorKey)) {
      return "";
    } else {
      return `The password obfuscator key doesn't match the regex: ${passwordObfuscatorKeyRegexes[passwordObfuscatorType].source}`;
    }
  } else {
    return `unsupported password obfuscator type: ${passwordObfuscatorType}`;
  }
}

export function encryptByPasswordObfuscator(passwordObfuscatorType, passwordObfuscatorKey, password) {
  const passwordObfuscatorErrorMessage = checkPasswordObfuscator(passwordObfuscatorType, passwordObfuscatorKey);
  if (passwordObfuscatorErrorMessage.length > 0) {
    return ["", passwordObfuscatorErrorMessage];
  } else {
    if (passwordObfuscatorType === "Plain" || passwordObfuscatorType === "") {
      return [password, ""];
    } else if (passwordObfuscatorType === "AES") {
      return [encryptByAes(passwordObfuscatorKey, password), ""];
    } else if (passwordObfuscatorType === "DES") {
      return [encryptByDes(passwordObfuscatorKey, password), ""];
    }
  }
}

function encryptByDes(key, password) {
  const iv = CryptoJS.lib.WordArray.random(8);
  return encrypt(CryptoJS.DES, key, iv, password);
}

function encryptByAes(key, password) {
  const iv = CryptoJS.lib.WordArray.random(16);
  return encrypt(CryptoJS.AES, key, iv, password);
}

function getRandomHexKey(length) {
  const characters = "123456789abcdef";
  let key = "";
  for (let i = 0; i < length; i++) {
    const randomIndex = Math.floor(Math.random() * characters.length);
    key += characters[randomIndex];
  }
  return key;
}
