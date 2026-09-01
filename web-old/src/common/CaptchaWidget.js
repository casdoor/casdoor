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

import React, {useEffect} from "react";

export const CaptchaWidget = (props) => {
  const {captchaType, subType, siteKey, clientSecret, clientId2, clientSecret2, onChange} = props;

  const loadScript = (src) => {
    const tag = document.createElement("script");
    tag.async = false;
    tag.src = src;
    const body = document.getElementsByTagName("body")[0];
    body.appendChild(tag);
  };

  useEffect(() => {
    switch (captchaType) {
    case "reCAPTCHA" :
    case "reCAPTCHA v2": {
      const reTimer = setInterval(() => {
        if (!window.grecaptcha) {
          loadScript("https://recaptcha.net/recaptcha/api.js");
        }
        if (window.grecaptcha && window.grecaptcha.render) {
          window.grecaptcha.render("captcha", {
            sitekey: siteKey,
            callback: onChange,
          });
          clearInterval(reTimer);
        }
      }, 300);
      break;
    }
    case "reCAPTCHA v3": {
      const reTimer = setInterval(() => {
        if (!window.grecaptcha) {
          loadScript(`https://recaptcha.net/recaptcha/api.js?render=${siteKey}`);
        }
        if (window.grecaptcha && window.grecaptcha.render) {
          const clientId = window.grecaptcha.render("captcha", {
            "sitekey": siteKey,
            "badge": "inline",
            "size": "invisible",
            "callback": onChange,
            "error-callback": function() {
              const logoWidth = `${document.getElementById("captcha").offsetWidth + 40}px`;
              document.getElementsByClassName("grecaptcha-logo")[0].firstChild.style.width = logoWidth;
              document.getElementsByClassName("grecaptcha-badge")[0].style.width = logoWidth;
            },
          });

          window.grecaptcha.ready(function() {
            window.grecaptcha.execute(clientId, {action: "submit"});
          });
          clearInterval(reTimer);
        }
      }, 300);
      break;
    }
    case "hCaptcha": {
      const hTimer = setInterval(() => {
        if (!window.hcaptcha) {
          loadScript("https://js.hcaptcha.com/1/api.js");
        }
        if (window.hcaptcha && window.hcaptcha.render) {
          window.hcaptcha.render("captcha", {
            sitekey: siteKey,
            callback: onChange,
          });
          clearInterval(hTimer);
        }
      }, 300);
      break;
    }
    case "Aliyun Captcha": {
      window.AliyunCaptchaConfig = {
        region: "cn",
        prefix: clientSecret2,
      };

      const AWSCTimer = setInterval(() => {
        if (!window.initAliyunCaptcha) {
          loadScript("https://o.alicdn.com/captcha-frontend/aliyunCaptcha/AliyunCaptcha.js");
        }

        if (window.initAliyunCaptcha) {
          if (clientSecret2 && clientSecret2 !== "***") {
            window.initAliyunCaptcha({
              SceneId: clientId2,
              mode: "embed",
              element: "#captcha",
              captchaVerifyCallback: (data) => {
                onChange(data.toString());
              },
              slideStyle: {
                width: 320,
                height: 40,
              },
              language: "cn",
              immediate: true,
            });
          }
          clearInterval(AWSCTimer);
        }
      }, 300);
      break;
    }
    case "GEETEST": {
      let getLock = false;
      const gTimer = setInterval(() => {
        if (!window.initGeetest4) {
          loadScript("https://static.geetest.com/v4/gt4.js");
        }
        if (window.initGeetest4 && siteKey && !getLock) {
          const captchaId = String(siteKey);
          window.initGeetest4({
            captchaId,
            product: "float",
          }, function(captchaObj) {
            if (!getLock) {
              captchaObj.appendTo("#captcha");
              getLock = true;
            }
            captchaObj.onSuccess(function() {
              const result = captchaObj.getValidate();
              onChange(`lot_number=${result.lot_number}&captcha_output=${result.captcha_output}&pass_token=${result.pass_token}&gen_time=${result.gen_time}&captcha_id=${siteKey}`);
            });
          });
          clearInterval(gTimer);
        }
      }, 500);
      break;
    }
    case "Cloudflare Turnstile": {
      const tTimer = setInterval(() => {
        if (!window.turnstile) {
          loadScript("https://challenges.cloudflare.com/turnstile/v0/api.js");
        }
        if (window.turnstile && window.turnstile.render) {
          window.turnstile.render("#captcha", {
            sitekey: siteKey,
            callback: onChange,
          });
          clearInterval(tTimer);
        }
      }, 300);
      break;
    }
    default:
      break;
    }
  }, [captchaType, subType, siteKey, clientSecret, clientId2, clientSecret2]);

  return <div id="captcha" />;
};
