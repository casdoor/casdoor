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

import React, { useRef } from "react";
import {Button, Card, Col, Input, Row, Select, Switch} from 'antd';
import * as Setting from "./Setting";
import {LinkOutlined} from "@ant-design/icons";
import i18next from "i18next";
import CropperDiv from "./CropperDiv.tsx";

class UploadAvatarPage extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            account: props.account
        };
    }

    render() {
        return (<CropperDiv name={this.props.account.owner} password={this.props.account.password}/>)
    }
}

export default UploadAvatarPage;
