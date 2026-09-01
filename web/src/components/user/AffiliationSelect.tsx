import * as React from "react";
import i18next from "i18next";
import {Input} from "@/components/ui/input";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import * as UserBackend from "@/backend/UserBackend";

interface CascaderOption {
  value: string;
  label: string;
  children?: CascaderOption[];
}

interface AffiliationOption {
  id: number;
  name: string;
}

/**
 * When the application sets `affiliationUrl` ("<addressUrl>|<affiliationUrl>"),
 * the address becomes a cascader fed by that endpoint and the affiliation becomes
 * a select scoped to the picked address, which also carries the numeric `score`.
 * Port of web/src/common/select/AffiliationSelect.js.
 */
export function useAffiliation(application: any, user: any) {
  const affiliationUrl: string = application?.affiliationUrl ?? "";
  const enabled = affiliationUrl !== "";
  const [addressOptions, setAddressOptions] = React.useState<CascaderOption[]>([]);
  const [affiliationOptions, setAffiliationOptions] = React.useState<AffiliationOption[]>([]);

  React.useEffect(() => {
    if (!enabled) {
      setAddressOptions([]);
      return;
    }
    UserBackend.getAddressOptions(affiliationUrl.split("|")[0])
      .then((options) => setAddressOptions(options ?? []))
      .catch(() => setAddressOptions([]));
  }, [enabled, affiliationUrl]);

  const loadAffiliationOptions = React.useCallback(
    (address: string[] | undefined) => {
      if (!enabled || !address || address.length === 0) {
        setAffiliationOptions([]);
        return;
      }
      const code = address[address.length - 1];
      UserBackend.getAffiliationOptions(affiliationUrl.split("|")[1], code)
        .then((options) => setAffiliationOptions(options ?? []))
        .catch(() => setAffiliationOptions([]));
    },
    [enabled, affiliationUrl],
  );

  // the user arrives with an address already saved, so seed the affiliation list once
  const seededFor = React.useRef<string | null>(null);
  React.useEffect(() => {
    const key = JSON.stringify(user?.address ?? []);
    if (!enabled || seededFor.current === key) {
      return;
    }
    seededFor.current = key;
    loadAffiliationOptions(user?.address);
  }, [enabled, user?.address, loadAffiliationOptions]);

  return {enabled, addressOptions, affiliationOptions, loadAffiliationOptions};
}

/** The cascader replacement: one combobox per level of the address tree. */
export function AffiliationAddressSelect({
  value,
  options,
  onChange,
}: {
  value: string[] | undefined;
  options: CascaderOption[];
  onChange: (value: string[]) => void;
}) {
  const levels: CascaderOption[][] = [];
  let current = options;
  const picked = value ?? [];

  for (let i = 0; current && current.length > 0; i++) {
    levels.push(current);
    const next = current.find((item) => item.value === picked[i]);
    if (!next?.children || next.children.length === 0) {
      break;
    }
    current = next.children;
  }

  return (
    <div className="flex flex-wrap gap-2">
      {levels.map((levelOptions, level) => (
        <SearchableSelect
          key={level}
          className="w-[220px]"
          value={picked[level] ?? ""}
          allowUnknownValue={false}
          placeholder={i18next.t("signup:Please input your address!")}
          options={levelOptions.map((item) => ({value: item.value, label: item.label, keywords: item.label}))}
          onChange={(next) => onChange([...picked.slice(0, level), next])}
        />
      ))}
    </div>
  );
}

/** Affiliation: a free-text input, or a select over the address's options. */
export function AffiliationField({
  enabled,
  value,
  options,
  onChange,
}: {
  enabled: boolean;
  value: string | undefined;
  options: AffiliationOption[];
  /** `score` is the affiliation's id, only set when picked from the options */
  onChange: (affiliation: string, score?: number) => void;
}) {
  if (!enabled) {
    return <Input value={value ?? ""} onChange={(e) => onChange(e.target.value)} />;
  }

  return (
    <SearchableSelect
      value={value ?? ""}
      allowUnknownValue={false}
      onChange={(name) => onChange(name, options.find((item) => item.name === name)?.id ?? 0)}
      options={[
        {value: "", label: `(${i18next.t("general:empty")})`},
        ...options.map((item) => ({value: item.name, label: item.name})),
      ]}
    />
  );
}
