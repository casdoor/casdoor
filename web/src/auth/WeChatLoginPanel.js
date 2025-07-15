import React from "react";
import * as AuthBackend from "./AuthBackend";
import i18next from "i18next";

class WeChatLoginPanel extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      qrCode: null,
    };
  }

  componentDidMount() {
    this.fetchQrCode();
  }

  componentDidUpdate(prevProps) {
    if (this.props.loginMethod === "wechat" && prevProps.loginMethod !== "wechat") {
      this.fetchQrCode();
    }
    if (prevProps.loginMethod === "wechat" && this.props.loginMethod !== "wechat") {
      this.setState({qrCode: null});
    }
  }

  fetchQrCode() {
    const {application} = this.props;
    const wechatProviderItem = application?.providers?.find(p => p.provider?.type === "WeChat");
    if (wechatProviderItem) {
      AuthBackend.getWechatQRCode(`${wechatProviderItem.provider.owner}/${wechatProviderItem.provider.name}`).then(res => {
        if (res.status === "ok" && res.data) {
          this.setState({qrCode: res.data});
        } else {
          this.setState({qrCode: null});
        }
      });
    }
  }

  render() {
    const {application, loginWidth = 320} = this.props;
    return (
      <div style={{width: loginWidth, margin: "0 auto", textAlign: "center", marginTop: 16}}>
        {application.signinItems?.filter(item => item.name === "Logo").map(signinItem => this.props.renderFormItem(application, signinItem))}
        {this.props.renderMethodChoiceBox()}
        {application.signinItems?.filter(item => item.name === "Languages").map(signinItem => this.props.renderFormItem(application, signinItem))}
        {this.state.qrCode ? (
          <div style={{marginTop: 2}}>
            <img src={`data:image/png;base64,${this.state.qrCode}`} alt="WeChat QR code" style={{width: 250, height: 250}} />
            <div style={{marginTop: 8}}>
              <div>{i18next.t("login:Please scan the QR code with WeChat to login")}</div>
              <a onClick={e => {e.preventDefault(); this.fetchQrCode();}}>
                {i18next.t("login:Refresh")}
              </a>
            </div>
          </div>
        ) : null}
      </div>
    );
  }
}

export default WeChatLoginPanel;
