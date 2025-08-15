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

import React from "react";
import {Button, Card, Col, Form, Input, InputNumber, List, Result, Row, Select, Space, Spin, Switch, Tag, Tooltip} from "antd";
import {withRouter} from "react-router-dom";
import {TotpMfaType} from "./auth/MfaSetupPage";
import * as GroupBackend from "./backend/GroupBackend";
import * as UserBackend from "./backend/UserBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import EnableMfaModal from "./common/modal/EnableMfaModal";
import * as Setting from "./Setting";
import i18next from "i18next";
import CropperDivModal from "./common/modal/CropperDivModal.js";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import PasswordModal from "./common/modal/PasswordModal";
import ResetModal from "./common/modal/ResetModal";
import AffiliationSelect from "./common/select/AffiliationSelect";
import moment from "moment";
import OAuthWidget from "./common/OAuthWidget";
import SamlWidget from "./common/SamlWidget";
import RegionSelect from "./common/select/RegionSelect";
import WebAuthnCredentialTable from "./table/WebauthnCredentialTable";
import ManagedAccountTable from "./table/ManagedAccountTable";
import PropertyTable from "./table/propertyTable";
import {CountryCodeSelect} from "./common/select/CountryCodeSelect";
import PopconfirmModal from "./common/modal/PopconfirmModal";
import {DeleteMfa} from "./backend/MfaBackend";
import {CheckCircleOutlined, HolderOutlined, UsergroupAddOutlined} from "@ant-design/icons";
import * as MfaBackend from "./backend/MfaBackend";
import AccountAvatar from "./account/AccountAvatar";
import FaceIdTable from "./table/FaceIdTable";
import MfaAccountTable from "./table/MfaAccountTable";
import MfaTable from "./table/MfaTable";

const {Option} = Select;

class UserEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      userName: props.userName !== undefined ? props.userName : props.match.params.userName,
      user: null,
      application: null,
      groups: null,
      organizations: [],
      applications: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
      loading: true,
      returnUrl: null,
      idCardInfo: ["ID card front", "ID card back", "ID card with person"],
      openFaceRecognitionModal: false,
    };
  }

  UNSAFE_componentWillMount() {
    this.getUser();
    if (Setting.isLocalAdminUser(this.props.account)) {
      this.getOrganizations();
    }
    this.getApplicationsByOrganization(this.state.organizationName);
    this.getUserApplication();
    this.setReturnUrl();
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    if (prevState.application !== this.state.application) {
      this.getGroups(this.state.organizationName);
    }
  }

  getUser() {
    UserBackend.getUser(this.state.organizationName, this.state.userName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          user: res.data,
          multiFactorAuths: res.data?.multiFactorAuths ?? [],
          loading: false,
        });
      });
  }

  addUserKeys() {
    UserBackend.addUserKeys(this.state.user)
      .then((res) => {
        if (res.status === "ok") {
          this.getUser();
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: res.data || [],
        });
      });
  }

  getApplicationsByOrganization(organizationName) {
    ApplicationBackend.getApplicationsByOrganization("admin", organizationName)
      .then((res) => {
        this.setState({
          applications: res.data || [],
        });

        const applications = res.data;
        if (this.state.user) {
          if (this.state.user.signupApplication === "" || applications.filter(application => application.name === this.state.user.signupApplication).length === 0) {
            if (applications.length > 0) {
              this.updateUserField("signupApplication", applications[0].name);
            } else {
              this.updateUserField("signupApplication", "");
            }
          }
        }
      });
  }

  getUserApplication() {
    ApplicationBackend.getUserApplication(this.state.organizationName, this.state.userName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          application: res.data,
        });
      });
  }

  getUserOrganization() {
    return this.state.application?.organizationObj;
  }

  isGroupsVisible() {
    const organization = this.getUserOrganization();
    if (!organization) {
      return false;
    } else {
      return organization.accountItems?.some((item) => item.name === "Groups" && item.visible);
    }
  }

  getGroups(organizationName) {
    if (!Setting.isLocalAdminUser(this.props.account)) {
      return;
    }

    if (this.isGroupsVisible()) {
      GroupBackend.getGroups(organizationName)
        .then((res) => {
          if (res.status === "ok") {
            this.setState({
              groups: res.data,
            });
          }
        });
    }
  }

  setReturnUrl() {
    const searchParams = new URLSearchParams(this.props.location.search);
    const returnUrl = searchParams.get("returnUrl");
    if (returnUrl !== null) {
      this.setState({
        returnUrl: returnUrl,
      });
    }
  }

  parseUserField(key, value) {
    if (["score", "karma", "ranking"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateUserField(key, value, idx) {
    if (this.props.account === null) {
      return;
    }

    value = this.parseUserField(key, value);

    const user = this.state.user;
    if (key === "address") {
      if (!user[key]) {
        user[key] = ["", ""];
      }
      user[key][idx] = value;
    } else {
      user[key] = value;
    }

    this.setState({
      user: user,
    });
  }

  unlinked() {
    this.getUser();
  }

  isSelf() {
    return (this.state.user.id === this.props.account?.id);
  }

  isSelfOrAdmin() {
    return this.isSelf() || Setting.isLocalAdminUser(this.props.account);
  }

  getCountryCode() {
    return this.props.account.countryCode;
  }

  deleteMfa = () => {
    this.setState({
      RemoveMfaLoading: true,
    });

    DeleteMfa({
      owner: this.state.user.owner,
      name: this.state.user.name,
    }).then((res) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully deleted"));
        this.setState({
          multiFactorAuths: res.data,
        });
      } else {
        Setting.showMessage("error", i18next.t("general:Failed to delete"));
      }
    }).finally(() => {
      this.setState({
        RemoveMfaLoading: false,
      });
    });
  };

  renderAccountItem(accountItem) {
    const isAdmin = Setting.isLocalAdminUser(this.props.account);

    let disabled = false;
    if (accountItem.modifyRule === "Self") {
      if (!this.isSelfOrAdmin()) {
        disabled = true;
      }
    } else if (accountItem.modifyRule === "Admin") {
      if (!isAdmin) {
        disabled = true;
      }
    } else if (accountItem.modifyRule === "Immutable") {
      disabled = true;
    }

    if (accountItem.name === "Organization" || accountItem.name === "Name") {
      if (this.state.user.owner === "built-in" && this.state.user.name === "admin") {
        disabled = true;
      }
    }

    if (accountItem.name === "ID card info" || accountItem.name === "ID card") {
      if (this.state.user.properties?.isIdCardVerified === "true") {
        disabled = true;
      }
    }

    let isKeysGenerated = false;
    if (this.state.user.accessKey !== "" && this.state.user.accessKey !== "") {
      isKeysGenerated = true;
    }

    if (accountItem.name === "Organization") {
      return (
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={disabled} value={this.state.user.owner} onChange={(value => {
              this.getApplicationsByOrganization(value);
              this.updateUserField("owner", value);
              this.getGroups(value);
            })}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Groups") {
      return (
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Groups"), i18next.t("general:Groups - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="multiple" style={{width: "100%"}} disabled={disabled} value={this.state.user.groups ?? []} onChange={(value => {
              if (this.state.groups?.filter(group => value.includes(`${group.owner}/${group.name}`))
                .filter(group => group.type === "Physical").length > 1) {
                Setting.showMessage("error", i18next.t("general:You can only select one physical group"));
                return;
              }

              this.updateUserField("groups", value);
            })}
            >
              {
                this.state.groups?.map((group) => <Option key={group.name} value={`${group.owner}/${group.name}`}>
                  <Space>
                    {group.type === "Physical" ? <UsergroupAddOutlined /> : <HolderOutlined />}
                    {group.displayName}
                  </Space>
                </Option>)
              }
            </Select>
          </Col>
        </Row>
      );
    } else if (accountItem.name === "ID") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel("ID", i18next.t("general:ID - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.id} disabled={disabled} onChange={e => {
              this.updateUserField("id", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Name") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.name} disabled={disabled} onChange={e => {
              this.updateUserField("name", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Display name") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.displayName} onChange={e => {
              this.updateUserField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Avatar") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Avatar"), i18next.t("general:Avatar - Tooltip"))} :
          </Col>
          {
            this.renderImage(this.state.user.avatar, i18next.t("user:Upload a photo"), i18next.t("user:Set new profile picture"), "avatar", false)
          }
        </Row>
      );
    } else if (accountItem.name === "User type") {
      let userTypes = ["normal-user", "paid-user"];
      const organization = this.getUserOrganization();
      if (organization && organization.userTypes && organization.userTypes.length > 0) {
        userTypes = organization.userTypes;
      }

      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:User type"), i18next.t("general:User type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.user.type} onChange={(value => {this.updateUserField("type", value);})}
              options={userTypes.map(item => Setting.getOption(item, item))}
            />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Password") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Password"), i18next.t("general:Password - Tooltip"))} :
          </Col>
          <Col span={22} >
            {
              (this.state.user.name === this.state.userName) ? (
                <PasswordModal user={this.state.user} userName={this.state.userName} organization={this.getUserOrganization()} account={this.props.account} disabled={disabled} />
              ) : (
                <Tooltip placement={"topLeft"} title={i18next.t("user:You have changed the username, please save your change first before modifying the password")}>
                  <span>
                    <PasswordModal user={this.state.user} userName={this.state.userName} organization={this.getUserOrganization()} account={this.props.account} disabled={true} />
                  </span>
                </Tooltip>
              )
            }
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Email") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Email"), i18next.t("general:Email - Tooltip"))} :
          </Col>
          <Col style={{paddingRight: "20px"}} span={11} >
            <Input
              value={this.state.user.email}
              style={{width: "280Px"}}
              disabled={!Setting.isLocalAdminUser(this.props.account) ? true : disabled}
              onChange={e => {
                this.updateUserField("email", e.target.value);
              }}
            />
          </Col>
          <Col span={Setting.isMobile() ? 22 : 11} >
            {/* backend auto get the current user, so admin can not edit. Just self can reset*/}
            {this.isSelf() ? <ResetModal application={this.state.application} disabled={disabled} buttonText={i18next.t("user:Reset Email...")} destType={"email"} /> : null}
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Phone") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Phone"), i18next.t("general:Phone - Tooltip"))} :
          </Col>
          <Col style={{paddingRight: "20px"}} span={11} >
            <Input.Group compact style={{width: "280Px"}}>
              <CountryCodeSelect
                style={{width: "30%"}}
                // disabled={!Setting.isLocalAdminUser(this.props.account) ? true : disabled}
                initValue={this.state.user.countryCode}
                onChange={(value) => {
                  this.updateUserField("countryCode", value);
                }}
                countryCodes={this.getUserOrganization()?.countryCodes}
              />
              <Input value={this.state.user.phone}
                style={{width: "70%"}}
                disabled={!Setting.isLocalAdminUser(this.props.account) ? true : disabled}
                onChange={e => {
                  this.updateUserField("phone", e.target.value);
                }} />
            </Input.Group>
          </Col>
          <Col span={Setting.isMobile() ? 24 : 11} >
            {this.isSelf() ? (<ResetModal application={this.state.application} countryCode={this.getCountryCode()} disabled={disabled} buttonText={i18next.t("user:Reset Phone...")} destType={"phone"} />) : null}
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Country/Region") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Country/Region"), i18next.t("user:Country/Region - Tooltip"))} :
          </Col>
          <Col span={22} >
            <RegionSelect defaultValue={this.state.user.region} onChange={(value) => {
              this.updateUserField("region", value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Location") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Location"), i18next.t("user:Location - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.location} onChange={e => {
              this.updateUserField("location", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Address") {
      return (
        <React.Fragment>
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Address"), i18next.t("user:Address - Tooltip"))} :
            </Col>
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              <span>{i18next.t("user:Address line") + " 1"}</span> :
            </Col>
            <Col span={20} >
              <Input value={!this.state.user.address ? "" : this.state.user.address[0]} onChange={e => {
                this.updateUserField("address", e.target.value, 0);
              }} />
            </Col>
          </Row>
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            </Col>
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              <span>{i18next.t("user:Address line") + " 2"}</span> :
            </Col>
            <Col span={20} >
              <Input value={!this.state.user.address ? "" : this.state.user.address[1]} onChange={e => {
                this.updateUserField("address", e.target.value, 1);
              }} />
            </Col>
          </Row>
        </React.Fragment>
      );
    } else if (accountItem.name === "Affiliation") {
      return (
        (this.state.application === null || this.state.user === null) ? null : (
          <AffiliationSelect labelSpan={(Setting.isMobile()) ? 22 : 2} application={this.state.application} user={this.state.user} onUpdateUserField={(key, value) => {return this.updateUserField(key, value);}} />
        )
      );
    } else if (accountItem.name === "Title") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Title"), i18next.t("user:Title - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.title} onChange={e => {
              this.updateUserField("title", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "ID card type") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:ID card type"), i18next.t("user:ID card type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.idCardType} onChange={e => {
              this.updateUserField("idCardType", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "ID card") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:ID card"), i18next.t("user:ID card - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.idCard} disabled={disabled} onChange={e => {
              this.updateUserField("idCard", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "ID card info") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:ID card info"), i18next.t("user:ID card info - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: "20px"}} >
              {
                [
                  {name: "ID card front", value: "idCardFront"},
                  {name: "ID card back", value: "idCardBack"},
                  {name: "ID card with person", value: "idCardWithPerson"},
                ].map((entry) => {
                  return this.renderImage(this.state.user.properties === null ? "" : (this.state.user.properties[entry.value] || ""), this.getIdCardType(entry.name), this.getIdCardText(entry.name), entry.value, disabled);
                })
              }
            </Row>
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Homepage") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Homepage"), i18next.t("user:Homepage - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.homepage} onChange={e => {
              this.updateUserField("homepage", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Bio") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Bio"), i18next.t("user:Bio - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.bio} onChange={e => {
              this.updateUserField("bio", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Tag") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Tag"), i18next.t("user:Tag - Tooltip"))} :
          </Col>
          <Col span={22} >
            {
              this.getUserOrganization()?.tags?.length > 0 ? (
                <Select virtual={false} style={{width: "100%"}} value={this.state.user.tag}
                  onChange={(value => {this.updateUserField("tag", value);})}
                  options={this.getUserOrganization()?.tags?.map((tag) => {
                    const tokens = tag.split("|");
                    const value = tokens[0];
                    const displayValue = Setting.getLanguage() !== "zh" ? tokens[0] : tokens[1];
                    return Setting.getOption(displayValue, value);
                  })} />
              ) : (
                <Input value={this.state.user.tag} onChange={e => {
                  this.updateUserField("tag", e.target.value);
                }} />
              )
            }
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Language") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Language"), i18next.t("user:Language - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.language} onChange={e => {
              this.updateUserField("language", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Gender") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Gender"), i18next.t("user:Gender - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.gender} onChange={e => {
              this.updateUserField("gender", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Birthday") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Birthday"), i18next.t("user:Birthday - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.birthday} onChange={e => {
              this.updateUserField("birthday", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Education") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Education"), i18next.t("user:Education - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.education} onChange={e => {
              this.updateUserField("education", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Balance") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Balance"), i18next.t("user:Balance - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.user.balance} onChange={value => {
              this.updateUserField("balance", value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Score") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Score"), i18next.t("user:Score - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.user.score} onChange={value => {
              this.updateUserField("score", value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Karma") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Karma"), i18next.t("user:Karma - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.user.karma} onChange={value => {
              this.updateUserField("karma", value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Ranking") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Ranking"), i18next.t("user:Ranking - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.user.ranking} onChange={value => {
              this.updateUserField("ranking", value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Signup application") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Signup application"), i18next.t("general:Signup application - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={disabled} value={this.state.user.signupApplication}
              onChange={(value => {this.updateUserField("signupApplication", value);})}
              options={this.state.applications.map((application) => Setting.getOption(application.name, application.name))
              } />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "API key") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:API key"), i18next.t("general:API key - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                {Setting.getLabel(i18next.t("general:Access key"), i18next.t("general:Access key - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Input value={this.state.user.accessKey} disabled={true} />
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                {Setting.getLabel(i18next.t("general:Access secret"), i18next.t("general:Access secret - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Input value={this.state.user.accessSecret} disabled={true} />
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col span={22} >
                <Button onClick={() => this.addUserKeys()}>{i18next.t(isKeysGenerated ? "general:update" : "general:generate")}</Button>
              </Col>
            </Row>
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Roles") {
      return (
        <Row style={{marginTop: "20px", alignItems: "center"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Roles"), i18next.t("general:Roles - Tooltip"))} :
          </Col>
          <Col span={22} >
            {
              Setting.getTags(this.state.user.roles.map(role => role.name))
            }
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Permissions") {
      return (
        <Row style={{marginTop: "20px", alignItems: "center"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Permissions"), i18next.t("general:Permissions - Tooltip"))} :
          </Col>
          <Col span={22} >
            {
              Setting.getTags(this.state.user.permissions.map(permission => permission.name))
            }
          </Col>
        </Row>
      );
    } else if (accountItem.name === "3rd-party logins") {
      return (
        !this.isSelfOrAdmin() ? null : (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:3rd-party logins"), i18next.t("user:3rd-party logins - Tooltip"))} :
            </Col>
            <Col span={22} >
              <div style={{marginBottom: 20}}>
                {
                  (this.state.application === null || this.state.user === null) ? null : (
                    this.state.application?.providers.filter(providerItem => Setting.isProviderVisible(providerItem)).map((providerItem) =>
                      (providerItem.provider.category === "OAuth" || providerItem.provider.category === "Web3") ? (
                        <OAuthWidget
                          key={providerItem.name}
                          labelSpan={(Setting.isMobile()) ? 10 : 3}
                          user={this.state.user}
                          application={this.state.application}
                          providerItem={providerItem}
                          account={this.props.account}
                          onUnlinked={() => {return this.unlinked();}} />
                      ) : (
                        <SamlWidget
                          key={providerItem.name}
                          labelSpan={(Setting.isMobile()) ? 10 : 3}
                          user={this.state.user}
                          application={this.state.application}
                          providerItem={providerItem}
                          onUnlinked={() => {return this.unlinked();}} />
                      )
                    )
                  )
                }
              </div>
            </Col>
          </Row>
        )
      );
    } else if (accountItem.name === "Properties") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Properties"), i18next.t("user:Properties - Tooltip"))} :
          </Col>
          <Col span={22} >
            <PropertyTable properties={this.state.user.properties} onUpdateTable={(value) => {this.updateUserField("properties", value);}} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Is admin") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Is admin"), i18next.t("user:Is admin - Tooltip"))} :
          </Col>
          <Col span={(Setting.isMobile()) ? 22 : 2} >
            <Switch disabled={disabled} checked={this.state.user.isAdmin} onChange={checked => {
              this.updateUserField("isAdmin", checked);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Is forbidden") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Is forbidden"), i18next.t("user:Is forbidden - Tooltip"))} :
          </Col>
          <Col span={(Setting.isMobile()) ? 22 : 2} >
            <Switch checked={this.state.user.isForbidden} onChange={checked => {
              this.updateUserField("isForbidden", checked);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Is deleted") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Is deleted"), i18next.t("user:Is deleted - Tooltip"))} :
          </Col>
          <Col span={(Setting.isMobile()) ? 22 : 2} >
            <Switch checked={this.state.user.isDeleted} onChange={checked => {
              this.updateUserField("isDeleted", checked);
              this.updateUserField("deletedTime", checked ? moment().format() : "");
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "MFA items") {
      return (<Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("general:MFA items"), i18next.t("general:MFA items - Tooltip"))} :
        </Col>
        <Col span={22} >
          <MfaTable
            title={i18next.t("general:MFA items")}
            table={this.state.user.mfaItems ?? []}
            onUpdateTable={(value) => {this.updateUserField("mfaItems", value);}}
          />
        </Col>
      </Row>);
    } else if (accountItem.name === "Multi-factor authentication") {
      return (
        !this.isSelfOrAdmin() ? null : (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
              {Setting.getLabel(i18next.t("mfa:Multi-factor authentication"), i18next.t("mfa:Multi-factor authentication - Tooltip "))} :
            </Col>
            <Col span={22} >
              <Card size="small" title={i18next.t("mfa:Multi-factor methods")}
                extra={this.state.multiFactorAuths?.some(mfaProps => mfaProps.enabled) ?
                  <PopconfirmModal
                    text={i18next.t("general:Disable")}
                    title={i18next.t("general:Sure to disable") + "?"}
                    onConfirm={() => this.deleteMfa()}
                  /> : null
                }>
                <List
                  rowKey="mfaType"
                  itemLayout="horizontal"
                  dataSource={this.state.multiFactorAuths}
                  renderItem={(item, index) => (
                    <List.Item>
                      <Space>
                        {i18next.t("general:Type")}: {item.mfaType}
                        {item.secret}
                      </Space>
                      {item.enabled ? (
                        <Space>
                          <Tag icon={<CheckCircleOutlined />} color="success">
                            {i18next.t("general:Enabled")}
                          </Tag>
                          {item.isPreferred ?
                            <Tag icon={<CheckCircleOutlined />} color="blue" style={{marginRight: 20}} >
                              {i18next.t("mfa:preferred")}
                            </Tag> :
                            <Button type="primary" style={{marginRight: 20}} onClick={() => {
                              const values = {
                                owner: this.state.user.owner,
                                name: this.state.user.name,
                                mfaType: item.mfaType,
                              };
                              MfaBackend.SetPreferredMfa(values).then((res) => {
                                if (res.status === "ok") {
                                  this.setState({
                                    multiFactorAuths: res.data,
                                  });
                                }
                              });
                            }}>
                              {i18next.t("mfa:Set preferred")}
                            </Button>
                          }
                          {this.isSelf() ? <Button type={"default"} onClick={() => {
                            this.props.history.push(`/mfa/setup?mfaType=${item.mfaType}`);
                          }}>
                            {i18next.t("general:Edit")}
                          </Button> : null}
                        </Space>
                      ) :
                        <Space>
                          {item.mfaType !== TotpMfaType && Setting.isLocalAdminUser(this.props.account) && !this.isSelf() ?
                            <EnableMfaModal user={this.state.user} mfaType={item.mfaType} onSuccess={() => {
                              this.getUser();
                            }} /> : null}
                          {this.isSelf() ? <Button type={"default"} onClick={() => {
                            this.props.history.push(`/mfa/setup?mfaType=${item.mfaType}`);
                          }}>
                            {i18next.t("mfa:Setup")}
                          </Button> : null}
                        </Space>}
                    </List.Item>
                  )}
                />
              </Card>
            </Col>
          </Row>
        )
      );
    } else if (accountItem.name === "WebAuthn credentials") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:WebAuthn credentials"), i18next.t("user:WebAuthn credentials"))} :
          </Col>
          <Col span={22} >
            <WebAuthnCredentialTable isSelf={this.isSelf()} table={this.state.user.webauthnCredentials} updateTable={(table) => {this.updateUserField("webauthnCredentials", table);}} refresh={this.getUser.bind(this)} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Last change password time") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Last change password time"), i18next.t("user:Last change password time"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.user.lastChangePasswordTime} onChange={e => {
              this.updateUserField("lastChangePasswordTime", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Managed accounts") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Managed accounts"), i18next.t("user:Managed accounts"))} :
          </Col>
          <Col span={22} >
            <ManagedAccountTable
              title={i18next.t("user:Managed accounts")}
              table={this.state.user.managedAccounts}
              onUpdateTable={(table) => {this.updateUserField("managedAccounts", table);}}
              applications={this.state.applications}
            />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Face ID") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Face IDs"), i18next.t("user:Face IDs"))} :
          </Col>
          <Col span={22} >
            <FaceIdTable
              title={i18next.t("user:Face IDs")}
              table={this.state.user.faceIds}
              {...this.props}
              onUpdateTable={(table) => {this.updateUserField("faceIds", table);}}
            />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "MFA accounts") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:MFA accounts"), i18next.t("user:MFA accounts"))} :
          </Col>
          <Col span={22} >
            <MfaAccountTable
              title={i18next.t("user:MFA accounts")}
              table={this.state.user.mfaAccounts}
              accessToken={this.props.account?.accessToken}
              icon={this.state.user.avatar}
              onUpdateTable={(table) => {this.updateUserField("mfaAccounts", table);}}
            />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Need update password") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Need update password"), i18next.t("user:Need update password - Tooltip"))} :
          </Col>
          <Col span={(Setting.isMobile()) ? 22 : 2} >
            <Switch disabled={(!this.state.user.phone) && (!this.state.user.email) && (!this.state.user.mfaProps)} checked={this.state.user.needUpdatePassword} onChange={checked => {
              this.updateUserField("needUpdatePassword", checked);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "IP whitelist") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:IP whitelist"), i18next.t("general:IP whitelist - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.user.ipWhitelist} onChange={e => {
              this.updateUserField("ipWhitelist", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "First name") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:First name"), i18next.t("general:First name - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.user.firstName} onChange={e => {
              this.updateUserField("firstName", e.target.value);
            }} />
          </Col>
        </Row>
      );
    } else if (accountItem.name === "Last name") {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Last name"), i18next.t("general:Last name - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.user.lastName} onChange={e => {
              this.updateUserField("lastName", e.target.value);
            }} />
          </Col>
        </Row>
      );
    }
  }

  renderImage(imgUrl, title, set, tag, disabled) {
    return (
      <Col span={4} style={{textAlign: "center", margin: "auto", marginLeft: "20px"}} key={tag}>
        {
          imgUrl ?
            <div style={{marginBottom: "10px"}}>
              <a target="_blank" rel="noreferrer" href={imgUrl} style={{marginBottom: "10px"}}>
                <AccountAvatar src={imgUrl} alt={imgUrl} height={150} />
              </a>
            </div>
            :
            <Col style={{height: "78%", border: "1px dotted grey", borderRadius: 3, marginBottom: "10px"}}>
              <div style={{fontSize: 30, margin: 10}}>+</div>
              <div style={{verticalAlign: "middle", marginBottom: 10}}>{`(${i18next.t("general:empty")})`}</div>
            </Col>
        }
        {
          (this.props.account === null) ? null : (
            <CropperDivModal disabled={disabled} tag={tag} setTitle={set} buttonText={`${title}...`} title={title} user={this.state.user} organization={this.getUserOrganization()} />
          )
        }
      </Col>
    );
  }

  renderUser() {
    return (
      <Card size="small" title={
        (this.props.account === null) ? i18next.t("user:User Profile") : (
          <div>
            {this.state.mode === "add" ? i18next.t("user:New User") : (this.isSelf() ? i18next.t("account:My Account") : i18next.t("user:Edit User"))}&nbsp;&nbsp;&nbsp;&nbsp;
            <Button onClick={() => this.submitUserEdit(false)}>{i18next.t("general:Save")}</Button>
            <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitUserEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
            {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteUser()}>{i18next.t("general:Cancel")}</Button> : null}
          </div>
        )
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Form>
          {
            this.getUserOrganization()?.accountItems?.map(accountItem => {
              if (!accountItem.visible) {
                return null;
              }

              const isAdmin = Setting.isLocalAdminUser(this.props.account);

              if (accountItem.viewRule === "Self") {
                if (!this.isSelfOrAdmin()) {
                  return null;
                }
              } else if (accountItem.viewRule === "Admin") {
                if (!isAdmin) {
                  return null;
                }
              }
              return (
                <React.Fragment key={accountItem.name}>
                  <Form.Item name={accountItem.name}
                    validateTrigger="onChange"
                    rules={[
                      {
                        pattern: accountItem.regex ? new RegExp(accountItem.regex, "g") : null,
                        message: i18next.t("user:This field value doesn't match the pattern rule"),
                      },
                    ]}
                    style={{margin: 0}}>
                    {this.renderAccountItem(accountItem)}
                  </Form.Item>
                </React.Fragment>
              );
            })
          }
        </Form>
      </Card>
    );
  }

  getIdCardType(key) {
    if (key === "ID card front") {
      return i18next.t("user:ID card front");
    } else if (key === "ID card back") {
      return i18next.t("user:ID card back");
    } else if (key === "ID card with person") {
      return i18next.t("user:ID card with person");
    } else {
      return "Unknown Id card name: " + key;
    }
  }

  getIdCardText(key) {
    if (key === "ID card front") {
      return i18next.t("user:Upload ID card front picture");
    } else if (key === "ID card back") {
      return i18next.t("user:Upload ID card back picture");
    } else if (key === "ID card with person") {
      return i18next.t("user:Upload ID card with person picture");
    } else {
      return "Unknown Id card name: " + key;
    }
  }

  submitUserEdit(exitAfterSave) {
    const user = Setting.deepCopy(this.state.user);
    UserBackend.updateUser(this.state.organizationName, this.state.userName, user)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            organizationName: this.state.user.owner,
            userName: this.state.user.name,
          });

          if (this.props.history !== undefined) {
            if (exitAfterSave) {
              const userListUrl = sessionStorage.getItem("userListUrl");
              if (userListUrl !== null) {
                this.props.history.push(userListUrl);
              } else {
                if (Setting.isLocalAdminUser(this.props.account)) {
                  this.props.history.push("/users");
                } else {
                  this.props.history.push("/");
                }
              }
            } else {
              if (location.pathname !== "/account") {
                this.props.history.push(`/users/${this.state.user.owner}/${this.state.user.name}`);
              }
            }
          } else {
            if (exitAfterSave) {
              if (this.state.returnUrl) {
                window.location.href = this.state.returnUrl;
              }
            }
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateUserField("owner", this.state.organizationName);
          this.updateUserField("name", this.state.userName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteUser() {
    UserBackend.deleteUser(this.state.user)
      .then((res) => {
        if (res.status === "ok") {
          const userListUrl = sessionStorage.getItem("userListUrl");
          if (userListUrl !== null) {
            this.props.history.push(userListUrl);
          } else {
            this.props.history.push("/users");
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.loading ? <Spin size="large" style={{marginLeft: "50%", marginTop: "10%"}} /> : (
            this.state.user !== null ? this.renderUser() :
              <Result
                status="404"
                title="404 NOT FOUND"
                subTitle={i18next.t("general:Sorry, the user you visited does not exist or you are not authorized to access this user.")}
                extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>}
              />
          )
        }
        {
          (this.state.user === null || this.props.account === null) ? null :
            <div style={{marginTop: "20px", marginLeft: "40px"}}>
              <Button size="large" onClick={() => this.submitUserEdit(false)}>{i18next.t("general:Save")}</Button>
              <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitUserEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
              {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteUser()}>{i18next.t("general:Cancel")}</Button> : null}
            </div>
        }
      </div>
    );
  }
}

export default withRouter(UserEditPage);
