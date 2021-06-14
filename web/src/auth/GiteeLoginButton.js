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
    return <svg className="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="26" height="26"><path d="M512 1024C229.233778 1024 0 794.766222 0 512S229.233778 0 512 0s512 229.233778 512 512-229.233778 512-512 512z m259.157333-568.888889l-290.759111 0.014222c-13.966222 0-25.287111 11.320889-25.287111 25.272889l-0.028444 63.217778c0 13.966222 11.306667 25.287111 25.272889 25.287111h177.024c13.966222 0 25.287111 11.306667 25.287111 25.272889v12.643556A75.847111 75.847111 0 0 1 606.819556 682.666667h-240.213334a25.287111 25.287111 0 0 1-25.287111-25.272889V417.194667a75.847111 75.847111 0 0 1 75.847111-75.847111L771.086222 341.333333c13.966222 0 25.272889-11.306667 25.287111-25.272889L796.444444 252.871111c0-13.966222-11.306667-25.287111-25.272888-25.301333l-353.991112 0.014222C312.462222 227.569778 227.555556 312.476444 227.555556 417.194667v353.962666c0 13.966222 11.320889 25.287111 25.287111 25.287111H625.777778c94.264889 0 170.666667-76.401778 170.666666-170.666666V480.398222c0-13.952-11.320889-25.272889-25.287111-25.272889z" fill="#ffffff" /></svg>;
}

const config = {
    text: "Sign in with Gitee",
    icon: Icon,
    iconFormat: name => `fa fa-${name}`,
    style: {background: "#c71d23"},
    activeStyle: {background: "#f01130"},
};

const GiteeLoginButton = createButton(config);

export default GiteeLoginButton;
