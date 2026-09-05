import i18next from "i18next";
import {checkPasswordComplexity} from "@/lib/password-checker";
import * as Setting from "@/lib/setting";

/**
 * The client-side checks the antd signup form runs as antd Form rules. Ported
 * here so the same organization configuration is enforced — in particular each
 * signup item's own `regex`, which the form otherwise silently ignores.
 *
 * Returns a message for the first failing field, keyed by the form field name,
 * or an empty object when everything passes.
 */

/**
 * The field of `values` a signup item edits, and the json key the backend reads
 * it from (`form.AuthForm`). Lower-casing the item name is not enough: "ID card"
 * has to become `idCard`, not `iDcard`, or the value never arrives.
 */
const SIGNUP_ITEM_FIELDS: Record<string, string> = {
  "Username": "username",
  "Display name": "name",
  "Password": "password",
  "Confirm password": "confirm",
  "Email": "email",
  "Email or Phone": "email",
  "Phone or Email": "email",
  "Phone": "phone",
  "Country/Region": "region",
  "ID card": "idCard",
  "First name": "firstName",
  "Last name": "lastName",
  "Invitation code": "invitationCode",
  "Affiliation": "affiliation",
  "Bio": "bio",
  "Tag": "tag",
  "Education": "education",
  "Gender": "gender",
};

export function getSignupItemField(itemName: string): string {
  return (
    SIGNUP_ITEM_FIELDS[itemName] ??
    itemName.replace(/\s+/g, "").replace(/^./, (c) => c.toLowerCase())
  );
}

const ID_CARD_REGEX =
  /^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9X]$/;

function requiredMessage(item: any): string {
  return i18next.t("signup:Please input your {label}!").replace("{label}", item.label || item.name);
}

function isBlank(value: any): boolean {
  if (Array.isArray(value)) {
    return value.length === 0;
  }
  return value === undefined || value === null || `${value}`.trim() === "";
}

/** One item's error message, or "" when it is fine. */
export function validateSignupItem(item: any, values: Record<string, any>, application: any): string {
  // a "Text N" item is decoration, it edits nothing
  if (Setting.isCustomFormItem(item)) {
    return "";
  }

  const field = getSignupItemField(item.name);
  const value = values[field];
  const required = !!item.required;

  switch (item.name) {
  case "Display name": {
    // the "First, last" rule replaces the single field with two of its own
    if (item.rule === "First, last" && Setting.getLanguage() !== "zh") {
      if (required && isBlank(values.firstName)) {
        return i18next.t("signup:Please input your first name!");
      }
      if (required && isBlank(values.lastName)) {
        return i18next.t("signup:Please input your last name!");
      }
      // the antd form put the item's regex on both halves of the name
      if (item.regex) {
        const regex = new RegExp(item.regex);
        if ((!isBlank(values.firstName) && !regex.test(values.firstName)) ||
            (!isBlank(values.lastName) && !regex.test(values.lastName))) {
          return i18next.t("signup:The input doesn't match the signup item regex!");
        }
      }
      return "";
    }
    if (required && isBlank(value)) {
      return item.rule === "Real name"
        ? i18next.t("signup:Please input your real name!")
        : i18next.t("signup:Please input your display name!");
    }
    if (!isBlank(value) && item.regex && !new RegExp(item.regex).test(value)) {
      return i18next.t("signup:The input doesn't match the signup item regex!");
    }
    return "";
  }
  case "Email":
  case "Email or Phone":
  case "Phone or Email": {
    if (required && isBlank(value)) {
      return i18next.t("login:Please input your Email!");
    }
    if (!isBlank(value) && !Setting.isValidEmail(value)) {
      return i18next.t("login:The input is not valid Email!");
    }
    if (!isBlank(value) && item.regex && !new RegExp(item.regex).test(value)) {
      return i18next.t("signup:The input Email doesn't match the signup item regex!");
    }
    if (item.rule !== "No verification" && required && isBlank(values.emailCode)) {
      return i18next.t("code:Please input your verification code!");
    }
    return "";
  }
  case "Phone": {
    if (required && isBlank(values.countryCode) && !application?.organizationObj?.countryCodes?.[0]) {
      return i18next.t("signup:Please select your country code!");
    }
    if (required && isBlank(values.phone)) {
      return i18next.t("signup:Please input your phone number!");
    }
    if (!isBlank(values.phone) && !Setting.isValidPhone(values.phone, values.countryCode)) {
      return i18next.t("signup:The input is not valid Phone!");
    }
    if (item.rule !== "No verification" && required && isBlank(values.phoneCode)) {
      return i18next.t("code:Please input your phone verification code!");
    }
    return "";
  }
  case "Password": {
    if (required && isBlank(value)) {
      return requiredMessage(item);
    }
    if (isBlank(value)) {
      return "";
    }
    return checkPasswordComplexity(value, application?.organizationObj?.passwordOptions);
  }
  case "Confirm password": {
    if (required && isBlank(values.confirm)) {
      return i18next.t("signup:Please confirm your password!");
    }
    if (!isBlank(values.confirm) && values.confirm !== values.password) {
      return i18next.t("signup:Your confirmed password is inconsistent with the password!");
    }
    return "";
  }
  case "ID card": {
    if (required && isBlank(value)) {
      return i18next.t("signup:Please input your ID card number!");
    }
    if (!isBlank(value) && !ID_CARD_REGEX.test(value)) {
      return i18next.t("signup:Please input the correct ID card number!");
    }
    return "";
  }
  case "Affiliation": {
    if (required && isBlank(value)) {
      return i18next.t("signup:Please input your affiliation!");
    }
    if (!isBlank(value) && item.regex && !new RegExp(item.regex).test(value)) {
      return i18next.t("signup:The input doesn't match the signup item regex!");
    }
    return "";
  }
  case "Country/Region":
    return required && isBlank(value) ? i18next.t("signup:Please select your country/region!") : "";
  case "Tag":
    return required && isBlank(value) ? i18next.t("signup:Please select your tag!") : "";
  case "Invitation code":
    return required && isBlank(value) ? i18next.t("signup:Please input your invitation code!") : "";
  case "Username": {
    if (required && isBlank(value)) {
      return i18next.t("forget:Please input your username!");
    }
    if (!isBlank(value) && item.regex && !new RegExp(item.regex).test(value)) {
      return i18next.t("signup:The input doesn't match the signup item regex!");
    }
    return "";
  }
  case "Agreement":
    return required && values.agreement !== true ? i18next.t("signup:Please accept the agreement!") : "";
  // these render no input of their own
  case "ID":
  case "Languages":
  case "Providers":
  case "Signup button":
  case "Text 1":
  case "Text 2":
  case "Text 3":
  case "Text 4":
  case "Text 5":
    return "";
  default: {
    if (required && isBlank(value)) {
      return requiredMessage(item);
    }
    // a plain input honours the item's own regex, as the antd form does
    if (!isBlank(value) && item.regex && !new RegExp(item.regex).test(value)) {
      return i18next.t("signup:The input doesn't match the signup item regex!");
    }
    return "";
  }
  }
}

/** Every visible item's error, keyed by the field it edits. */
export function validateSignupItems(
  items: any[],
  values: Record<string, any>,
  application: any,
): Record<string, string> {
  const errors: Record<string, string> = {};
  (items ?? []).forEach((item) => {
    if (!item?.visible) {
      return;
    }
    const message = validateSignupItem(item, values, application);
    if (message) {
      errors[getSignupItemField(item.name)] = message;
    }
  });
  return errors;
}
