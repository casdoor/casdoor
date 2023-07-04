import {Button, Space, notification} from "antd";
import i18next from "i18next";
import react from "react";
import {MfaRuleRequired} from "../../Setting";

const close = () => {
  // eslint-disable-next-line no-console
  console.log(
    "Notification was closed. Either the close button was clicked or duration time elapsed."
  );
};
const EnableMfaNotification = ({mfaItems, onupdate}) => {
  const [api, contextHolder] = notification.useNotification();

  react.useEffect(() => {
    // eslint-disable-next-line no-console
    console.log(mfaItems);
    if (mfaItems.some((item) => item.rule === MfaRuleRequired)) {
      openRequiredEnableNotification();
    } else {
      openPromptEnableNotification();
    }
  }, []);

  const openPromptEnableNotification = () => {
    const key = `open${Date.now()}`;
    const btn = (
      <Space>
        <Button type="link" size="small" onClick={() => api.destroy(key)}>
          {i18next.t("general:close")}
        </Button>
        <Button type="primary" size="small" onClick={() => {
          onupdate(false);
          api.destroy(key);
        }}
        >
          {i18next.t("general:later")}
        </Button>
      </Space>
    );
    api.open({
      message: "Notification Title",
      description:
        i18next.t("mfa: "),
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
      message: "Notification Title",
      description:
        i18next.t("mfa: "),
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
