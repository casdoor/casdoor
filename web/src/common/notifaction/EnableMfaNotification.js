import {Button, Space, notification} from "antd";
import i18next from "i18next";
import {useEffect} from "react";
import {useHistory, useLocation} from "react-router-dom";
import * as Setting from "../../Setting";
import {MfaRulePrompted, MfaRuleRequired} from "../../Setting";

const close = () => {
  // eslint-disable-next-line no-console
  console.log(
    "Notification was closed. Either the close button was clicked or duration time elapsed."
  );
};
const EnableMfaNotification = ({account, onupdate}) => {
  const [api, contextHolder] = notification.useNotification();
  const history = useHistory();
  const location = useLocation();

  useEffect(() => {
    if (account === null) {
      return;
    }

    // eslint-disable-next-line no-console
    console.log(location);
    const mfaItems = Setting.getMfaItemsByRules(account, account?.organization, [MfaRuleRequired, MfaRulePrompted]);
    if (location.state?.from === "/login" && mfaItems.length !== 0) {
      if (mfaItems.some((item) => item.rule === MfaRuleRequired)) {
        openRequiredEnableNotification();
      } else {
        openPromptEnableNotification();
      }
    }
  }, [account]);

  const openPromptEnableNotification = () => {
    const key = `open${Date.now()}`;
    const btn = (
      <Space>
        <Button type="link" size="small" onClick={() => api.destroy(key)}>
          {i18next.t("general:later")}
        </Button>
        <Button type="primary" size="small" onClick={() => {
          history.push("/mfa/setup", {from: "notification"});
          api.destroy(key);
        }}
        >
          {i18next.t("general:Go to enable")}
        </Button>
      </Space>
    );
    api.open({
      message: i18next.t("mfa:Enable multi-factor authentication"),
      description:
        i18next.t("mfa:To ensure the security of your account, the organization recommends you to enable multi-factor authentication."),
      btn,
      key,
      onClose: close,
    });
  };

  const openRequiredEnableNotification = () => {
    const key = `open${Date.now()}`;
    const btn = (
      <Space>
        <Button type="primary" size="small" onClick={() => {
          api.destroy(key);
        }}
        >
          {i18next.t("general:close")}
        </Button>
      </Space>
    );
    api.open({
      message: i18next.t("mfa:Enable multi-factor authentication"),
      description:
        i18next.t("mfa:To ensure the security of your account, the organization requires you to enable multi-factor authentication."),
      btn,
      key,
      onClose: close,
    });
  };

  return (
    <>
      {contextHolder}
    </>
  );
};

export default EnableMfaNotification;
