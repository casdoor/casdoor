/**
 * `organization.accountItems` drives which rows the user page shows and which of
 * them are editable. Ported from the `isAccountItemVisible` / `renderAccountItem`
 * gating in `web/src/UserEditPage.js`.
 *
 * The backend enforces none of this — it is a per-organization UI policy — but the
 * antd frontend honours it, so an organization that hides "Balance" or marks "Name"
 * `Immutable` has to behave the same here.
 */
export interface AccountItem {
  name: string;
  visible?: boolean;
  viewRule?: "Public" | "Self" | "Admin" | string;
  modifyRule?: "Self" | "Admin" | "Immutable" | string;
  tab?: string;
  regex?: string;
}

export interface AccountItemRules {
  isAdmin: boolean;
  isSelfOrAdmin: boolean;
  user: any;
}

/** The rows whose editing is frozen once the user's identity has been verified. */
const ID_VERIFICATION_ITEMS = new Set(["ID card info", "ID card", "ID card type", "Real name"]);

export function getAccountItemMap(organization: any): Record<string, AccountItem> {
  const items: AccountItem[] = organization?.accountItems ?? [];
  const map: Record<string, AccountItem> = {};
  items.forEach((item) => {
    if (item?.name) {
      map[item.name] = item;
    }
  });
  return map;
}

export function isAccountItemVisible(item: AccountItem | undefined, rules: AccountItemRules): boolean {
  if (!item || !item.visible) {
    return false;
  }
  if (item.viewRule === "Self") {
    return rules.isSelfOrAdmin;
  }
  if (item.viewRule === "Admin") {
    return rules.isAdmin;
  }
  return true;
}

export function isAccountItemDisabled(item: AccountItem | undefined, rules: AccountItemRules): boolean {
  if (!item) {
    return false;
  }

  let disabled = false;
  if (item.modifyRule === "Self") {
    disabled = !rules.isSelfOrAdmin;
  } else if (item.modifyRule === "Admin") {
    disabled = !rules.isAdmin;
  } else if (item.modifyRule === "Immutable") {
    disabled = true;
  }

  // the built-in admin cannot be moved between organizations or renamed
  if (item.name === "Organization" || item.name === "Name") {
    if (rules.user?.owner === "built-in" && rules.user?.name === "admin") {
      disabled = true;
    }
  }

  if (ID_VERIFICATION_ITEMS.has(item.name) && rules.user?.isVerified) {
    disabled = true;
  }

  return disabled;
}

/** `true` when the row's `regex` rejects the current value, like antd's Form rule. */
export function checkAccountItemRegex(item: AccountItem | undefined, value: any): boolean {
  if (!item?.regex || value === undefined || value === null || value === "") {
    return false;
  }
  try {
    return !new RegExp(item.regex).test(String(value));
  } catch (error) {
    // an organization can save an invalid regex; antd would throw here, we just pass
    return false;
  }
}
