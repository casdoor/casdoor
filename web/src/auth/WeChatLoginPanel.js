import React, {useState} from "react";
import QRCode from "qrcode.react";
import i18next from "i18next";
import * as Provider from "./Provider";

export default function WeChatLoginPanel({
  application,
  renderFormItem,
  loginWidth,
}) {
  const [wechatQrRefreshKey, setWechatQrRefreshKey] = useState(Date.now());

  const wechatProvider = application?.providers?.find(p => p.provider?.type === "WeChat");
  if (!wechatProvider) {
    return <div style={{textAlign: "center", color: "red"}}>{i18next.t("login:Please configure WeChat login")}</div>;
  }

  const qrUrl = Provider.getAuthUrl(application, wechatProvider.provider, "login") + `&refreshKey=${wechatQrRefreshKey}`;

  return (
    <div style={{width: `${loginWidth}px`, margin: "0 auto"}}>
      {application.signinItems?.filter(item => item.name === "Logo").map(item => renderFormItem(application, item))}
      {application.signinItems?.filter(item => item.name === "Signin methods").map(item => renderFormItem(application, item))}
      <div style={{textAlign: "center", marginTop: 16, marginBottom: 32}}>
        <QRCode value={qrUrl} size={200} />
        <div style={{marginTop: 12, color: "#888", fontSize: 15, display: "flex", alignItems: "center", justifyContent: "center", gap: 8}}>
          <span>{i18next.t("login:Scan with WeChat to login")}</span>
          <a style={{cursor: "pointer", color: "#1890ff", fontSize: 14}} onClick={() => setWechatQrRefreshKey(Date.now())}>
            {i18next.t("login:Refresh")}
          </a>
        </div>
      </div>
      {application.signinItems?.map(signinItem => {
        if (["Logo", "Username", "Password", "Forgot password?", "Login button", "Signin methods", "Signup link", "Providers"].includes(signinItem.name)) {
          return null;
        }
        return renderFormItem(application, signinItem);
      })}
    </div>
  );
}
