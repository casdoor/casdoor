// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import i18next from "i18next";
import {createButton} from "react-social-login-buttons";

function LoginButton({type, logoUrl, align = "center", style = {background: "#ffffff", color: "#000000"}, activeStyle = {background: "#ededee"}}) {
  function Icon({width = 24, height = 24, color}) {
    return <img src={logoUrl} alt={`Sign in with ${type}`} style={{width: width, height: height}} />;
  }
  const config = {
    text: `Sign in with ${type}`,
    icon: Icon,
    style: style,
    activeStyle: activeStyle,
  };
  const Button = createButton(config);
  const text = i18next.t("login:Sign in with {type}").replace("{type}", type);
  return <Button text={text} align={align} />;
}

export default LoginButton;
