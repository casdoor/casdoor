import {Button, Result} from "antd";
import i18next from "i18next";

export const UnauthorizedResult = () => {
  return (
    <Result status="403"
      title="403 Unauthorized"
      subTitle={i18next.t("general:Sorry, you do not have permission to access this page or logged in status invalid.")}
      extra={<Button type="primary" href={"/"}>{i18next.t("general:Back Home")}</Button>}
    />
  );
};

export default UnauthorizedResult;
