// Default custom CSS attached to each signin item, ported from web/src/table/SigninTable.js.

export const SigninTableDefaultCssMap: Record<string, string> = {
  "Back button": ".back-button {\n      top: 65px;\n      left: 15px;\n      position: absolute;\n}\n.back-inner-button{}",
  "Languages": ".login-languages {\n    top: 55px;\n    right: 5px;\n    position: absolute;\n}",
  "Logo": ".login-logo-box {}",
  "Signin methods": ".signin-methods {}",
  "Username": ".login-username {}\n.login-username-input{}",
  "Password": ".login-password {}\n.login-password-input{}",
  "Verification code": ".verification-code {}\n.verification-code-input{}",
  "Agreement": ".login-agreement {}",
  "Forgot password?": ".login-forget-password {\n    display: inline-flex;\n    justify-content: space-between;\n    width: 320px;\n    margin-bottom: 25px;\n}",
  "Login button": ".login-button-box {\n    margin-bottom: 5px;\n}\n.login-button {\n    width: 100%;\n}",
  "Signup link": ".login-signup-link {\n    margin-bottom: 24px;\n    display: flex;\n    justify-content: end;\n}",
  "Providers": ".provider-link {\n      display: inline-block;\n      border-radius: 50%;\n}\n.provider-img {\n      width: 30px;\n      margin: 5px;\n      border-radius: 50%;\n      transition: filter 0.25s ease, opacity 0.25s ease;\n}\n.provider-link:hover .provider-img {\n      filter: drop-shadow(0 0 10px rgba(120,120,120,0.75)) drop-shadow(0 0 4px rgba(120,120,120,0.5)) brightness(1.1);\n}\n.provider-link:active .provider-img {\n      filter: drop-shadow(0 0 6px rgba(120,120,120,0.45)) brightness(0.9);\n      opacity: 0.8;\n      transition: filter 0.08s ease, opacity 0.08s ease;\n}\n.provider-big-img {\n      margin-bottom: 10px;\n}\n.provider-big-img > a {\n      display: block;\n      border-radius: 4px;\n      transition: filter 0.25s ease, opacity 0.25s ease;\n}\n.provider-big-img > a:hover {\n      filter: brightness(1.04) drop-shadow(0 0 14px rgba(120,120,120,0.5)) drop-shadow(0 0 5px rgba(120,120,120,0.3));\n}\n.provider-big-img > a:active {\n      filter: brightness(0.94);\n      opacity: 0.85;\n      transition: filter 0.08s ease, opacity 0.08s ease;\n}",
};
