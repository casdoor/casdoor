/** What a delegated admin (分级管理) administers, as get-account tells the console. */
export interface DelegatedAdminInfo {
  active: boolean;
  groups: string[];
  rootGroups?: string[];
  applications: string[];
  actions: string[];
  message?: string;
}

export function getDelegatedAdmin(account: any): DelegatedAdminInfo | null {
  return account?.delegatedAdmin ?? null;
}

/** The account administers some groups or applications as a delegated admin and the license covers it. */
export function isDelegatedAdmin(account: any): boolean {
  return getDelegatedAdmin(account)?.active === true;
}

/** The delegated admin may do the action: "users", "groups", "applications" or "passwords". */
export function canDelegated(account: any, action: string): boolean {
  const info = getDelegatedAdmin(account);
  return info?.active === true && info.actions.includes(action);
}

/** The delegated admin administers the user: someone else in the same organization the backend lists for them. */
export function isDelegatedAdminOf(account: any, user: any): boolean {
  if (!isDelegatedAdmin(account) || !user || user.owner !== account.owner || user.name === account.name || user.isAdmin) {
    return false;
  }
  const groups: string[] = getDelegatedAdmin(account)?.groups ?? [];
  return (user.groups ?? []).some((group: string) => groups.includes(group.includes("/") ? group.split("/")[1] : group));
}

export function isDelegatedRootGroup(account: any, groupName: string): boolean {
  return (getDelegatedAdmin(account)?.rootGroups ?? []).includes(groupName);
}
