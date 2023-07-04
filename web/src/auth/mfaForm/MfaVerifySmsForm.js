import {UserOutlined} from "@ant-design/icons";
import {Button, Form, Input} from "antd";
import i18next from "i18next";
import React, {useEffect} from "react";
import {CountryCodeSelect} from "../../common/select/CountryCodeSelect";
import {SendCodeInput} from "../../common/SendCodeInput";
import * as Setting from "../../Setting";
import {EmailMfaType, SmsMfaType} from "../MfaSetupPage";
import {mfaAuth} from "./MfaVerifyForm";

export const MfaVerifySmsForm = ({mfaProps, application, onFinish, method, user}) => {
  const [dest, setDest] = React.useState("");
  const [form] = Form.useForm();

  useEffect(() => {
    if (method === mfaAuth) {
      setDest(mfaProps.secret);
      return;
    }
    if (mfaProps.mfaType === SmsMfaType) {
      setDest(user.phone);
      return;
    }

    if (mfaProps.mfaType === EmailMfaType) {
      setDest(user.email);
    }
  }, [mfaProps.mfaType]);

  const isShowText = () => {
    // eslint-disable-next-line no-console
    console.log(mfaProps.secret);
    if (method === mfaAuth) {
      return true;
    }
    if (mfaProps.mfaType === SmsMfaType && user.phone !== "") {
      return true;
    }
    if (mfaProps.mfaType === EmailMfaType && user.email !== "") {
      return true;
    }
    return false;
  };

  const isEmail = () => {
    return mfaProps.mfaType === EmailMfaType;
  };

  return (
    <Form
      form={form}
      style={{width: "300px"}}
      onFinish={onFinish}
      initialValues={{
        countryCode: mfaProps.countryCode,
      }}
    >
      {isShowText() ?
        <div style={{marginBottom: 20, textAlign: "left", gap: 8}}>
          {isEmail() ? i18next.t("mfaForm:Your email is") : i18next.t("mfaForm:Your phone is")} {dest}
        </div> :
        (<React.Fragment>
          <p>{isEmail() ? i18next.t("mfaForm:Please bind your email first, the system will automatically uses the mail for multi-factor authentication") :
            i18next.t("mfaForm:Please bind your phone first, the system automatically uses the phone for multi-factor authentication")}
          </p>
          <Input.Group compact style={{width: "300Px", marginBottom: "30px"}}>
            {isEmail() ? null :
              <Form.Item
                name="countryCode"
                noStyle
                rules={[
                  {
                    required: false,
                    message: i18next.t("signup:Please select your country code!"),
                  },
                ]}
              >
                <CountryCodeSelect
                  initValue={mfaProps.countryCode}
                  style={{width: "30%"}}
                  countryCodes={application.organizationObj.countryCodes}
                />
              </Form.Item>
            }
            <Form.Item
              name="dest"
              noStyle
              rules={[{required: true, message: i18next.t("login:Please input your Email or Phone!")}]}
            >
              <Input
                style={{width: isEmail() ? "100% " : "70%"}}
                onChange={(e) => {setDest(e.target.value);}}
                prefix={<UserOutlined />}
                placeholder={isEmail() ? i18next.t("general:Email") : i18next.t("general:Phone")}
              />
            </Form.Item>
          </Input.Group>
        </React.Fragment>
        )
      }
      <Form.Item
        name="passcode"
        rules={[{required: true, message: i18next.t("login:Please input your code!")}]}
      >
        <SendCodeInput
          countryCode={form.getFieldValue("countryCode")}
          method={method}
          onButtonClickArgs={[mfaProps.secret || dest, isEmail() ? "email" : "phone", Setting.getApplicationName(application)]}
          application={application}
        />
      </Form.Item>
      <Form.Item>
        <Button
          style={{marginTop: 24}}
          loading={false}
          block
          type="primary"
          htmlType="submit"
        >
          {i18next.t("forget:Next Step")}
        </Button>
      </Form.Item>
    </Form>
  );
};

export default MfaVerifySmsForm;
