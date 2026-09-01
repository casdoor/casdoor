import * as React from "react";
import i18next from "i18next";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import {SearchableSelect} from "@/components/common/SearchableSelect";

interface OrganizationSelectProps {
  value?: string;
  onChange?: (value: string) => void;
  withAll?: boolean;
  className?: string;
  disabled?: boolean;
  placeholder?: string;
  /** pass a pre-loaded list to skip the request */
  organizations?: {name: string; displayName: string}[];
}

export function OrganizationSelect({
  value,
  onChange,
  withAll,
  className,
  disabled,
  placeholder,
  organizations: preloaded,
}: OrganizationSelectProps) {
  const [organizations, setOrganizations] = React.useState<{name: string; displayName: string}[]>(preloaded ?? []);

  const load = React.useCallback(() => {
    if (preloaded !== undefined) {
      return;
    }
    OrganizationBackend.getOrganizationNames("admin").then((res: any) => {
      if (res.status === "ok") {
        setOrganizations(res.data ?? []);
      }
    });
  }, [preloaded]);

  React.useEffect(() => {
    load();
    window.addEventListener("storageOrganizationsChanged", load);
    return () => window.removeEventListener("storageOrganizationsChanged", load);
  }, [load]);

  React.useEffect(() => {
    if (preloaded !== undefined) {
      setOrganizations(preloaded);
    }
  }, [preloaded]);

  const options = React.useMemo(() => {
    const items = organizations.map((organization) => ({
      value: organization.name,
      label: organization.displayName || organization.name,
    }));
    if (withAll) {
      items.unshift({value: "All", label: i18next.t("general:All")});
    }
    return items;
  }, [organizations, withAll]);

  return (
    <SearchableSelect
      value={value}
      onChange={(v) => onChange?.(v)}
      options={options}
      disabled={disabled}
      className={className}
      placeholder={placeholder ?? i18next.t("login:Please select an organization")}
    />
  );
}
