import * as UserWebauthnBackend from "@/backend/UserWebauthnBackend";
import * as Setting from "@/lib/setting";

/**
 * WebAuthn sign-in, ported from `signInWithWebAuthn` in web/src/auth/LoginPage.js.
 * The begin/finish endpoints, the base64url encoding and the response body are
 * unchanged, so the backend sees the same assertion it always did.
 */
export function signInWithWebAuthn(
  application: any,
  username: string,
  values: any,
  oAuthParams: any,
): Promise<{status: string; msg?: string; data?: any} | null> {
  const usernameParam = `&name=${encodeURIComponent(username)}`;

  return fetch(
    `${Setting.ServerUrl}/api/webauthn/signin/begin?owner=${application.organization}${username ? usernameParam : ""}`,
    {method: "GET", credentials: "include"},
  )
    .then((res) => res.json())
    .then((credentialRequestOptions: any) => {
      if ("status" in credentialRequestOptions) {
        return Promise.reject(new Error(credentialRequestOptions.msg));
      }
      credentialRequestOptions.publicKey.challenge =
        UserWebauthnBackend.webAuthnBufferDecode(credentialRequestOptions.publicKey.challenge);

      if (username) {
        credentialRequestOptions.publicKey.allowCredentials.forEach((item: any) => {
          item.id = UserWebauthnBackend.webAuthnBufferDecode(item.id);
        });
      }

      return navigator.credentials.get({publicKey: credentialRequestOptions.publicKey});
    })
    .then((assertion: any) => {
      const response = assertion.response;
      const resourceQuery = oAuthParams?.resource ? `&resource=${encodeURIComponent(oAuthParams.resource)}` : "";

      let finishUrl = `${Setting.ServerUrl}/api/webauthn/signin/finish?responseType=${values["type"]}`;
      if (values["type"] === "code") {
        finishUrl = `${Setting.ServerUrl}/api/webauthn/signin/finish?responseType=${values["type"]}&clientId=${oAuthParams.clientId}&scope=${oAuthParams.scope}&redirectUri=${oAuthParams.redirectUri}&nonce=${oAuthParams.nonce}&state=${oAuthParams.state}&codeChallenge=${oAuthParams.codeChallenge}&challengeMethod=${oAuthParams.challengeMethod}${resourceQuery}`;
      }

      return fetch(finishUrl, {
        method: "POST",
        credentials: "include",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({
          id: assertion.id,
          rawId: UserWebauthnBackend.webAuthnBufferEncode(assertion.rawId),
          type: assertion.type,
          response: {
            authenticatorData: UserWebauthnBackend.webAuthnBufferEncode(response.authenticatorData),
            clientDataJSON: UserWebauthnBackend.webAuthnBufferEncode(response.clientDataJSON),
            signature: UserWebauthnBackend.webAuthnBufferEncode(response.signature),
            userHandle: UserWebauthnBackend.webAuthnBufferEncode(response.userHandle),
          },
        }),
      }).then((res) => res.json());
    });
}
