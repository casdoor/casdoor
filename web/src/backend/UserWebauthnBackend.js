// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

import * as Setting from "../Setting";

export function registerWebauthnCredential() {
  return fetch(`${Setting.ServerUrl}/api/webauthn/signup/begin`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  })
    .then(res => res.json())
    .then((credentialCreationOptions) => {
      credentialCreationOptions.publicKey.challenge = webAuthnBufferDecode(credentialCreationOptions.publicKey.challenge);
      credentialCreationOptions.publicKey.user.id = webAuthnBufferDecode(credentialCreationOptions.publicKey.user.id);
      if (credentialCreationOptions.publicKey.excludeCredentials) {
        for (let i = 0; i < credentialCreationOptions.publicKey.excludeCredentials.length; i++) {
          credentialCreationOptions.publicKey.excludeCredentials[i].id = webAuthnBufferDecode(credentialCreationOptions.publicKey.excludeCredentials[i].id);
        }
      }

      return navigator.credentials.create({
        publicKey: credentialCreationOptions.publicKey,
      });
    })
    .then((credential) => {
      const attestationObject = credential.response.attestationObject;
      const clientDataJSON = credential.response.clientDataJSON;
      const rawId = credential.rawId;
      return fetch(`${Setting.ServerUrl}/api/webauthn/signup/finish`, {
        method: "POST",
        credentials: "include",
        body: JSON.stringify({
          id: credential.id,
          rawId: webAuthnBufferEncode(rawId),
          type: credential.type,
          response: {
            attestationObject: webAuthnBufferEncode(attestationObject),
            clientDataJSON: webAuthnBufferEncode(clientDataJSON),
          },
        }),
      })
        .then(res => res.json());
    });
}

export function deleteUserWebAuthnCredential(credentialID) {
  const form = new FormData();
  form.append("credentialID", credentialID);

  return fetch(`${Setting.ServerUrl}/api/webauthn/delete-credential`, {
    method: "POST",
    credentials: "include",
    body: form,
    dataType: "text",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

// Base64URL to ArrayBuffer
export function webAuthnBufferDecode(value) {
  value = value.replace(/-/g, "+").replace(/_/g, "/");
  while (value.length % 4) {
    value += "=";
  }
  return Uint8Array.from(atob(value), c => c.charCodeAt(0));
}

// ArrayBuffer to URLBase64
export function webAuthnBufferEncode(value) {
  return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}
