import * as React from "react";
import * as Setting from "@/lib/setting";
import {useAccount} from "@/hooks/use-account";

/**
 * The organization the console is currently scoped to. Admins can switch it from
 * the header; the antd frontend broadcast the change through a
 * "storageOrganizationChanged" window event and this hook keeps that contract.
 */
export function useRequestOrganization(overrideName?: string): string {
  const {account} = useAccount();
  const compute = React.useCallback(
    () => overrideName ?? (account ? Setting.getRequestOrganization(account) : ""),
    [account, overrideName],
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
