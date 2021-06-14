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

import {createButton} from "react-social-login-buttons";

function Icon({ width = 24, height = 24, color }) {
    return <svg className="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="26" height="26"><path d="M164.565333 50.730667C216.32 70.954667 281.6 102.4 403.072 154.197333c157.482667 67.498667 319.445333 116.992 402.688 150.741334 71.978667 29.226667 98.986667 71.978667 83.242667 105.728-18.005333 40.490667-69.76 132.736-155.221334 272.213333h121.472l-233.984 305.962667 51.754667-206.976h-94.506667c15.744-67.498667 27.008-112.469333 33.749334-137.216a674.688 674.688 0 0 1-128.213334 31.488A230.101333 230.101333 0 0 1 366.933333 610.901333c-36.010667-33.749333-42.666667-58.496-24.746666-67.498666a1780.48 1780.48 0 0 1 231.722666-36.010667H344.576c-76.501333 0-125.994667-112.469333-137.216-144-13.482667-38.4 0-42.666667 13.482667-40.490667 15.744 2.261333 130.474667 27.008 341.973333 74.24a2442.453333 2442.453333 0 0 1-350.933333-123.733333C198.4 266.666667 157.909333 178.901333 148.906667 79.914667c-2.261333-9.002667 2.261333-33.749333 15.744-29.226667z" fill="#ffffff" /></svg>;
}

const config = {
    text: "Sign in with DingTalk",
    icon: Icon,
    iconFormat: name => `fa fa-${name}`,
    style: {background: "#0191e0"},
    activeStyle: {background: "rgb(76,143,208)"},
};

const DingTalkLoginButton = createButton(config);

export default DingTalkLoginButton;
