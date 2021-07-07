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

import React from 'react'
import * as Util from "./Util";
import * as AuthBackend from "./AuthBackend";
import * as Setting from "../Setting";

class LoginPageCustom extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            Html: props.Html,
            remember: false
        }
    }

    onFinish(values) {
        const oAuthParams = Util.getOAuthGetParameters();

        AuthBackend.login(values, oAuthParams)
            .then((res) => {
                if (res.status === 'ok') {
                    Util.showMessage("success", `Logged in successfully`);
                    Setting.goToLink("/");
                } else {
                    Util.showMessage("error", `Failed to log in: ${res.msg}`);
                }
            });
    };

    componentDidMount() {
        if (document.getElementById("login") && document.getElementById("username") &&
            document.getElementById("password") && document.getElementById("remember")) {

            document.getElementById("login").onclick = () => {
                const username = document.getElementById("username").value
                const password = document.getElementById("password").value

                var values = {
                    application: this.props.name,
                    organization: this.props.organization,
                    password: password,
                    username: username,
                    remember: this.state.remember,
                    type: "login"
                }

                this.onFinish(values)
            }

            document.getElementById("remember").onchange = () => {
                this.setState({
                    remember: document.getElementById("remember").checked
                })
            }
        }

        if (document.getElementById("forget") && document.getElementById("signup")) {

            document.getElementById("forget").onclick = () => {
                Setting.goToForget(this.props.Parent, this.props.Application)
            }

            document.getElementById("signup").onclick = () => {
                Setting.goToSignup(this.props.Parent, this.props.Application)
            }
        }
    }

    render() {
        return (
            <div>
                <div dangerouslySetInnerHTML={{__html: this.state.Html}}/>
            </div>
        );
    }
}

export default LoginPageCustom