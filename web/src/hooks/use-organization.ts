import * as React from "react";
import * as Setting from "@/lib/setting";
import {useAccount} from "@/hooks/use-account";

function resolveFilter(account) {
  return Setting.isDefaultOrganizationSelected(account) ? "" : Setting.getRequestOrganization(account);
}

/**
 * Admins switch the console's organization from the header; the antd frontend
 * broadcast the change through a "storageOrganizationChanged" window event and
 * these hooks keep that contract.
 */
function useOrganization(resolve: (account: any) => string, overrideName?: string): string {
  const {account} = useAccount();
  const compute = React.useCallback(
    () => overrideName ?? (account ? resolve(account) : ""),
    [account, overrideName, resolve],
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

/** The organization the console is scoped to. "All" means the account's own organization. */
export function useRequestOrganization(overrideName?: string): string {
  return useOrganization(Setting.getRequestOrganization, overrideName);
}

/**
 * The organization a list request should be filtered by. "All" means no filter at
 * all, so this yields an empty string instead of the admin's own organization,
 * which would hide every other organization's rows.
 */
export function useOrganizationFilter(overrideName?: string): string {
  return useOrganization(resolveFilter, overrideName);
}
