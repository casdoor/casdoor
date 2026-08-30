import * as React from "react";
import {useSearchParams} from "react-router-dom";
import {CaptchaModal} from "@/components/common/CaptchaModal";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as Setting from "@/lib/setting";

/**
 * Standalone captcha challenge at /captcha, used when another site delegates its
 * captcha to Casdoor. Ported from web/src/CaptchaPage.js: the solved token is
 * handed back on the redirect URI as query parameters.
 */
export default function CaptchaPage() {
  const [search] = useSearchParams();
  const applicationName = search.get("state");
  const redirectUri = search.get("redirect_uri");

  const [application, setApplication] = React.useState<any>(null);

  React.useEffect(() => {
    if (applicationName === null) {
      return;
    }
    ApplicationBackend.getApplication("admin", applicationName).then((res: any) => {
      setApplication(res.status === "error" ? null : res.data);
    });
  }, [applicationName]);

  const captchaItems = (application?.providers ?? []).filter(
    (item: any) => item.provider?.category === "Captcha",
  );
  const always = captchaItems.filter((item: any) => item.rule === "Always");
  const dynamic = captchaItems.filter((item: any) => item.rule === "Dynamic");
  const provider = always.length > 0 ? always[0].provider : dynamic[0]?.provider;

  if (!provider) {
    return null;
  }

  const callback = (values: Record<string, string>) => {
    Setting.goToLink(`${redirectUri}?code=${values.captchaToken}&type=${values.captchaType}&secret=${values.clientSecret}&applicationId=${values.applicationId ?? ""}`);
  };

  return (
    <CaptchaModal
      owner={provider.owner}
      name={provider.name}
      visible={true}
      isCurrentProvider={true}
      onOk={(captchaType, captchaToken, clientSecret) =>
        callback({captchaType, captchaToken, clientSecret, applicationId: `${provider.owner}/${provider.name}`})
      }
      onCancel={() => callback({captchaType: "none", captchaToken: "", clientSecret: ""})}
    />
  );
}
