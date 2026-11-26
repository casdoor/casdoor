// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

import * as Provider from "@/auth/Provider";
import * as Setting from "@/lib/setting";

/**
 * An identity provider plugin signs in on the sign-in page itself, such as with a crypto wallet, when it has no
 * authorization page (Auth URL) and ships a sign-in script instead. The script is downloaded from this Casdoor only
 * when the user clicks the provider, so nothing of it is in the bundle.
 */
export function hasPluginScript(provider: any) {
  return provider?.type === "Plugin" && provider?.category === "OAuth" && !provider?.customAuthUrl;
}

function getQuery(application: any, provider: any) {
  return new URLSearchParams({application: `${application.owner}/${application.name}`, provider: provider.name}).toString();
}

/** Runs the sign-in script of the plugin of the provider, see sdk.ChallengeIdentityProvider of casdoor-integrations. */
export async function runPluginSignIn(application: any, provider: any, method: string) {
  const query = getQuery(application, provider);
  let module: any;
  try {
    module = await import(/* @vite-ignore */ `${Setting.ServerUrl}/api/get-plugin-script?${query}`);
  } catch {
    // the API answers with a JSON error, such as a missing license, which a module import can't read
    const res = await fetch(`${Setting.ServerUrl}/api/get-plugin-script?${query}`, {credentials: "include", headers: {"Accept-Language": Setting.getAcceptLanguage()}})
      .then((r) => r.json())
      .catch(() => null);
    Setting.showMessage("error", res?.msg ?? `Failed to load the sign-in script of ${provider.name}`);
    return;
  }

  const casdoor = {
    application,
    provider,
    method,
    language: Setting.getLanguage(),
    getChallenge: async (input: Record<string, string>) => {
      const res = await fetch(`${Setting.ServerUrl}/api/get-plugin-challenge?${query}`, {
        method: "POST",
        credentials: "include",
        headers: {"Accept-Language": Setting.getAcceptLanguage(), "Content-Type": "application/json"},
        body: JSON.stringify(input ?? {}),
      }).then((r) => r.json());
      if (res.status !== "ok") {
        throw new Error(res.msg);
      }
      return res.data as string;
    },
    finish: (code: string) => {
      Setting.setSigninLanguage(Setting.getLanguage());
      Setting.goToLink(`${Provider.getAuthUrl(application, provider, method)}&code=${encodeURIComponent(code)}`);
    },
    showMessage: (type: "success" | "error" | "info" | "warning", text: string) => Setting.showMessage(type, text),
  };
  await module.default(casdoor);
}
