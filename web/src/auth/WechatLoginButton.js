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
    return <svg className="icon" viewBox="0 0 1025 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="26" height="26"><path d="M464.116992 442.88c71.68-74.24 161.28-102.4 263.68-94.72-10.24-48.64-33.28-92.16-66.56-130.56C528.116992 76.8 292.596992 51.2 133.876992 163.84 44.276992 225.28-6.923008 309.76 0.756992 422.4c5.12 89.6 53.76 156.16 122.88 207.36 17.92 12.8 20.48 25.6 12.8 46.08-10.24 25.6-17.92 48.64-28.16 79.36 38.4-17.92 69.12-35.84 102.4-51.2 12.8-5.12 30.72-7.68 46.08-5.12 43.52 5.12 89.6 10.24 135.68 15.36C371.956992 609.28 392.436992 517.12 464.116992 442.88zM494.836992 240.64c30.72-2.56 58.88 25.6 58.88 56.32 0 30.72-25.6 56.32-56.32 56.32-30.72 0-56.32-25.6-58.88-53.76C438.516992 268.8 466.676992 240.64 494.836992 240.64zM297.716992 302.08c-2.56 30.72-28.16 53.76-58.88 53.76-30.72-2.56-56.32-28.16-53.76-58.88 2.56-30.72 28.16-53.76 58.88-53.76C274.676992 243.2 300.276992 271.36 297.716992 302.08z" fill="#ffffff" /><path d="M950.516992 463.36c-112.64-110.08-294.4-125.44-427.52-35.84-148.48 99.84-156.16 294.4-12.8 404.48 81.92 61.44 176.64 79.36 279.04 56.32 25.6-5.12 46.08-5.12 69.12 10.24 20.48 12.8 40.96 23.04 66.56 35.84-2.56-12.8-5.12-20.48-7.68-25.6-20.48-43.52-12.8-71.68 25.6-104.96C1050.356992 701.44 1050.356992 560.64 950.516992 463.36zM622.836992 591.36c-20.48 0-40.96-17.92-40.96-38.4-2.56-20.48 17.92-43.52 40.96-43.52 20.48 0 40.96 17.92 40.96 38.4C661.236992 570.88 643.316992 591.36 622.836992 591.36zM819.956992 593.92c-23.04 0-38.4-17.92-38.4-40.96 0-23.04 17.92-38.4 38.4-38.4 23.04 0 40.96 17.92 40.96 38.4C860.916992 576 842.996992 593.92 819.956992 593.92z" fill="#ffffff" /></svg>;
}

const config = {
    text: "Sign in with Wechat",
    icon: Icon,
    iconFormat: name => `fa fa-${name}`,
    style: {background: "#06d30a"},
    activeStyle: {background: "#05c109"},
};

const WechatLoginButton = createButton(config);

export default WechatLoginButton;
