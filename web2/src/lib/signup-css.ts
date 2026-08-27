// Default custom CSS attached to each signup item, ported from web/src/table/SignupTable.js.

const EmailCss = ".signup-email{}\n.signup-email-input{}\n.signup-email-code{}\n.signup-email-code-input{}\n";
const PhoneCss = ".signup-phone{}\n.signup-phone-input{}\n.phone-code{}\n.signup-phone-code-input{}";

export const SignupTableDefaultCssMap: Record<string, string> = {
  "Username": ".signup-username {}\n.signup-username-input {}",
  "Display name": ".signup-first-name {}\n.signup-first-name-input{}\n.signup-last-name{}\n.signup-last-name-input{}\n.signup-name{}\n.signup-name-input{}",
  "Affiliation": ".signup-affiliation{}\n.signup-affiliation-input{}",
  "Country/Region": ".signup-country-region{}\n.signup-region-select{}",
  "ID card": ".signup-idcard{}\n.signup-idcard-input{}",
  "Password": ".signup-password{}\n.signup-password-input{}",
  "Confirm password": ".signup-confirm{}",
  "Email": EmailCss,
  "Phone": PhoneCss,
  "Email or Phone": EmailCss + PhoneCss,
  "Phone or Email": EmailCss + PhoneCss,
  "Invitation code": ".signup-invitation-code{}\n.signup-invitation-code-input{}",
  "Agreement": ".login-agreement{}",
  "Signup button": ".signup-button{}\n.signup-link{}",
  "Providers": ".provider-img {\n width: 30px;\n margin: 5px;\n }\n .provider-big-img {\n margin-bottom: 10px;\n }\n ",
  "Languages": ".signup-languages {\n    top: 55px;\n    right: 5px;\n    position: absolute;\n}",
};
