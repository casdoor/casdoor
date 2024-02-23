import * as phoneNumber from "libphonenumber-js";
import {getLanguage} from "./Setting";

export function initCountries() {
  const countries = require("i18n-iso-countries");
  countries.registerLocale(require("i18n-iso-countries/langs/" + getLanguage() + ".json"));
  return countries;
}

export function getCountryCode(country) {
  if (phoneNumber.isSupportedCountry(country)) {
    return phoneNumber.getCountryCallingCode(country);
  }
  return "";
}

export function getCountryCodeData(countryCodes = phoneNumber.getCountries()) {
  return countryCodes?.map((countryCode) => {
    if (phoneNumber.isSupportedCountry(countryCode)) {
      const name = initCountries().getName(countryCode, getLanguage());
      return {
        code: countryCode,
        name: name || "",
        phone: phoneNumber.getCountryCallingCode(countryCode),
      };
    }
  }).filter(item => item.name !== "")
    .sort((a, b) => a.phone - b.phone);
}
