// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

import i18n from "i18next";
import * as Conf from "./Conf";
import {initReactI18next} from "react-i18next";
import en from "./locales/en/data.json";

// Load backend-provided frontend config before language detection runs.
Conf.initConfigFromCookie();

const resourcesToBackend = (res) => ({
  type: "backend",
  init(services, backendOptions, i18nextOptions) {/* use services and options */},
  read(language, namespace, callback) {
    if (typeof res === "function") {
      if (res.length < 3) {
        try {
          const r = res(language, namespace);
          if (r && typeof r.then === "function") {
            r.then((data) => callback(null, (data && data.default) || data)).catch(callback);
          } else {
            callback(null, r);
          }
        } catch (err) {
          callback(err);
        }
        return;
      }
      res(language, namespace, callback);
      return;
    }
    callback(null, res && res[language] && res[language][namespace]);
  },
});

function initLanguage() {
  let language = localStorage.getItem("language");
  if (language === undefined || language === null || language === "") {
    if (Conf.ForceLanguage !== "") {
      language = Conf.ForceLanguage;
    } else {
      const supportedLanguages = ["en", "zh", "es", "fr", "de", "id", "ja", "ko", "ru", "vi", "pt", "it", "ms", "tr", "ar", "he", "nl", "pl", "fi", "sv", "uk", "kk", "fa", "cs", "sk", "az"];
      const baseLanguage = navigator.language.split("-")[0];
      language = supportedLanguages.includes(baseLanguage) ? baseLanguage : Conf.DefaultLanguage;
    }
  }

  return language;
}

i18n.use(resourcesToBackend(async(language, namespace) => {
  const res = await import(`./locales/${language}/data.json`);
  return res.default[namespace];
}
))
  .use(initReactI18next)
  .init({
    lng: initLanguage(),
    ns: Object.keys(en),
    fallbackLng: "en",

    keySeparator: false,

    interpolation: {
      escapeValue: true,
    },
    // debug: true,
    saveMissing: true,
  });
export default i18n;
