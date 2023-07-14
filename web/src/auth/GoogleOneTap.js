import {useGoogleOneTapLogin} from "react-google-one-tap-login";
import * as AuthBackend from "./AuthBackend";
import * as Setting from "../Setting";

function GoogleOneTap(prop) {
  const application = prop.application;
  const googleProvider = prop.application.providers.find(providerItem => providerItem.provider.type === "Google");

  if (prop.preview !== "auto" && googleProvider !== undefined) {
    useGoogleOneTapLogin({
      googleAccountConfigs: {
        client_id: googleProvider.provider.clientId,
        callback: response => {
          const body = {
            type: "login",
            application: application.name,
            provider: googleProvider.name,
            code: response.credential,
            state: application.name,
            redirectUri: `${window.location.origin}/callback`,
            method: "signup",
            simple: true,
          };

          AuthBackend.login(body)
            .then((res) => {
              if (res.status === "ok") {
                Setting.showMessage("success", "Logged in successfully");

                const link = Setting.getFromLink();
                Setting.goToLink(link);
              }
            });
        },
      },
    });
  }

}

export default GoogleOneTap;
