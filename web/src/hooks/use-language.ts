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

import * as React from "react";
import i18next from "i18next";

// The screens call i18next.t() straight from their render instead of going through
// useTranslation(), so React has to be told when the bundle changed - otherwise the
// old strings stay on screen until the page is reloaded.
let version = 0;
const listeners = new Set<() => void>();

const bump = () => {
  version += 1;
  listeners.forEach((listener) => listener());
};

// "languageChanged" fires once the new bundle is loaded, "loaded" covers a namespace
// that arrives later through the lazy locale loader.
i18next.on("languageChanged", bump);
i18next.on("loaded", bump);

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function getVersion() {
  return version;
}

export function useLanguageVersion(): number {
  return React.useSyncExternalStore(subscribe, getVersion);
}

export function useLanguage(): string {
  useLanguageVersion();
  return i18next.language;
}
