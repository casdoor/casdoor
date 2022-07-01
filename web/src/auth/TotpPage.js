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

import React, {useState} from "react";
import {Button, Col, Input, message, Row, Select, Spin, Steps} from "antd";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as UserBackend from "../backend/UserBackend";
import {CheckOutlined, KeyOutlined, UserOutlined} from "@ant-design/icons";
import CustomGithubCorner from "../CustomGithubCorner";
import QRCode from "qrcode.react";
import {useFormik} from "formik";

const {Step} = Steps;
const {Option} = Select;

function CheckPassword({user, onSuccess, onFail}) {
	const formik = useFormik({
		initialValues: {
			password: ""
		},
		onSubmit: ({password}) => {
			const data = {...user, password};
			UserBackend.checkUserPassword(data).then(res => {
					if (res.status === "ok") {
						onSuccess(res);
					} else {
						onFail(res);
					}
				}
			).finally(() => {
				formik.setSubmitting(false);
			});
		}
	});

	return (
		<form style={{width: "300px"}} onSubmit={formik.handleSubmit}>
			<Input
				onChange={formik.handleChange("password")}
				prefix={<UserOutlined/>}
				placeholder={i18next.t("two-factor:Password")}
				type="password"
			/>
			<Button style={{marginTop: 24}} loading={formik.isSubmitting} block
					type="primary" htmlType="submit">
				{i18next.t("two-factor:Next step")}
			</Button>
		</form>
	);
}

function VerityTotp({totp, onSuccess, onFail}) {
	const formik = useFormik(
		{
			initialValues: {
				passcode: ""
			},
			onSubmit: ({passcode}) => {
				const data = {secret: totp.secret, passcode};
				UserBackend.twoFactorSetupVerityTotp(data).then(res => {
						if (res.status === "ok") {
							onSuccess(res);
						} else {
							onFail(res);
						}
					}
				).finally(() => {
					formik.setSubmitting(false);
				});
			}
		});
	return (
		<form style={{width: "300px"}} onSubmit={formik.handleSubmit}>
			<QRCode value={totp.url} size={200}/>
			<Input
				style={{marginTop: 24}}
				onChange={formik.handleChange("passcode")}
				prefix={<UserOutlined/>}
				placeholder={i18next.t("two-factor:Passcode")}
				type="text"
			/>
			<Button style={{marginTop: 24}} loading={formik.isSubmitting} block
					type="primary"
					htmlType="submit">
				{i18next.t("two-factor:Next step")}
			</Button>
		</form>
	);
}

function EnableTotp({user, totp, onSuccess, onFail}) {
	const [loading, setLoading] = useState(false);
	const requestEnableTotp = () => {
		const data = {
			userId: user.owner + "/" + user.name,
			secret: totp.secret,
			recoveryCode: totp.recoveryCode
		};
		setLoading(true);
		UserBackend.twoFactorEnableTotp(data).then(res => {
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
			<p>{i18next.t(
				"two-factor:Please save this recovery code. Once your device cannot provide an authentication code, you can reset two-factor authentication by this recovery code")}</p>
			<br/>
			<code style={{fontStyle: 'solid'}}>{totp.recoveryCode}</code>
			<Button style={{marginTop: 24}} loading={loading} onClick={() => {
				requestEnableTotp();
			}} block type="primary">
				{i18next.t("two-factor:Enable")}
			</Button>
		</div>
	);
}

class TotpPage extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			account: props.account,
			owner: props.match.params.owner,
			organization: props.match.params.organization,
			organizationOwner: props.match.params.organizationOwner,
			userOwner: props.match.params.userOwner,
			userName: props.match.params.userName,
			application: null,
			current: 0,
			totp: null
		};
	}

	componentDidMount() {
		this.getApplication();
	}

	getApplication() {
		ApplicationBackend.getApplication(this.state.organizationOwner,
			this.state.organization).then(
			(application) => {
				this.setState({
					application: application
				});
			}
		);
	}

	getUser() {
		return {
			name: this.state.userName,
			owner: this.state.userOwner
		};
	}

	getUserId() {
		return this.state.userOwner + "/" + this.state.userName;
	}

	renderStep() {
		switch (this.state.current) {
			case 0:
				return <CheckPassword
					user={this.getUser()}
					onSuccess={() => {
						UserBackend.twoFactorSetupInitTotp({
							userId: this.getUserId()
						}).then((res) => {
							if (res.status === "ok") {
								this.setState({
									totp: res.data,
									current: this.state.current + 1
								});
							} else {
								Setting.showMessage("error",
									i18next.t(`signup:${res.msg}`));
							}
						});
					}}
					onFail={(res) => {
						Setting.showMessage("error",
							i18next.t(`signup:${res.msg}`));
					}}
				/>;
			case 1:
				return <VerityTotp
					totp={this.state?.totp}
					onSuccess={() => {
						this.setState({
							current: this.state.current + 1
						});
					}}
					onFail={(res) => {
						Setting.showMessage("error",
							i18next.t(`signup:${res.msg}`));
					}}
				/>;
			case 2:
				return <EnableTotp
					user={this.getUser()}
					totp={this.state?.totp}
					onSuccess={() => {
						message.success(i18next.t('two-factor:Enabled successfully'))
						Setting.goToLinkSoft(this, "/account");
					}}
					onFail={(res) => {
						Setting.showMessage("error",
							i18next.t(`signup:${res.msg}`));
					}}
				/>;
		}
	}

	render() {
		const application = this.state.application;
		if (!application) {
			return <Spin/>;
		}

		return (
			<Row>
				<Col span={24} style={{justifyContent: "center"}}>
					<Row>
						<Col span={24}>
							<div style={{
								marginTop: "80px",
								marginBottom: "10px",
								textAlign: "center"
							}}>
								{
									Setting.renderHelmet(application)
								}
								<CustomGithubCorner/>
								{
									Setting.renderLogo(application)
								}
							</div>
						</Col>
					</Row>
					<Row>
						<Col span={24}>
							<div style={{textAlign: "center", fontSize: "28px"}}>
								{i18next.t(
									"two-factor:Protect your account with two-factor authentication")}
							</div>
							<div style={{
								textAlign: "center",
								fontSize: "16px",
								marginTop: "10px"
							}}>
								{i18next.t(
									"two-factor:Each time you sign in to your Account, you'll need your password and a authentication code")}
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
									marginTop: "80px"
								}}
							>
								<Step
									title={i18next.t("two-factor:Verify Password")}
									icon={<UserOutlined/>}
								/>
								<Step
									title={i18next.t("two-factor:Verify Code")}
									icon={<KeyOutlined/>}
								/>
								<Step
									title={i18next.t("two-factor:Enable")}
									icon={<CheckOutlined/>}
								/>
							</Steps>
						</Col>
					</Row>
				</Col>
				<Col span={24} style={{ display: "flex", justifyContent: "center" }}>
					<div style={{ marginTop: "10px", textAlign: "center" }}>
						{this.renderStep()}
					</div>
				</Col>
			</Row>
		);
	}
}

export default TotpPage;
