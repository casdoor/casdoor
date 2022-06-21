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
import i18next from "i18next";
import {Link} from "react-router-dom";
import {Input, Switch, Table, Select, Space, Button, Card, Row, Col} from 'antd';
import * as Setting from "./Setting";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as UserBackend from "./backend/UserBackend";
import SelectRegionBox from "./SelectRegionBox";

const {Search} = Input;
const {Option} = Select;

const thirdPartyAccount = [
  "github", "google", "qq", "wechat", "facebook", "dingtalk", "weibo", "gitee", "linkedin", "wecom", "lark", "gitlab", "adfs", "baidu", "alipay", "infoflow", "apple", "azuread", "slack", "steam", "bilibili", "okta", "douyin", "ldap"
]

class UserMergePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      mainAccount: null,
      mergedAccount: null,
      newAccount: null,
      organizations: [],

      organizationName: "",
    };
  }
  UNSAFE_componentWillMount() {
    if (!Setting.isAdminUser(this.props.account)) {
      Setting.showMessage("error", i18next.t("user:Not authorized"));
      this.props.history.push("/users")
      return
    }
    OrganizationBackend.getOrganizations("admin").then(
      (res) => {
        console.log(res)
        this.setState({organizations: res})
      }
    )
  }

  render() {
    let generateInputFor3rdAccount = () => {
      let res = []
      for (let i in thirdPartyAccount) {
        res.push(
          (
            <Row style={{marginTop: '20px'}} key={thirdPartyAccount[i]}>
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(thirdPartyAccount[i], thirdPartyAccount[i])} :
              </Col>
              <Col span={22} >
                <Input value={this.state.newAccount === null ? "" : this.state.newAccount[thirdPartyAccount[i]]} />
              </Col>
            </Row>
          )
        )
      }
      return res
    }


    return (
      <div >
        {/* Select two user to merge */}
        <Space>
          <span>{i18next.t("user:Choose an Organization")}</span>
          <Select
            style={{width: 120, height: 30, margin: 0, verticalAlign: "top"}} onChange={this.onOrganizationChange.bind(this)}>
            {this.renderSelect()}
          </Select>
        </Space>
        <hr />
        <Space>
          <Search
            addonBefore={i18next.t("user:Main User")}
            placeholder={i18next.t("user:input user name")}
            onSearch={(value) => {this.searchUser(value, "mainAccount")}}
            style={{width: 300}}
          />
          <Search
            addonBefore={i18next.t("user:User to be merged")}
            placeholder={i18next.t("user:input user name")}
            onSearch={(value) => {this.searchUser(value, "mergedAccount")}}
            style={{width: 300}}
          />
          <Button type="primary" onClick={this.onGenerateNewUser.bind(this)}>{i18next.t("user:Preview Merged Account")}</Button>
          <Button type="primary" onClick={this.onSubmitMerge.bind(this)}>{i18next.t("user:Submit")}</Button>
        </Space>
        <hr />
        {/* Show information of 2 user */}
        {this.renderTable()}
        <hr />
        {/* Allow user to edit it */}
        <Card size="small" style={(Setting.isMobile()) ? {margin: '5px'} : {}} type="inner" title={i18next.t("user:New Merged User")}>
          {/* Oranization */}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={this.state.newAccount?.owner} disabled={true} />
            </Col>
          </Row>
          {/* ID */}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:ID"), i18next.t("general:ID - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={this.state.newAccount?.id} disabled={true} />
            </Col>
          </Row>
          {/* Username */}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={this.state.newAccount?.name} disabled={true} />
            </Col>
          </Row>
          {/* Display name */}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={this.state.newAccount?.displayName} onChange={e => {
                this.updateUserField('displayName', e.target.value);
              }} />
            </Col>
          </Row>
          {/* todo: avator */}
          {/*User type */}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:User type"), i18next.t("general:User type - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Select virtual={false} style={{width: '100%'}} value={this.state.newAccount?.type} onChange={(value => {this.updateUserField('type', value);})}>
                {
                  ['normal-user']
                    .map((item, index) => <Option key={index} value={item}>{item}</Option>)
                }
              </Select>
            </Col>
          </Row>
          {/* todo: password */}
          {/*Email*/}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Email"), i18next.t("general:Email - Tooltip"))} :
            </Col>
            <Col style={{paddingRight: '20px'}} span={11} >
              <Input value={this.state.newAccount?.email}
                onChange={e => {
                  this.updateUserField('email', e.target.value);
                }} />
            </Col>
          </Row>
          {/*Phone*/}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Phone"), i18next.t("general:Phone - Tooltip"))} :
            </Col>
            <Col style={{paddingRight: '20px'}} span={11} >
              <Input value={this.state.newAccount?.phone} addonBefore={`+${this.state.application?.organizationObj.phonePrefix}`}
                onChange={e => {
                  this.updateUserField('phone', e.target.value);
                }} />
            </Col>
          </Row>
          {/*Region*/}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Country/Region"), i18next.t("user:Country/Region - Tooltip"))} :
            </Col>
            <Col span={22} >
              <SelectRegionBox defaultValue={this.state.newAccount?.region} onChange={(value) => {
                this.updateUserField("region", value);
              }} />
            </Col>
          </Row>
          {/*City*/}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Location"), i18next.t("user:Location - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={this.state.newAccount?.location} onChange={e => {
                this.updateUserField('location', e.target.value);
              }} />
            </Col>
          </Row>

          {/*title*/}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Title"), i18next.t("user:Title - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={this.state.newAccount?.title} onChange={e => {
                this.updateUserField('title', e.target.value);
              }} />
            </Col>
          </Row>
          {/*Homepage*/}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Homepage"), i18next.t("user:Homepage - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={this.state.newAccount?.homepage} onChange={e => {
                this.updateUserField('homepage', e.target.value);
              }} />
            </Col>
          </Row>
          {/*Bio*/}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Bio"), i18next.t("user:Bio - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={this.state.newAccount?.bio} onChange={e => {
                this.updateUserField('bio', e.target.value);
              }} />
            </Col>
          </Row>
          {/*tag*/}
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Tag"), i18next.t("user:Tag - Tooltip"))} :
            </Col>
            <Col span={22} >
              {
                this.state.application?.organizationObj.tags?.length > 0 ? (
                  <Select virtual={false} style={{width: '100%'}} value={this.state.newAccount?.tag} onChange={(value => {this.updateUserField('tag', value);})}>
                    {
                      this.state.application.organizationObj.tags?.map((tag, index) => {
                        const tokens = tag.split("|");
                        const value = tokens[0];
                        const displayValue = Setting.getLanguage() !== "zh" ? tokens[0] : tokens[1];
                        return <Option key={index} value={value}>{displayValue}</Option>
                      })
                    }
                  </Select>
                ) : (
                  <Input value={this.state.newAccount?.tag} onChange={e => {
                    this.updateUserField('tag', e.target.value);
                  }} />
                )
              }
            </Col>
          </Row>

          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Signup application"), i18next.t("general:Signup application - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input style={{width: '100%'}} disabled={true} value={this.state.newAccount?.signupApplication} />
            </Col>
          </Row>

          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Is admin"), i18next.t("user:Is admin - Tooltip"))} :
            </Col>
            <Col span={(Setting.isMobile()) ? 22 : 2} >
              <Switch checked={this.state.newAccount?.isAdmin} onChange={checked => {
                this.updateUserField('isAdmin', checked);
              }} />
            </Col>
          </Row>
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Is global admin"), i18next.t("user:Is global admin - Tooltip"))} :
            </Col>
            <Col span={(Setting.isMobile()) ? 22 : 2} >
              <Switch checked={this.state.newAccount?.isGlobalAdmin} onChange={checked => {
                this.updateUserField('isGlobalAdmin', checked);
              }} />
            </Col>
          </Row>
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Is forbidden"), i18next.t("user:Is forbidden - Tooltip"))} :
            </Col>
            <Col span={(Setting.isMobile()) ? 22 : 2} >
              <Switch checked={this.state.newAccount?.isForbidden} onChange={checked => {
                this.updateUserField('isForbidden', checked);
              }} />
            </Col>
          </Row>
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("user:Is deleted"), i18next.t("user:Is deleted - Tooltip"))} :
            </Col>
            <Col span={(Setting.isMobile()) ? 22 : 2} >
              <Switch checked={this.state.newAccount?.isDeleted} onChange={checked => {
                this.updateUserField('isDeleted', checked);
              }} />
            </Col>
          </Row>

          {/*3rd accounts*/}
          {generateInputFor3rdAccount()}
        </Card>
      </div>
    )
  }

  renderTable() {
    let columns = [
      {
        title: "",
        fixed: 'left',
        render: (text, record, index) => {
          switch (index) {
            case 0:
              return (i18next.t("user:Main Account"))
            case 1:
              return (i18next.t("user:Merged Account"))
            case 2:
              return (i18next.t("user:New Account"))
            default:
              ;
          }
        }
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: 'owner',
        key: 'owner',
        width: (Setting.isMobile()) ? "100px" : "120px",
        fixed: 'left',
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text}`}>
              {text}
            </Link>
          )
        }
      },

      {
        title: i18next.t("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: (Setting.isMobile()) ? "80px" : "110px",
        fixed: 'left',
        render: (text, record, index) => {
          return (
            <Link to={`/users/${record.owner}/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: 'createdTime',
        key: 'createdTime',
        width: '160px',
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: 'displayName',
        key: 'displayName',
      },
      {
        title: i18next.t("user:ID"),
        dataIndex: 'id',
        key: 'id',
      },
      {
        title: i18next.t("general:Avatar"),
        dataIndex: 'avatar',
        key: 'avatar',
        width: '80px',
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              <img src={text} alt={text} width={50} />
            </a>
          )
        }
      },
      {
        title: i18next.t("general:Application"),
        dataIndex: 'signupApplication',
        key: 'signupApplication',
        width: (Setting.isMobile()) ? "100px" : "120px",
        render: (text, record, index) => {
          return (
            <Link to={`/applications/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Email"),
        dataIndex: 'email',
        key: 'email',
        width: '160px',
        render: (text, record, index) => {
          return (
            <a href={`mailto:${text}`}>
              {text}
            </a>
          )
        }
      },
      {
        title: i18next.t("general:Phone"),
        dataIndex: 'phone',
        key: 'phone',
        width: '120px',
      },
      {
        title: i18next.t("user:Affiliation"),
        dataIndex: 'affiliation',
        key: 'affiliation',
        width: '140px',
      },
      {
        title: i18next.t("user:Country/Region"),
        dataIndex: 'region',
        key: 'region',
        width: '140px',
      },
      {
        title: i18next.t("user:Tag"),
        dataIndex: 'tag',
        key: 'tag',
        width: '110px',
      },
      {
        title: i18next.t("user:Is admin"),
        dataIndex: 'isAdmin',
        key: 'isAdmin',
        width: '110px',
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          )
        }
      },
      {
        title: i18next.t("user:Is global admin"),
        dataIndex: 'isGlobalAdmin',
        key: 'isGlobalAdmin',
        width: '140px',
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          )
        }
      },
      {
        title: i18next.t("user:Is forbidden"),
        dataIndex: 'isForbidden',
        key: 'isForbidden',
        width: '110px',
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          )
        }
      },
      {
        title: i18next.t("user:Is deleted"),
        dataIndex: 'isDeleted',
        key: 'isDeleted',
        width: '110px',
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          )
        }
      },
    ];
    //add 3rd accounts
    for (let i in thirdPartyAccount) {
      columns.push({
        title: thirdPartyAccount[i],
        dataIndex: thirdPartyAccount[i],
        key: thirdPartyAccount[i],
        width: '110px'
      })
    }
    return (
      <Table scroll={{x: 'max-content'}} columns={columns}
        dataSource={[{...this.state.mainAccount, key: 0}, {...this.state.mergedAccount, key: 1}]}
        rowKey="key" size="large" bordered pagination={false} />
    )
  }

  renderSelect() {
    let res = []
    for (let i = 0; i < this.state.organizations.length; i++) {
      let name = this.state.organizations[i].name
      res.push((<Option value={name} key={name}>{name}</Option>))
    }
    return res
  }

  onOrganizationChange(value) {
    this.setState({organizationName: value})
  }

  onGenerateNewUser() {
    if (this.state.mainAccount === null) {
      Setting.showMessage("error", i18next.t("user:Please select the MainAccount"));
      return
    }

    if (this.state.mergedAccount === null) {
      Setting.showMessage("error", i18next.t("user:Please select the MergedAccount"));
      return
    }
    if (this.state.mainAccount.name === this.state.mergedAccount.name) {
      Setting.showMessage("error", i18next.t("user:Please select different users to mege"));
      return
    }
    let newAccount = {...this.state.mainAccount}
    for (let key in this.state.mergedAccount) {
      if (!newAccount[key]) {
        newAccount[key] = this.state.mergedAccount[key]
      }
    }
    this.setState({newAccount: newAccount})
  }

  async onSubmitMerge() {
    if (this.state.newAccount === null || this.state.mergedAccount === null || this.state.mainAccount === null) {
      Setting.showMessage("error", i18next.t("user:Please Generate new merged account"));
      return
    }
    //step 1: update main account
    let res = await UserBackend.updateUser(this.state.organizationName, this.state.newAccount.name, this.state.newAccount)
    console.log(res)
    if (res.status !== "ok") {
      Setting.showMessage("error", res?.msg);
      return
    }

    //step 2: delete merged account
    res = await UserBackend.deleteUser(this.state.mergedAccount)
    if (res.status !== "ok") {
      Setting.showMessage("error", res?.msg);
      return
    }
    Setting.showMessage("success", i18next.t("user:Users are merged successfully"));
    this.setState({
      mainAccount: null,
      mergedAccount: null,
      newAccount: null
    })

  }

  searchUser(value, which) {
    if (this.state.organizationName === "") {
      Setting.showMessage("error", i18next.t("user:Please Select an organization"));
      return
    }
    if (value === "") {
      Setting.showMessage("error", i18next.t("user:Main Account username should not be empty"));
      return
    }

    UserBackend.getUser(this.state.organizationName, value).then(
      (res) => {
        console.log(res)
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
        } else {
          if (which === "mainAccount") {

            this.setState({mainAccount: res})
          } else if (which === "mergedAccount") {
            this.setState({mergedAccount: res})
          }
        }
      }
    )

  }

  updateUserField(key, value) {
    let user = this.state.newAccount;
    if (!user) {
      return
    }
    user[key] = value;
    this.setState({
      newAccount: user,
    });
  }

}
export default UserMergePage;