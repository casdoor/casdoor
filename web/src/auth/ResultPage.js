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
import {Link} from "react-router-dom";
import {Result, Button} from 'antd';
import i18next from "i18next";

class ResultPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  render() {
    return (
      <div>
        <Result
          status="success"
          title={i18next.t("signup:Your account has been created!")}
          subTitle={i18next.t("signup:Please click the below button to sign in")}
          extra={[
            <Link to="/login">
              <Button type="primary" key="login">
                {i18next.t("login:Sign In")}
              </Button>
            </Link>
          ]}
        />
      </div>
    );
  }
}

export default ResultPage;
