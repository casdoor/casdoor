import i18next from "i18next";

export function isValidOption_atLeast8(password) {
  if (password.length < 8) {
    // return "AtLeast8";
    return i18next.t("user:AtLeast8");
  }
  return "";
}

export function isValidOption_Aa123(password) {
  const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9]).+$/;
  if (!regex.test(password)) {
    // return "Aa123";
    return i18next.t("user:Aa123");
  }
  return "";
}

export function isValidOption_SpecialChar(password) {
  const regex = /^(?=.*[!@#$%^&*]).+$/;
  if (!regex.test(password)) {
    // return "SpecialChar";
    return i18next.t("user:SpecialChar");
  }
  return "";
}

export function isValidOption_noRepeat(password) {
  const regex = /(.)\1+/;
  if (regex.test(password)) {
    // return "NoRepeat";
    return i18next.t("user:NoRepeat");
  }
  return "";
}
