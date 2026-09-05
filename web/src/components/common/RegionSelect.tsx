import * as React from "react";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {useLanguageVersion} from "@/hooks/use-language";
import * as Setting from "@/lib/setting";

/**
 * Country/region combobox, the port of web/src/common/select/RegionSelect.js.
 * The stored value is the ISO country code, the label carries the flag so the
 * list reads the same as the antd one.
 */
export function RegionSelect({
  value,
  onChange,
  disabled,
  className,
}: {
  value?: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  className?: string;
}) {
  // the country names are translated, so they have to be rebuilt on a language change
  const languageVersion = useLanguageVersion();
  const options = React.useMemo(
    () =>
      Setting.getCountryCodeData()
        .map((item: any) => ({
          value: item.code,
          keywords: `${item.name} ${item.code}`,
          label: (
            <span className="flex items-center">
              {Setting.getCountryImage(item)}
              {`${item.name} (${item.code})`}
            </span>
          ),
        }))
        .sort((a: any, b: any) => a.keywords.localeCompare(b.keywords)),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [languageVersion],
  );

  return (
    <SearchableSelect
      value={value ?? ""}
      onChange={onChange}
      options={options}
      disabled={disabled}
      className={className}
      placeholder="Please select country/region"
    />
  );
}
