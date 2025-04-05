/* eslint-disable */
// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React, {useEffect, useState} from "react";
import {Button, Card, ConfigProvider, Form, Image, Input, InputNumber, Radio, Select, Switch, theme} from "antd";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as LdapBackend from "./backend/LdapBackend";
import * as Setting from "./Setting";
import * as Conf from "./Conf";
import * as Obfuscator from "./auth/Obfuscator";
import i18next from "i18next";
import {EyeOutlined, LinkOutlined} from "@ant-design/icons";
import LdapTable from "./table/LdapTable";
import AccountTable from "./table/AccountTable";
import ThemeEditor from "./common/theme/ThemeEditor";
import MfaTable from "./table/MfaTable";
import {NavItemTree} from "./common/NavItemTree";
import {WidgetItemTree} from "./common/WidgetItemTree";
import {cloneDeep, isEmpty, isEqual} from "lodash-es";

const passwordTypeOptions = [
  "plain", "salt", "sha512-salt", "md5-salt", "bcrypt", "pbkdf2-salt", "argon2id"
].map(item => Setting.getOption(item, item))
const passwordComplexityOptions = [
  {value: "AtLeast6", label: i18next.t("user:The password must have at least 6 characters")},
  {value: "AtLeast8", label: i18next.t("user:The password must have at least 8 characters")},
  {value: "Aa123", label: i18next.t("user:The password must contain at least one uppercase letter, one lowercase letter and one digit")},
  {value: "SpecialChar", label: i18next.t("user:The password must contain at least one special character")},
  {value: "NoRepeat", label: i18next.t("user:The password must not contain any repeated characters")},
];
const passwordObfuscatorTypeOptions = ["Plain", "AES", "DES"].map(item => Setting.getOption(item, item));

const OrganizationEditPage = (props) => {
  const mode = props.location.mode ?? "edit";

  const [form] = Form.useForm();

  const [organizationName, setOrganizationName] = useState(props.match.params.organizationName);
  const [initOrg, setInitOrg] = useState(null);
  const [applications, setApplications] = useState([]);
  const [ldaps, setLdaps] = useState(null);

  useEffect(() => {
    getOrganization();
    getApplications();
    getLdaps();
  }, []);

  const getOrganization = () => {
    OrganizationBackend.getOrganization("admin", organizationName)
      .then((res) => {
        if (res.status === "ok") {
          const org = res.data;
          if (isEmpty(org)) {
            props.history.push("/404");
            return;
          }

          const passwordObfuscatorType = isEmpty(org.passwordObfuscatorType) ? "Plain" : org.passwordObfuscatorType
          setInitOrg({
            ...org,
            logo: isEmpty(org.logo) ? Setting.getLogo([""]) : org.logo,
            enableDarkLogo: !isEmpty(org.logoDark),
            passwordObfuscatorType: passwordObfuscatorType,
            languages: isEmpty(org.languages) ? [] : org.languages,
            defaultApplication: isEmpty(org.defaultApplication) ? null : org.defaultApplication,
            themeData: org?.themeData ?? {...Conf.ThemeDefault, isEnabled: false},
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  };

  const getApplications = () => {
    ApplicationBackend.getApplicationsByOrganization("admin", organizationName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        setApplications(res.data || []);
      });
  };

  const getLdaps = () => {
    LdapBackend.getLdaps(organizationName)
      .then(res => {
        let resData = [];
        if (res.status === "ok") {
          if (res.data !== null) {
            resData = res.data;
          }
        }
        setLdaps(resData);
      });
  };

  const labelWithTooltip = (label) => {
    return Setting.getLabel(i18next.t(label), i18next.t(`${label} - Tooltip`));
  };

  const ImagePreview = (props) => {
    return (
      <Image
        src={props.value}
        style={{height: "120px", ...props.style}}
        preview={{
          mask: (
            <><EyeOutlined/>&nbsp;{i18next.t("general:Preview")}</>
          ),
        }}
        fallback="data:image/png;base64,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"
      />
    );
  };

  const organizationRender = () => {
    return (
      <Card
        size="small"
        title={
          <div>
            {
              mode === "add"
                ? i18next.t("organization:New Organization")
                : i18next.t("organization:Edit Organization")
            }
            &nbsp;&nbsp;&nbsp;&nbsp;
            <Button onClick={() => submitOrganizationEdit(false)}>{i18next.t("general:Save")}</Button>
            <Button style={{marginLeft: "20px"}} type="primary" onClick={() => submitOrganizationEdit(true)}>
              {i18next.t("general:Save & Exit")}
            </Button>
            {
              mode === "add"
                ? (
                  <Button style={{marginLeft: "20px"}} onClick={() => deleteOrganization()}>
                    {i18next.t("general:Cancel")}
                  </Button>
                ) : null
            }
          </div>
        }
        style={Setting.isMobile() ? {margin: "5px"} : {}}
        type="inner"
      >
        <Form
          form={form}
          initialValues={initOrg}
          autoComplete="off"
          labelWrap
          labelAlign="left"
          labelCol={{sm: 4, md: 3, lg: 2}}
          wrapperCol={{sm: 20, md: 21, lg: 22}}
          onValuesChange={val => {
            console.log("form", val)
          }}
        >
          <Form.Item name="name" label={labelWithTooltip("general:Name")}>
            <Input disabled={organizationName === "built-in"} />
          </Form.Item>

          <Form.Item name="displayName" label={labelWithTooltip("general:Display name")}>
            <Input />
          </Form.Item>

          <Form.Item name="enableDarkLogo" label={labelWithTooltip("general:Enable dark logo")}>
            <Switch
              onChange={val => {
                form.setFieldValue("logoDark", val ? Setting.getLogo(["dark"]) : "")
              }}
            />
          </Form.Item>

          <Form.Item label={labelWithTooltip("general:Logo")}>
            <Form.Item name="logo">
              <Input prefix={<LinkOutlined />} />
            </Form.Item>
            <Form.Item name="logo" style={{marginBottom: 0}}>
              <ImagePreview />
            </Form.Item>
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.enableDarkLogo !== curr.enableDarkLogo}>
            {
              ({getFieldValue}) => (
                <Form.Item
                  label={labelWithTooltip("general:Logo dark")}
                  hidden={getFieldValue("enableDarkLogo") !== true}
                >
                  <Form.Item name="logoDark">
                    <Input prefix={<LinkOutlined />} />
                  </Form.Item>
                  <Form.Item name="logoDark" style={{marginBottom: 0}}>
                    <ImagePreview style={{backgroundColor: "#141414"}} />
                  </Form.Item>
                </Form.Item>
              )
            }
          </Form.Item>

          <Form.Item label={labelWithTooltip("general:Favicon")}>
            <Form.Item name="favicon">
              <Input prefix={<LinkOutlined />} />
            </Form.Item>
            <Form.Item name="favicon" style={{marginBottom: 0}}>
              <ImagePreview />
            </Form.Item>
          </Form.Item>

          <Form.Item name="websiteUrl" label={labelWithTooltip("organization:Website URL")}>
            <Input prefix={<LinkOutlined />} />
          </Form.Item>

          <Form.Item name="passwordType" label={labelWithTooltip("general:Password type")}>
            <Select options={passwordTypeOptions} />
          </Form.Item>

          <Form.Item name="passwordSalt" label={labelWithTooltip("general:Password salt")}>
            <Input />
          </Form.Item>

          <Form.Item name="passwordOptions" label={labelWithTooltip("general:Password complexity options")}>
            <Select mode="multiple" options={passwordComplexityOptions}/>
          </Form.Item>

          <Form.Item name="passwordObfuscatorType" label={labelWithTooltip("general:Password obfuscator")}>
            <Select
              options={passwordObfuscatorTypeOptions}
              onChange={value => form.setFieldValue("passwordObfuscatorKey", Obfuscator.getRandomKeyForObfuscator(value))}
            />
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.passwordObfuscatorType !== curr.passwordObfuscatorType}>
            {
              ({getFieldValue}) => (
                <Form.Item
                  name="passwordObfuscatorKey"
                  label={labelWithTooltip("general:Password obf key")}
                  hidden={["Plain", ""].includes(getFieldValue("passwordObfuscatorType"))}
                >
                  <Input />
                </Form.Item>
              )
            }
          </Form.Item>

          <Form.Item name="passwordExpireDays" label={labelWithTooltip("organization:Password expire days")}>
            <InputNumber />
          </Form.Item>

          <Form.Item name="countryCodes" label={labelWithTooltip("general:Supported country codes")}>
            <Select
              mode="multiple"
              filterOption={(input, option) => (option?.text ?? "").toLowerCase().includes(input.toLowerCase())}
            >
              {Setting.getCountryCodeOption({name: i18next.t("organization:All"), code: "All", phone: 0})}
              {Setting.getCountryCodeData().map((country) => Setting.getCountryCodeOption(country))}
            </Select>
          </Form.Item>

          <Form.Item name="languages" label={labelWithTooltip("general:Languages")}>
            <Select mode="multiple" options={Setting.Countries.map(item => Setting.getOption(item.label, item.key))}/>
          </Form.Item>

          <Form.Item label={labelWithTooltip("general:Default avatar")}>
            <Form.Item name="defaultAvatar">
              <Input prefix={<LinkOutlined />} />
            </Form.Item>
            <Form.Item name="defaultAvatar" style={{marginBottom: 0}}>
              <ImagePreview />
            </Form.Item>
          </Form.Item>

          <Form.Item name="defaultApplication" label={labelWithTooltip("general:Default application")}>
            <Select
              options={applications?.map((item) => Setting.getOption(Setting.getApplicationDisplayName(item.name), item.name))}
            />
          </Form.Item>

          <Form.Item name="userTypes" label={labelWithTooltip("organization:User types")}>
            <Select
              mode="tags"
              options={initOrg.userTypes?.map(item => Setting.getOption(item, item))}
            />
          </Form.Item>

          <Form.Item name="tags" label={labelWithTooltip("organization:Tags")}>
            <Select
              mode="tags"
              options={initOrg.tags?.map(item => Setting.getOption(item, item))}
            />
          </Form.Item>

          <Form.Item name="masterPassword" label={labelWithTooltip("general:Master password")}>
            <Input />
          </Form.Item>

          <Form.Item name="defaultPassword" label={labelWithTooltip("general:Default password")}>
            <Input />
          </Form.Item>

          <Form.Item name="masterVerificationCode" label={labelWithTooltip("general:Master verification code")}>
            <Input />
          </Form.Item>

          <Form.Item name="ipWhitelist" label={labelWithTooltip("general:IP whitelist")}>
            <Input />
          </Form.Item>

          <Form.Item name="initScore" label={labelWithTooltip("organization:Init score")}>
            <InputNumber />
          </Form.Item>

          <Form.Item name="enableSoftDeletion" label={labelWithTooltip("organization:Soft deletion")}>
            <Switch />
          </Form.Item>

          <Form.Item name="isProfilePublic" label={labelWithTooltip("organization:Is profile public")}>
            <Switch />
          </Form.Item>

          <Form.Item name="useEmailAsUsername" label={labelWithTooltip("organization:Use Email as username")}>
            <Switch />
          </Form.Item>

          <Form.Item name="enableTour" label={labelWithTooltip("general:Enable tour")}>
            <Switch />
          </Form.Item>

          <Form.Item
            name="navItems"
            label={labelWithTooltip("organization:Navbar items")}
            getValueProps={value => ({checkedKeys: value ?? ["all"]})}
            normalize={cloneDeep}
            trigger="onCheck"
          >
            <NavItemTree
              disabled={!Setting.isAdminUser(props.account)}
              defaultExpandedKeys={["all"]}
            />
          </Form.Item>

          <Form.Item
            name="widgetItems"
            label={labelWithTooltip("organization:Widget items")}
            getValueProps={value => ({checkedKeys: value ?? ["all"]})}
            normalize={cloneDeep}
            trigger="onCheck"
          >
            <WidgetItemTree
              disabled={!Setting.isAdminUser(props.account)}
              defaultExpandedKeys={["all"]}
            />
          </Form.Item>

          <Form.Item
            name="accountItems"
            label={labelWithTooltip("organization:Account items")}
            getValueProps={value => ({table: value ?? []})}
            normalize={cloneDeep}
            trigger="onUpdateTable"
          >
            <AccountTable title={i18next.t("organization:Account items")}/>
          </Form.Item>

          <Form.Item
            name="mfaItems"
            label={labelWithTooltip("general:MFA items")}
            getValueProps={value => ({table: value ?? []})}
            normalize={cloneDeep}
            trigger="onUpdateTable"
          >
            <MfaTable title={i18next.t("general:MFA items")}/>
          </Form.Item>

          <Form.Item label={labelWithTooltip("theme:Theme")}>
            <Form.Item name={["themeData", "isEnabled"]}>
              <Radio.Group buttonStyle="solid">
                <Radio.Button value={false}>{i18next.t("organization:Follow global theme")}</Radio.Button>
                <Radio.Button value={true}>{i18next.t("theme:Customize theme")}</Radio.Button>
              </Radio.Group>
            </Form.Item>
            <Form.Item noStyle shouldUpdate={(prev, curr) => !isEqual(prev.themeData, curr.themeData)}>
              {
                ({getFieldValue, setFieldValue}) => (
                  <Form.Item
                    name="themeData"
                    label={labelWithTooltip("theme:Theme")}
                    hidden={getFieldValue("themeData")?.isEnabled !== true}
                  >
                    <ThemeEditor
                      themeData={getFieldValue("themeData")}
                      onThemeChange={(_, nextThemeData) => {
                        const {isEnabled} = getFieldValue("themeData") ?? {...Conf.ThemeDefault, isEnabled: false};
                        setFieldValue("themeData", {...nextThemeData, isEnabled});
                      }}
                    />
                  </Form.Item>
                )
              }
            </Form.Item>
          </Form.Item>

          <Form.Item label={labelWithTooltip("general:LDAPs")}>
            <LdapTable
              title={i18next.t("general:LDAPs")}
              table={ldaps}
              organizationName={organizationName}
              onUpdateTable={(value) => setLdaps(value)}
            />
          </Form.Item>

        </Form>
      </Card>
    );
  };

  const submitOrganizationEdit = (exitAfterSave) => {
    form.setFieldValue("accountItems", form.getFieldValue("accountItems")?.filter(accountItem => accountItem.name !== "Please select an account item"))

    const orgOwner = initOrg.owner;
    const orgName = form.getFieldValue("name");
    const passwordObfuscatorErrorMessage = Obfuscator.checkPasswordObfuscator(form.getFieldValue("passwordObfuscatorType"), form.getFieldValue("passwordObfuscatorKey"));
    if (passwordObfuscatorErrorMessage.length > 0) {
      Setting.showMessage("error", passwordObfuscatorErrorMessage);
      return;
    }

    console.log("update", form.getFieldsValue())

    OrganizationBackend.updateOrganization(orgOwner, organizationName, {...initOrg, ...form.getFieldsValue()})
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));

          if (props.account.organization.name === organizationName) {
            props.onChangeTheme(Setting.getThemeData(form.getFieldsValue()));
          }

          setOrganizationName(orgName);
          window.dispatchEvent(new Event("storageOrganizationsChanged"));

          if (exitAfterSave) {
            props.history.push("/organizations");
          } else {
            props.history.push(`/organizations/${orgName}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          form.setFieldValue("name", initOrg.name)
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  const deleteOrganization = () => {
    OrganizationBackend.deleteOrganization(form.getFieldsValue())
      .then((res) => {
        if (res.status === "ok") {
          props.history.push("/organizations");
          window.dispatchEvent(new Event("storageOrganizationsChanged"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  return (
    <div>
      {
        initOrg !== null ? organizationRender() : null
      }
      <div style={{marginTop: "20px", marginLeft: "40px"}}>
        <Button size="large" onClick={() => submitOrganizationEdit(false)}>{i18next.t("general:Save")}</Button>
        <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => submitOrganizationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        {mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => deleteOrganization()}>{i18next.t("general:Cancel")}</Button> : null}
      </div>
    </div>
  );
};

export default OrganizationEditPage;
