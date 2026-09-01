import * as React from "react";

interface CaptchaWidgetProps {
  captchaType: string;
  subType?: string;
  siteKey?: string;
  clientSecret?: string;
  clientId2?: string;
  clientSecret2?: string;
  onChange: (token: string) => void;
}

function loadScript(src: string) {
  const tag = document.createElement("script");
  tag.async = false;
  tag.src = src;
  document.getElementsByTagName("body")[0].appendChild(tag);
}

/**
 * Mounts the third-party captcha widget the application is configured with.
 * Ported from web/src/common/CaptchaWidget.js — the vendor scripts, the element
 * id ("captcha") and the token formats are unchanged, so the backend keeps
 * validating exactly what it validated before.
 */
export function CaptchaWidget({
  captchaType,
  subType,
  siteKey,
  clientSecret,
  clientId2,
  clientSecret2,
  onChange,
}: CaptchaWidgetProps) {
  const onChangeRef = React.useRef(onChange);
  onChangeRef.current = onChange;

  React.useEffect(() => {
    const emit = (token: string) => onChangeRef.current(token);
    let timer: number | undefined;

    switch (captchaType) {
    case "reCAPTCHA":
    case "reCAPTCHA v2": {
      timer = window.setInterval(() => {
        if (!(window as any).grecaptcha) {
          loadScript("https://recaptcha.net/recaptcha/api.js");
        }
        if ((window as any).grecaptcha?.render) {
          (window as any).grecaptcha.render("captcha", {sitekey: siteKey, callback: emit});
          window.clearInterval(timer);
        }
      }, 300);
      break;
    }
    case "reCAPTCHA v3": {
      timer = window.setInterval(() => {
        if (!(window as any).grecaptcha) {
          loadScript(`https://recaptcha.net/recaptcha/api.js?render=${siteKey}`);
        }
        if ((window as any).grecaptcha?.render) {
          const clientId = (window as any).grecaptcha.render("captcha", {
            "sitekey": siteKey,
            "badge": "inline",
            "size": "invisible",
            "callback": emit,
            "error-callback": function() {
              const element = document.getElementById("captcha");
              if (!element) {
                return;
              }
              const logoWidth = `${element.offsetWidth + 40}px`;
              const logo = document.getElementsByClassName("grecaptcha-logo")[0] as HTMLElement | undefined;
              const badge = document.getElementsByClassName("grecaptcha-badge")[0] as HTMLElement | undefined;
              if (logo?.firstChild) {
                (logo.firstChild as HTMLElement).style.width = logoWidth;
              }
              if (badge) {
                badge.style.width = logoWidth;
              }
            },
          });
          (window as any).grecaptcha.ready(function() {
            (window as any).grecaptcha.execute(clientId, {action: "submit"});
          });
          window.clearInterval(timer);
        }
      }, 300);
      break;
    }
    case "hCaptcha": {
      timer = window.setInterval(() => {
        if (!(window as any).hcaptcha) {
          loadScript("https://js.hcaptcha.com/1/api.js");
        }
        if ((window as any).hcaptcha?.render) {
          (window as any).hcaptcha.render("captcha", {sitekey: siteKey, callback: emit});
          window.clearInterval(timer);
        }
      }, 300);
      break;
    }
    case "Aliyun Captcha": {
      (window as any).AliyunCaptchaConfig = {region: "cn", prefix: clientSecret2};
      timer = window.setInterval(() => {
        if (!(window as any).initAliyunCaptcha) {
          loadScript("https://o.alicdn.com/captcha-frontend/aliyunCaptcha/AliyunCaptcha.js");
        }
        if ((window as any).initAliyunCaptcha) {
          if (clientSecret2 && clientSecret2 !== "***") {
            (window as any).initAliyunCaptcha({
              SceneId: clientId2,
              mode: "embed",
              element: "#captcha",
              captchaVerifyCallback: (data: any) => emit(data.toString()),
              slideStyle: {width: 320, height: 40},
              language: "cn",
              immediate: true,
            });
          }
          window.clearInterval(timer);
        }
      }, 300);
      break;
    }
    case "GEETEST": {
      let getLock = false;
      timer = window.setInterval(() => {
        if (!(window as any).initGeetest4) {
          loadScript("https://static.geetest.com/v4/gt4.js");
        }
        if ((window as any).initGeetest4 && siteKey && !getLock) {
          (window as any).initGeetest4({captchaId: String(siteKey), product: "float"}, function(captchaObj: any) {
            if (!getLock) {
              captchaObj.appendTo("#captcha");
              getLock = true;
            }
            captchaObj.onSuccess(function() {
              const result = captchaObj.getValidate();
              emit(
                `lot_number=${result.lot_number}&captcha_output=${result.captcha_output}&pass_token=${result.pass_token}&gen_time=${result.gen_time}&captcha_id=${siteKey}`,
              );
            });
          });
          window.clearInterval(timer);
        }
      }, 500);
      break;
    }
    case "Cloudflare Turnstile": {
      timer = window.setInterval(() => {
        if (!(window as any).turnstile) {
          loadScript("https://challenges.cloudflare.com/turnstile/v0/api.js");
        }
        if ((window as any).turnstile?.render) {
          (window as any).turnstile.render("#captcha", {sitekey: siteKey, callback: emit});
          window.clearInterval(timer);
        }
      }, 300);
      break;
    }
    default:
      break;
    }

    return () => {
      if (timer !== undefined) {
        window.clearInterval(timer);
      }
    };
  }, [captchaType, subType, siteKey, clientSecret, clientId2, clientSecret2]);

  return <div id="captcha" />;
}
