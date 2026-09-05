// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {useLanguageVersion} from "@/hooks/use-language";
import * as Setting from "@/lib/setting";

/**
 * Phone calling code combobox: the trigger shows "+86", the dropdown shows the
 * flag and the country name so that the countries sharing a code stay apart.
 */
export function CountryCodeSelect({
  value,
  onChange,
  countryCodes,
  disabled,
  className,
}: {
  value?: string | null;
  onChange: (value: string) => void;
  countryCodes?: string[];
  disabled?: boolean;
  className?: string;
}) {
  // the country names are translated, so they have to be rebuilt on a language change
  const languageVersion = useLanguageVersion();
  const options = React.useMemo(
    () => Setting.getCountryCodeOptions(countryCodes),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [countryCodes?.join(","), languageVersion],
  );

  return (
    <SearchableSelect
      value={value ?? ""}
      onChange={onChange}
      options={options}
      disabled={disabled}
      className={className}
      contentClassName="w-[min(20rem,calc(100vw-2rem))]"
    />
  );
}
