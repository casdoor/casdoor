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

import { Input } from "antd";
import React from "react";
import * as Setting from "../Setting";
import i18next from "i18next";

export const CountDownInput = (props) => {
  const {defaultButtonText, textBefore, placeHolder, onChange, onButtonClick, coolDownTime} = props;
  const [buttonText, setButtonText] = React.useState(defaultButtonText);
  let coolDown = false;

  const countDown = (leftTime) => {
    if (leftTime === 0) {
      coolDown = false;
      setButtonText(defaultButtonText);
      return;
    }
    setButtonText(`${leftTime} s`);
    setTimeout(() => countDown(leftTime - 1), 1000);
  }

  const clickButton = () => {
    if (coolDown) {
      Setting.showMessage("error", i18next.t("general:Cooling down"));
      return;
    }
    onButtonClick();
    coolDown = true;
    countDown(coolDownTime);
  }

  return (
    <Input addonBefore={textBefore} placeholder={placeHolder} onChange={e => onChange(e.target.value)} addonAfter={<button onClick={clickButton} style={{backgroundColor: "#fafafa", border: "none"}}>{buttonText}</button>}/>
  );
}