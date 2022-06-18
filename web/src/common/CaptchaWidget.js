// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

import React, { useEffect } from "react";

export const CaptchaWidget = ({ captchaType, siteKey, onChange }) => {
  const loadScript = (src) => {
    var tag = document.createElement("script");
    tag.async = false;
    tag.src = src;
    var body = document.getElementsByTagName("body")[0];
    body.appendChild(tag);
  };

  useEffect(() => {
    switch (captchaType) {
      case "reCAPTCHA":
        const reTimer = setInterval(() => {
          if (!window.grecaptcha) {
            loadScript("https://recaptcha.net/recaptcha/api.js");
          }
          if (window.grecaptcha) {
            window.grecaptcha.render("captcha", {
              sitekey: siteKey,
              callback: onChange,
            });
            clearInterval(reTimer);
          }
        }, 300);
        break;
      case "hCaptcha":
        const hTimer = setInterval(() => {
          if (!window.hcaptcha) {
            loadScript("https://js.hcaptcha.com/1/api.js");
          }
          if (window.hcaptcha) {
            window.hcaptcha.render("captcha", {
              sitekey: siteKey,
              callback: onChange,
            });
            clearInterval(hTimer);
          }
        }, 300);
        break;
      default:
        break;
    }
  }, [captchaType, siteKey]);

  return <div id="captcha"></div>;
};
