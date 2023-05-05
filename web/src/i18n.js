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
import resourcesToBackend from "i18next-resources-to-backend";
function initLanguage() {
  let language = localStorage.getItem("language");
  if (language === undefined || language === null || language === "") {
    if (Conf.ForceLanguage !== "") {
      language = Conf.ForceLanguage;
    } else {
      const userLanguage = navigator.language;
      switch (userLanguage) {
      case "en":
        language = "en";
        break;
      case "en-US":
        language = "en";
        break;
      case "zh-CN":
        language = "zh";
        break;
      case "zh":
        language = "zh";
        break;
      case "es":
        language = "es";
        break;
      case "fr":
        language = "fr";
        break;
      case "de":
        language = "de";
        break;
      case "id":
        language = "id";
        break;
      case "ja":
        language = "ja";
        break;
      case "ko":
        language = "ko";
        break;
      case "ru":
        language = "ru";
        break;
      case "vi":
        language = "vi";
        break;
      default:
        language = Conf.DefaultLanguage;
      }
    }
  }

  return language;
}

i18n.use(resourcesToBackend((language, namespace) => import(`./locales/${language}/${namespace}.json`)))
  .use(initReactI18next).init({
    lng: initLanguage(),

    keySeparator: false,

    interpolation: {
      escapeValue: true,
    },
    debug: true,
    saveMissing: true,
  });

export default i18n;
