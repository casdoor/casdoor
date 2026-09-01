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

import {createButton} from "react-social-login-buttons";
import {StaticBaseUrl} from "../Setting";
import {useGoogleOneTapLogin} from "react-google-one-tap-login";
import * as Setting from "../Setting";
import * as Provider from "./Provider";

function Icon({width = 24, height = 24, color}) {
  return <img src={`${StaticBaseUrl}/buttons/google.svg`} alt="Sign in with Google" />;
}

const config = {
  text: "Sign in with Google",
  icon: Icon,
  iconFormat: name => `fa fa-${name}`,
  style: {background: "#ffffff", color: "#000000"},
  activeStyle: {background: "#eff0ee"},
};

const GoogleLoginButton = createButton(config);

export function GoogleOneTapLoginVirtualButton(prop) {
  const application = prop.application;
  const providerConf = prop.providerConf;
  // https://stackoverflow.com/questions/62281579/google-one-tap-sign-in-ui-not-displayed-after-clicking-the-close-button
  // document.cookie = "g_state=;path=/;expires=Thu, 01 Jan 1970 00:00:01 GMT";
  useGoogleOneTapLogin({
    googleAccountConfigs: {
      client_id: providerConf.provider.clientId,
    },
    onError: (error) => {
      Setting.showMessage("error", error);
    },
    onSuccess: (response) => {
      const code = "GoogleIdToken-" + JSON.stringify(response);
      const authUrlParams = new URLSearchParams(Provider.getAuthUrl(application, providerConf.provider, "signup"));
      const state = authUrlParams.get("state");
      let redirectUri = authUrlParams.get("redirect_uri");
      redirectUri = `${redirectUri}?state=${state}&code=${encodeURIComponent(code)}`;
      Setting.goToLink(redirectUri);
    },
    disableCancelOnUnmount: false,
  });

}

export default GoogleLoginButton;
