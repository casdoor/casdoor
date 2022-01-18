// Copyright 2021 The casbin Authors. All Rights Reserved.
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
import {Button, Col, Form, Select, Input, Row, Steps, Image} from "antd";
import * as AuthBackend from "./AuthBackend";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Util from "./Util";
import * as Setting from "../Setting";
import i18next from "i18next";
import {CountDownInput} from "../common/CountDownInput";
import * as UserBackend from "../backend/UserBackend";
import {CheckCircleOutlined, KeyOutlined, LockOutlined, SolutionOutlined, UserOutlined} from "@ant-design/icons";
import CustomGithubCorner from "../CustomGithubCorner";
import QRCode from "qrcode.react"

const { Step } = Steps;
const { Option } = Select;

class TotpPage extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			classes: props,
			account: props.account,
			applicationName:
				props.applicationName !== undefined
					? props.applicationName
					: props.match === undefined
						? null
						: props.match.params.applicationName,
			application: null,
			msg: null,
			userId: "",
			username: "",
			password: "",
			qrcodeUrl: "",
			secret: "",
			current: 0,
		};
	}

	UNSAFE_componentWillMount() {
		if (this.state.applicationName !== undefined) {
			this.getApplication();
		} else {
			Util.showMessage(
				"error",
				i18next.t(`forget:Unknown forgot type: `) + this.state.type
			);
		}
	}

	getApplication() {
		if (this.state.applicationName === null) {
			return;
		}

		ApplicationBackend.getApplication("admin", this.state.applicationName).then(
			(application) => {
				this.setState({
					application: application,
				});
			}
		);
	}

	getApplicationObj() {
		if (this.props.application !== undefined) {
			return this.props.application;
		} else {
			return this.state.application;
		}
	}

	onFormFinish(name, info, forms) {
		switch (name) {
			case "step1":
				let user = this.state.account;
				user.password = forms.step1.getFieldValue("password");
				UserBackend.checkUserPassword(user).then(res => {
					if (res.status === "ok") {
						UserBackend.initTOTP().then((res) => {
							this.setState({current: 1, qrcodeUrl: decodeURI(res.qrcodeUrl), recoveryCode: res.recoveryCode, secret: res.secret})
						});
					} else {
						Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
					}
				});
				break;
			case "step2":
				const code = forms.step2.getFieldValue("code");
				UserBackend.setTOTP(this.state.secret, code).then(res => {
					if (res.status === "ok") {
						this.setState({current: 2})
					} else {
						Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
					}
				});
				break
			default:
				break
		}
	}

	onFinish() {
		Setting.goToLinkSoft(this, "/account")
	}

	onFinishFailed(values, errorFields) {}

	renderOptions() {
		let options = [];

		if (this.state.phone !== "") {
			options.push(
				<Option key={"phone"} value={"phone"}>
					&nbsp;&nbsp;{Setting.getMaskedPhone(this.state.phone)}
				</Option>
			);
		}

		if (this.state.email !== "") {
			options.push(
				<Option key={"email"} value={"email"}>
					&nbsp;&nbsp;{Setting.getMaskedEmail(this.state.email)}
				</Option>
			);
		}

		return options;
	}

	renderForm(application) {
		return (
			<Form.Provider onFormFinish={(name, {info, forms}) => {
				this.onFormFinish(name, info, forms);
			}}>
				{/* STEP 1: input username -> get email & phone */}
				<Form
					hidden={this.state.current !== 0}
					ref={this.form}
					name="step1"
					onFinishFailed={(errorInfo) => console.log(errorInfo)}
					initialValues={{
						application: application.name,
						organization: application.organization,
					}}
					style={{ width: "300px" }}
					size="large"
				>
					<Form.Item
						style={{ height: 0, visibility: "hidden" }}
						name="application"
						rules={[
							{
								required: true,
								message: i18next.t(
									`forget:Please input your application!`
								),
							},
						]}
					/>
					<Form.Item
						style={{ height: 0, visibility: "hidden" }}
						name="organization"
						rules={[
							{
								required: true,
								message: i18next.t(
									`forget:Please input your organization!`
								),
							},
						]}
					/>
					<Form.Item
						name="password"
						rules={[
							{
								required: true,
								message: i18next.t(
									"forget:Please input your password!"
								),
								whitespace: true,
							},
						]}
					>
						<Input
							onChange={(e) => {
								this.setState({
									password: e.target.value,
								});
							}}
							prefix={<UserOutlined />}
							placeholder={i18next.t("login:password")}
							type="password"
						/>
					</Form.Item>
					<br />
					<Form.Item>
						<Button block type="primary" htmlType="submit">
							{i18next.t("forget:Next Step")}
						</Button>
					</Form.Item>
				</Form>

				{/* STEP 2: verify code */}
				<Form
					hidden={this.state.current !== 1}
					ref={this.form}
					name="step2"
					onFinishFailed={(errorInfo) =>
						this.onFinishFailed(
							errorInfo.values,
							errorInfo.errorFields,
							errorInfo.outOfDate
						)
					}
					initialValues={{
						application: application.name,
						organization: application.organization,
					}}
					style={{ width: "300px" }}
					size="large"
				>
					<Form.Item
						style={{ height: 0, visibility: "hidden" }}
						name="application"
						rules={[
							{
								required: true,
								message: i18next.t(
									`forget:Please input your application!`
								),
							},
						]}
					/>
					<Form.Item
						style={{ height: 0, visibility: "hidden" }}
						name="organization"
						rules={[
							{
								required: true,
								message: i18next.t(
									`forget:Please input your organization!`
								),
							},
						]}
					/>
					<QRCode value={this.state.qrcodeUrl} size={200} />
					<Form.Item
						style={{marginTop: "20px"}}
						name="code"
						rules={[
							{
								required: true,
								message: i18next.t(
									"totp:Please input your code!"
								),
								whitespace: true,
							}
						]}
					>
						<Input
							prefix={<UserOutlined />}
							placeholder={i18next.t("totp:code")}
						/>
					</Form.Item>
					<br />
					<Form.Item>
						<Button
							block
							type="primary"
							htmlType="submit"
						>
							{i18next.t("forget:Next Step")}
						</Button>
					</Form.Item>
				</Form>

				{/* STEP 3 */}
				<Form
					hidden={this.state.current !== 2}
					ref={this.form}
					name="step3"
					onFinish={(values) => this.onFinish(values)}
					onFinishFailed={(errorInfo) =>
						this.onFinishFailed(
							errorInfo.values,
							errorInfo.errorFields,
							errorInfo.outOfDate
						)
					}
					size="large"
				>
					<h2>{i18next.t("totp:You have enabled 2-step verification successfully!")}</h2>
					<br />
					<Form.Item hidden={this.state.current !== 2}>
						<Button block type="primary"  htmlType="submit">
							{i18next.t("totp:Done")}
						</Button>
					</Form.Item>
				</Form>
			</Form.Provider>
		);
	}

	render() {
		const application = this.getApplicationObj();
		if (application === null) {
			return Util.renderMessageLarge(this, this.state.msg);
		}

		return (
			<Row>
				<Col span={24} style={{justifyContent: "center"}}>
					<Row>
						<Col span={24}>
							<div style={{marginTop: "80px", marginBottom: "10px", textAlign: "center"}}>
								{
									Setting.renderHelmet(application)
								}
								<CustomGithubCorner />
								{
									Setting.renderLogo(application)
								}
							</div>
						</Col>
					</Row>
					<Row>
						<Col span={24}>
							<div style={{textAlign: "center", fontSize: "28px"}}>
								{i18next.t("forget:Protect your account with 2-Step Verification")}
							</div>
							<div style={{textAlign: "center", fontSize: "16px", marginTop: "10px"}}>
								{i18next.t("forget:Each time you sign in to your Account, you'll need your password and a verification code.")}
							</div>
						</Col>
					</Row>
					<Row>
						<Col span={24}>
							<Steps
								current={this.state.current}
								style={{
									width: "90%",
									maxWidth: "500px",
									margin: "auto",
									marginTop: "80px",
								}}
							>
								<Step
									title={i18next.t("forget:Verify Password")}
									icon={<UserOutlined />}
								/>
								<Step
									title={i18next.t("forget:Verify Code")}
									icon={<SolutionOutlined />}
								/>
								<Step
									title={i18next.t("forget:Done")}
									icon={<KeyOutlined />}
								/>
							</Steps>
						</Col>
					</Row>
				</Col>
				<Col span={24} style={{ display: "flex", justifyContent: "center" }}>
					<div style={{ marginTop: "10px", textAlign: "center" }}>
						{this.renderForm(application)}
					</div>
				</Col>
			</Row>
		);
	}
}

export default TotpPage;