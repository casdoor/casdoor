import * as React from "react";
import * as Setting from "@/lib/setting";
import {useAccount} from "@/hooks/use-account";

/**
 * The organization the console is currently scoped to. Admins can switch it from
 * the header; the antd frontend broadcast the change through a
 * "storageOrganizationChanged" window event and this hook keeps that contract.
 * With includeAll, a global admin's "All" selection yields an empty filter.
 */
export function useRequestOrganization(overrideName?: string, includeAll = false): string {
  const {account} = useAccount();
  const compute = React.useCallback(
    () => {
      if (overrideName !== undefined && overrideName !== null) {
        return overrideName;
      }
      if (!account || (includeAll && Setting.isDefaultOrganizationSelected(account))) {
        return "";
      }
      return Setting.getRequestOrganization(account);
    },
    [account, overrideName, includeAll],
  );
  const [organizationName, setOrganizationName] = React.useState(compute);

  React.useEffect(() => {
    setOrganizationName(compute());
    const handler = () => setOrganizationName(compute());
    window.addEventListener("storageOrganizationChanged", handler);
    return () => window.removeEventListener("storageOrganizationChanged", handler);
  }, [compute]);

  return organizationName;
}
