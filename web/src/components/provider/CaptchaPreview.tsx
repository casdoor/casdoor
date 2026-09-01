import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {CaptchaModal} from "@/components/common/CaptchaModal";
import * as UserBackend from "@/backend/UserBackend";

interface CaptchaPreviewProps {
  owner: string;
  name: string;
  provider: any;
  captchaType: string;
  subType?: string;
  clientId?: string;
  clientSecret?: string;
  clientId2?: string;
  clientSecret2?: string;
  providerUrl?: string;
}

/**
 * "Preview" button of a Captcha provider — opens the real captcha challenge and
 * verifies the token against /api/verify-captcha. Ported from
 * web/src/common/CaptchaPreview.js.
 */
export function CaptchaPreview({
  owner,
  name,
  provider,
  captchaType,
  subType,
  clientId,
  clientSecret,
  clientId2,
  clientSecret2,
  providerUrl,
}: CaptchaPreviewProps) {
  const [visible, setVisible] = React.useState(false);

  const clickPreview = () => {
    provider.name = name;
    provider.clientId = clientId;
    provider.type = captchaType;
    provider.providerUrl = providerUrl;
    if (clientSecret !== "***") {
      provider.clientSecret = clientSecret;
    }
    setVisible(true);
  };

  const isButtonDisabled = () => {
    if (captchaType !== "Default") {
      if (!clientId || !clientSecret) {
        return true;
      }
      if (captchaType === "Aliyun Captcha") {
        if (!subType || !clientId2 || !clientSecret2) {
          return true;
        }
      }
    }
    return false;
  };

  const onOk = (type: string, captchaToken: string, secret: string) => {
    UserBackend.verifyCaptcha(owner, name, type, captchaToken, secret).then(() => {
      setVisible(false);
    });
  };

  return (
    <React.Fragment>
      <Button onClick={clickPreview} disabled={isButtonDisabled()}>
        {i18next.t("general:Preview")}
      </Button>
      <CaptchaModal
        owner={owner}
        name={name}
        visible={visible}
        isCurrentProvider={true}
        onOk={onOk}
        onCancel={() => setVisible(false)}
      />
    </React.Fragment>
  );
}
