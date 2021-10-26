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
import {createButton} from "react-social-login-buttons";

class SelfLoginButton extends React.Component {
    generateIcon() {
        const avatar = this.props.account.avatar;
        return () => {
            return <img width={36} height={36} src={avatar} alt="Sign in with Google"/>;
        };
    }

    render() {
        const config = {
            icon: this.generateIcon(),
            iconFormat: name => `fa fa-${name}`,
            style: {background: "#ffffff", color: "#000000"},
            activeStyle: {background: "#eff0ee"},
        };

        const SelfLoginButton = createButton(config);
        return <SelfLoginButton text={`${this.props.account.name} (${this.props.account.displayName})`} onClick={() => this.props.onClick()} align={"center"} />
    }
}

export default SelfLoginButton;
