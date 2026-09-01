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
import {initReactI18next} from "react-i18next";
import * as Conf from "./Conf";
import en from "./locales/en/data.json";

// Load backend-provided frontend config before language detection runs.
Conf.initConfigFromCookie();

export const SupportedLanguages = [
  "en", "zh", "es", "fr", "de", "id", "ja", "ko", "ru", "vi", "pt", "it", "ms",
  "tr", "ar", "he", "nl", "pl", "fi", "sv", "uk", "kk", "fa", "cs", "sk", "az",
];

// Languages that actually ship a translation bundle in src/locales.
const bundledLanguages = ["de", "en", "es", "fr", "ja", "pl", "pt", "tr", "uk", "vi", "zh"];

const localeLoaders = import.meta.glob("./locales/*/data.json");

function initLanguage(): string {
  let language = localStorage.getItem("language");
  if (language === undefined || language === null || language === "") {
    if (Conf.ForceLanguage !== "") {
      language = Conf.ForceLanguage;
    } else {
      const baseLanguage = navigator.language.split("-")[0];
      language = SupportedLanguages.includes(baseLanguage) ? baseLanguage : Conf.DefaultLanguage;
    }
  }
  return language;
}

const resourcesToBackend = () => ({
  type: "backend" as const,
  init() {/* no options needed */},
  read(language: string, namespace: string, callback: (err: any, data?: any) => void) {
    const lang = bundledLanguages.includes(language) ? language : "en";
    const loader = localeLoaders[`./locales/${lang}/data.json`];
    if (!loader) {
      callback(null, (en as any)[namespace]);
      return;
    }
    loader()
      .then((res: any) => callback(null, (res.default || res)[namespace]))
      .catch(callback);
  },
});

i18n
  .use(resourcesToBackend() as any)
  .use(initReactI18next)
  .init({
    lng: initLanguage(),
    ns: Object.keys(en),
    fallbackLng: "en",
    keySeparator: false,
    // Must be set explicitly: with only `keySeparator: false`, i18next treats a key
    // containing spaces as natural language and stops splitting off the "general:"
    // namespace, so every multi-word key would resolve to itself.
    nsSeparator: ":",
    interpolation: {
      escapeValue: true,
    },
    react: {
      useSuspense: false,
    },
  });

export default i18n;
