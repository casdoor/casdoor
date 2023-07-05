import {Button, Space, Tag, notification} from "antd";
import i18next from "i18next";
import {useEffect} from "react";
import {useHistory, useLocation} from "react-router-dom";
import * as Setting from "../../Setting";
import {MfaRulePrompted, MfaRuleRequired} from "../../Setting";

const EnableMfaNotification = ({account, onupdate}) => {
  const [api, contextHolder] = notification.useNotification();
  const history = useHistory();
  const location = useLocation();

  useEffect(() => {
    if (account === null) {
      return;
    }

    const mfaItems = Setting.getMfaItemsByRules(account, account?.organization, [MfaRuleRequired, MfaRulePrompted]);
    if (location.state?.from === "/login" && mfaItems.length !== 0) {
      if (mfaItems.some((item) => item.rule === MfaRuleRequired)) {
        // eslint-disable-next-line no-console
        console.log(mfaItems);
        openRequiredEnableNotification(mfaItems.find((item) => item.rule === MfaRuleRequired).name);
      } else {
        openPromptEnableNotification(mfaItems.filter((item) => item.rule === MfaRulePrompted)?.map((item) => item.name));
      }
    }
  }, [account]);

  const openPromptEnableNotification = (mfaTypes) => {
    const key = `open${Date.now()}`;
    const btn = (
      <Space>
        <Button type="link" size="small" onClick={() => api.destroy(key)}>
          {i18next.t("general:Later")}
        </Button>
        <Button type="primary" size="small" onClick={() => {
          history.push(`/mfa/setup?mfaType=${mfaTypes[0]}`, {from: "notification"});
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
      <Space direction={"vertical"}>
        {i18next.t("mfa:To ensure the security of your account, it is recommended that you enable multi-factor authentication")}
        <Space>{mfaTypes.map((item) => <Tag color="orange" key={item}>{item}</Tag>)}</Space>
      </Space>,
      btn,
      key,
    });
  };

  const openRequiredEnableNotification = (mfaType) => {
    const key = `open${Date.now()}`;
    const btn = (
      <Space>
        <Button type="primary" size="small" onClick={() => {
          api.destroy(key);
        }}
        >
          {i18next.t("general:Confirm")}
        </Button>
      </Space>
    );
    api.open({
      message: i18next.t("mfa:Enable multi-factor authentication"),
      description:
      <Space direction={"vertical"}>
        {i18next.t("mfa:To ensure the security of your account, it is required to enable multi-factor authentication")}
        <Space><Tag color="orange">{mfaType}</Tag></Space>
      </Space>,
      btn,
      key,
    });
  };

  return (
    <>
      {contextHolder}
    </>
  );
};

export default EnableMfaNotification;
