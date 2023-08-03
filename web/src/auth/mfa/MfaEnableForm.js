import {Button} from "antd";
import i18next from "i18next";
import React, {useState} from "react";
import * as MfaBackend from "../../backend/MfaBackend";

export function MfaEnableForm({user, mfaType, recoveryCodes, onSuccess, onFail}) {
  const [loading, setLoading] = useState(false);
  const requestEnableMfa = () => {
    const data = {
      mfaType,
      ...user,
    };
    setLoading(true);
    MfaBackend.MfaSetupEnable(data).then(res => {
      if (res.status === "ok") {
        onSuccess(res);
      } else {
        onFail(res);
      }
    }
    ).finally(() => {
      setLoading(false);
    });
  };

  return (
    <div style={{width: "400px"}}>
      <p>{i18next.t("mfa:Please save this recovery code. Once your device cannot provide an authentication code, you can reset mfa authentication by this recovery code")}</p>
      <br />
      <code style={{fontStyle: "solid"}}>{recoveryCodes[0]}</code>
      <Button style={{marginTop: 24}} loading={loading} onClick={() => {
        requestEnableMfa();
      }} block type="primary">
        {i18next.t("general:Enable")}
      </Button>
    </div>
  );
}

export default MfaEnableForm;
