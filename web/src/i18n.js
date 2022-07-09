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

import i18n from 'i18next';
import zh from './locales/zh/data.json';
import en from './locales/en/data.json';
import fr from './locales/fr/data.json';
import de from './locales/de/data.json';
import ko from './locales/ko/data.json';
import ru from './locales/ru/data.json';
import ja from './locales/ja/data.json';
import * as Conf from './Conf';
import * as Setting from './Setting';

const resources = {
  en: en,
  zh: zh,
  fr: fr,
  de: de,
  ko: ko,
  ru: ru,
  ja: ja,
};

function initLanguage() {
  let language = localStorage.getItem('language');
  if (language === undefined || language == null) {
    if (Conf.ForceLanguage !== '') {
      language = Conf.ForceLanguage;
    } else {
      let userLanguage;
      userLanguage = navigator.language;
      switch (userLanguage) {
      case 'zh-CN':
        language = 'zh';
        break;
      case 'zh':
        language = 'zh';
        break;
      case 'en':
        language = 'en';
        break;
      case 'en-US':
        language = 'en';
        break;
      case 'fr':
        language = 'fr';
        break;
      case 'de':
        language = 'de';
        break;
      case 'ko':
        language = 'ko';
        break;
      case 'ru':
        language = 'ru';
        break;
      case 'ja':
        language = 'ja';
        break;
      default:
        language = Conf.DefaultLanguage;
      }
    }
  }
  Setting.changeMomentLanguage(language);

  return language;
}

i18n.init({
  lng: initLanguage(),

  resources: resources,

  keySeparator: false,

  interpolation: {
    escapeValue: false,
  },
  //debug: true,
  saveMissing: true,
});

export default i18n;
