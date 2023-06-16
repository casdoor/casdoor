import i18next from "i18next";

function isValidOption_atLeast8(password) {
  if (password.length === 0) {
    return i18next.t("user:InputPassword");
  } else if (password.length < 8) {
    // return "AtLeast8";
    return i18next.t("user:AtLeast8");
  }
  return "";
}

function isValidOption_atLeast6(password) {
  if (password.length === 0) {
    return i18next.t("user:InputPassword");
  } else if (password.length < 6) {
    return i18next.t("user:AtLeast6");
  }
  return "";
}

function isValidOption_Aa123(password) {
  const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9]).+$/;
  if (!regex.test(password)) {
    // return "Aa123";
    return i18next.t("user:Aa123");
  }
  return "";
}

function isValidOption_SpecialChar(password) {
  const regex = /^(?=.*[!@#$%^&*]).+$/;
  if (!regex.test(password)) {
    // return "SpecialChar";
    return i18next.t("user:SpecialChar");
  }
  return "";
}

function isValidOption_noRepeat(password) {
  const regex = /(.)\1+/;
  if (regex.test(password)) {
    // return "NoRepeat";
    return i18next.t("user:NoRepeat");
  }
  return "";
}

export function checkPasswordComplexOption(password, complexOptions) {
  /*
    AtLeast8: The password length must be greater than 8
    Aa123: The password must contain at least one lowercase letter, one uppercase letter, and one digit
    SpecialChar: The password must contain at least one special character
    NoRepeat: The password must not contain any repeated characters
  */
  if (complexOptions.length < 1) {
    return "";
  }

  const validators = {
    AtLeast8: isValidOption_atLeast8,
    AtLeast6: isValidOption_atLeast6,
    Aa123: isValidOption_Aa123,
    SpecialChar: isValidOption_SpecialChar,
    NoRepeat: isValidOption_noRepeat,
  };
  for (const option of complexOptions) {
    const validateFunc = validators[option];
    if (validateFunc) {
      const msg = validateFunc(password);
      if (msg !== "") {
        return msg;
      }
    } else {
      // Invalid complex option
      return i18next.t("user:InvalidOption");
    }
  }

  return "";
}
