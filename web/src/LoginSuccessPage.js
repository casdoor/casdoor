import React from "react";
import i18next from "i18next";
import loginSuccessMp4 from "./static/success.mp4";
import zgsmLogo from "./static/zgsm-logo.png";

class LoginSuccessPage extends React.Component {
  renderLoginSuccessPage() {
    return (
      <div style={{position: "relative"}}>
        <div
          style={{
            position: "fixed",
            zIndex: 1000,
            color: "#F4F8FF",
            left: "50%",
            top: "340px",
            transform: "translate(-50%, 0)",
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
          }}
        >
          <img src={zgsmLogo} style={{alignSelf: "center"}}></img>
          <div style={{marginTop: "24px"}}>
            {i18next.t("login:Welcome to Shenma")}
          </div>
          <div style={{marginTop: "44px"}}>
            {i18next.t("login:Login success")}
          </div>
        </div>
        <video
          src={loginSuccessMp4}
          autoPlay
          preload="none"
          muted
          playsInline
          loop
          style={{
            objectFit: "cover",
            position: "fixed",
            top: 0,
            left: 0,
            minWidth: "100%",
            minHeight: "100%",
            zIndex: 1,
          }}
        />
      </div>
    );
  }

  render() {
    return this.renderLoginSuccessPage();
  }
}

export default LoginSuccessPage;
