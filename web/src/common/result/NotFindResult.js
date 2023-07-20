import {Button, Result} from "antd";
import i18next from "i18next";
import React, {useEffect} from "react";

export const NotFindResult = ({onUpdateApplication}) => {
  useEffect(() => {
    if (onUpdateApplication !== undefined) {
      onUpdateApplication(null);
    }
  }, [onUpdateApplication]);

  return (
    <Result
      style={{margin: "0 auto"}}
      status="404"
      title="404 NOT FOUND"
      subTitle={i18next.t("general:Sorry, the page you visited does not exist.")}
      extra={<Button type="primary" href={"/"}>{i18next.t("general:Back Home")}</Button>}
    />
  );
};

export default NotFindResult;
