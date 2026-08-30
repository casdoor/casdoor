import * as React from "react";
import {FormRow} from "@/components/crud/FormRow";
import {
  type AccountItem,
  type AccountItemRules,
  checkAccountItemRegex,
  getAccountItemMap,
  isAccountItemDisabled,
  isAccountItemVisible,
} from "@/lib/account-items";
import i18next from "i18next";

interface AccountItemsContextValue extends AccountItemRules {
  items: Record<string, AccountItem>;
  /**
   * The organization defines no accountItems (not loaded yet, or never configured).
   * antd would render an empty form; showing every row is the more useful fallback
   * and matches what this page did before the policy was honoured at all.
   */
  unconfigured: boolean;
}

const AccountItemsContext = React.createContext<AccountItemsContextValue>({
  items: {},
  unconfigured: true,
  isAdmin: false,
  isSelfOrAdmin: false,
  user: null,
});

export function AccountItemsProvider({
  organization,
  isAdmin,
  isSelfOrAdmin,
  user,
  children,
}: {
  organization: any;
  isAdmin: boolean;
  isSelfOrAdmin: boolean;
  user: any;
  children: React.ReactNode;
}) {
  const value = React.useMemo<AccountItemsContextValue>(() => {
    const items = getAccountItemMap(organization);
    return {
      items,
      unconfigured: Object.keys(items).length === 0,
      isAdmin,
      isSelfOrAdmin,
      user,
    };
  }, [organization, isAdmin, isSelfOrAdmin, user]);

  return <AccountItemsContext.Provider value={value}>{children}</AccountItemsContext.Provider>;
}

export function useAccountItem(name?: string) {
  const ctx = React.useContext(AccountItemsContext);
  if (!name || ctx.unconfigured) {
    return {visible: true, disabled: false, item: undefined as AccountItem | undefined};
  }
  const item = ctx.items[name];
  return {
    visible: isAccountItemVisible(item, ctx),
    disabled: isAccountItemDisabled(item, ctx),
    item,
  };
}

interface AccountItemRowProps {
  /** the `organization.accountItems` entry that governs this row, e.g. "Display name" */
  name?: string;
  labelKey?: string;
  label?: React.ReactNode;
  block?: boolean;
  className?: string;
  htmlFor?: string;
  /** the value to test against the item's `regex`, when it defines one */
  value?: any;
  children: React.ReactNode;
}

/**
 * A `FormRow` that obeys the organization's accountItems policy: hidden when the
 * item is invisible to the viewer, and wrapped in a disabled `<fieldset>` when the
 * item is not modifiable by them. A row with no `name` is always shown, which is
 * how the handful of rows this frontend adds on top of antd's list behave.
 */
export function AccountItemRow({name, labelKey, label, block, className, htmlFor, value, children}: AccountItemRowProps) {
  const {visible, disabled, item} = useAccountItem(name);

  if (!visible) {
    return null;
  }

  const regexError = !disabled && checkAccountItemRegex(item, value);

  return (
    <FormRow labelKey={labelKey} label={label} block={block} className={className} htmlFor={htmlFor}>
      {disabled ? (
        // `fieldset[disabled]` natively disables every control inside, so the row's
        // existing markup does not have to thread a `disabled` prop down by hand
        <fieldset disabled className="min-w-0 opacity-60">
          {children}
        </fieldset>
      ) : (
        children
      )}
      {regexError ? (
        <p className="mt-1 text-xs text-destructive">{i18next.t("user:This field value doesn't match the pattern rule")}</p>
      ) : null}
    </FormRow>
  );
}
