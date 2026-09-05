import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import type {ColumnDef, ColumnFilterOption} from "@/components/crud/types";

/**
 * Translated labels and colours for the stored enum values Casdoor shows in
 * tables and selects. The antd frontend spelled these out per page as a
 * `switch` returning `Setting.getTag(color, i18next.t(key))`; keeping one map
 * per enum here means a list column and its edit-page select can never drift.
 *
 * The variants follow the antd colours: processing → info, success → success,
 * warning/gold/orange → warning, error/red → destructive, default → secondary.
 */
export type BadgeVariant = "default" | "secondary" | "destructive" | "success" | "warning" | "info" | "outline";

export interface EnumEntry {
  /** i18n key; resolved at render so switching language updates the label */
  i18nKey: string;
  variant?: BadgeVariant;
}

export type EnumMap = Record<string, EnumEntry>;

export const TICKET_STATES: EnumMap = {
  "Open": {i18nKey: "ticket:Open", variant: "info"},
  "In Progress": {i18nKey: "ticket:In Progress", variant: "warning"},
  "Resolved": {i18nKey: "ticket:Resolved", variant: "success"},
  "Closed": {i18nKey: "ticket:Closed", variant: "secondary"},
};

export const SUBSCRIPTION_STATES: EnumMap = {
  "Pending": {i18nKey: "webhook:Pending", variant: "info"},
  "Active": {i18nKey: "subscription:Active", variant: "success"},
  "Upcoming": {i18nKey: "subscription:Upcoming", variant: "warning"},
  "Expired": {i18nKey: "subscription:Expired", variant: "warning"},
  "Error": {i18nKey: "general:Error", variant: "destructive"},
  "Suspended": {i18nKey: "subscription:Suspended", variant: "secondary"},
};

export const INVITATION_STATES: EnumMap = {
  "Active": {i18nKey: "subscription:Active", variant: "success"},
  "Suspended": {i18nKey: "subscription:Suspended", variant: "secondary"},
};

export const COUPON_STATES: EnumMap = {
  "Active": {i18nKey: "subscription:Active", variant: "success"},
  "Inactive": {i18nKey: "key:Inactive", variant: "secondary"},
  "Expired": {i18nKey: "subscription:Expired", variant: "destructive"},
};

export const KEY_STATES: EnumMap = {
  "Active": {i18nKey: "subscription:Active", variant: "success"},
  "Inactive": {i18nKey: "key:Inactive", variant: "secondary"},
};

export const PERMISSION_EFFECTS: EnumMap = {
  "Allow": {i18nKey: "permission:Allow", variant: "success"},
  "Deny": {i18nKey: "permission:Deny", variant: "destructive"},
};

export const PERMISSION_STATES: EnumMap = {
  "Approved": {i18nKey: "permission:Approved", variant: "success"},
  "Pending": {i18nKey: "webhook:Pending", variant: "destructive"},
};

export const PERMISSION_ACTIONS: EnumMap = {
  "Read": {i18nKey: "permission:Read"},
  "Write": {i18nKey: "permission:Write"},
  "Admin": {i18nKey: "general:Admin"},
};

/** the actions an "API" permission grants are HTTP verbs, not read/write */
export const PERMISSION_API_ACTIONS: EnumMap = {
  "POST": {i18nKey: "POST"},
  "GET": {i18nKey: "GET"},
};

export const PERMISSION_RESOURCE_TYPES: EnumMap = {
  "Application": {i18nKey: "general:Application"},
  "TreeNode": {i18nKey: "permission:TreeNode"},
  "Custom": {i18nKey: "general:Custom"},
  "API": {i18nKey: "API"},
};

export const COUPON_DISCOUNT_TYPES: EnumMap = {
  "percentage": {i18nKey: "coupon:Percentage"},
  "fixed": {i18nKey: "coupon:Fixed"},
};

export const COUPON_SCOPES: EnumMap = {
  "universal": {i18nKey: "coupon:Universal", variant: "info"},
  "product": {i18nKey: "coupon:Product specific", variant: "success"},
  "user": {i18nKey: "coupon:User specific", variant: "warning"},
};

export const KEY_TYPES: EnumMap = {
  "Organization": {i18nKey: "general:Organization"},
  "Application": {i18nKey: "general:Application"},
  "User": {i18nKey: "general:User"},
  "General": {i18nKey: "general:General"},
  "Prometheus": {i18nKey: "Prometheus"},
};

/** what a provider row's "Rule" means depends on the provider it points at */
export const PROVIDER_CAPTCHA_RULES: EnumMap = {
  "None": {i18nKey: "general:None"},
  "Dynamic": {i18nKey: "application:Dynamic"},
  "Always": {i18nKey: "application:Always"},
  "Internet-Only": {i18nKey: "application:Internet-Only"},
};

export const PROVIDER_GOOGLE_RULES: EnumMap = {
  "Default": {i18nKey: "general:Default"},
  "OneTap": {i18nKey: "One Tap"},
};

export const PROVIDER_CODE_RULES: EnumMap = {
  "all": {i18nKey: "All"},
  "signup": {i18nKey: "Signup"},
  "login": {i18nKey: "Login"},
  "forget": {i18nKey: "Forget Password"},
  "reset": {i18nKey: "Reset Password"},
  "mfaSetup": {i18nKey: "Set MFA"},
  "mfaAuth": {i18nKey: "MFA Auth"},
};

/** which of the user's identifiers an external account may bind to */
export const PROVIDER_BINDING_RULES: EnumMap = {
  "Email": {i18nKey: "general:Email"},
  "Name": {i18nKey: "general:Name"},
  "Phone": {i18nKey: "general:Phone"},
};

export const LDAP_PASSWORD_TYPES: EnumMap = {
  "Plain": {i18nKey: "general:Plain"},
  "SSHA": {i18nKey: "SSHA"},
  "MD5": {i18nKey: "MD5"},
};

/**
 * The label for a stored value. Values with no entry (and keys with no
 * namespace, such as "GET") fall through to the text itself, which is what the
 * antd `default:` branches did.
 */
export function enumLabel(map: EnumMap, value: any): string {
  const entry = map[value];
  if (!entry) {
    return `${value ?? ""}`;
  }
  return entry.i18nKey.includes(":") ? i18next.t(entry.i18nKey) : entry.i18nKey;
}

/** Options for `SearchableSelect` / `MultiSelect`. */
export function enumOptions(map: EnumMap) {
  return Object.keys(map).map((value) => ({value, label: enumLabel(map, value)}));
}

/** Options for `SelectField`, which takes `{id, name}`. */
export function enumSelectOptions(map: EnumMap) {
  return Object.keys(map).map((value) => ({id: value, name: enumLabel(map, value)}));
}

export function EnumBadge({map, value}: {map: EnumMap; value: any}) {
  if (value === undefined || value === null || value === "") {
    return null;
  }
  return <Badge variant={(map[value]?.variant ?? "outline") as any}>{enumLabel(map, value)}</Badge>;
}

/** A table column that renders a stored enum value as a translated badge. */
export function enumColumn<T>(options: {
  dataIndex: string;
  title: React.ReactNode;
  map: EnumMap;
  width?: number | string;
  sortable?: boolean;
  searchable?: boolean;
  /** render the header filter menu; pass `true` to build it from `map` */
  filters?: boolean | ColumnFilterOption[];
}): ColumnDef<T> {
  const {dataIndex, title, map, width = 110, sortable = true, searchable = false, filters} = options;
  return {
    dataIndex,
    title,
    width,
    sortable,
    searchable,
    filters: filters === true ? enumFilters(map) : filters || undefined,
    render: (value) => <EnumBadge map={map} value={value} />,
  };
}

/** the values of an enum map as a column filter menu, in declaration order */
export function enumFilters(map: EnumMap): ColumnFilterOption[] {
  return Object.keys(map).map((value) => ({value, label: enumLabel(map, value)}));
}
