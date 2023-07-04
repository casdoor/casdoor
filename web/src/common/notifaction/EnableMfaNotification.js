import {Button, Space, notification} from "antd";
import i18next from "i18next";

const close = () => {
  // eslint-disable-next-line no-console
  console.log(
    "Notification was closed. Either the close button was clicked or duration time elapsed."
  );
};
const EnableMfaNotification = ({onupdate}) => {
  const [api, contextHolder] = notification.useNotification();
  const openNotification = () => {
    const key = `open${Date.now()}`;
    const btn = (
      <Space>
        <Button type="link" size="small" onClick={() => api.destroy(key)}>
          {i18next.t("general:close")}
        </Button>
        <Button type="primary" size="small" onClick={() => {
          onupdate(false);
          api.destroy(key);
        }
        }>
          {i18next.t("general:later")}
        </Button>
      </Space>
    );
    api.open({
      message: "Notification Title",
      description:
        "A function will be be called after the notification is closed (automatically after the \"duration\" time of manually).",
      btn,
      key,
      onClose: close,
    });
  };
  return (
    <>
      {contextHolder}
      <Button type="primary" onClick={openNotification}>
        Open the notification box
      </Button>
    </>
  );
};

export default EnableMfaNotification;
