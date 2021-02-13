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

import React from 'react';
import Face from "../Face";

class LoginPage extends React.Component {
  render() {
    return (
      <Face applicationName={"app-built-in"} account={this.props.account} onLoggedIn={this.props.onLoggedIn.bind(this)} {...this.props} />
    )
  }
}

export default LoginPage;
