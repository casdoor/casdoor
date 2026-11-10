import type {EnumMap} from "@/lib/enum-labels";

/** The kinds of privileged account, see object/privileged_account.go */
export const PAM_ACCOUNT_TYPES: EnumMap = {
  "Linux": {i18nKey: "pam:Linux or Unix over SSH"},
  "Windows": {i18nKey: "pam:Windows (no rotation)"},
  "MySQL": {i18nKey: "MySQL"},
  "PostgreSQL": {i18nKey: "PostgreSQL"},
  "LDAP": {i18nKey: "LDAP"},
  "Active Directory": {i18nKey: "Active Directory"},
  "Other": {i18nKey: "pam:Web or other (no rotation)"},
  "Role": {i18nKey: "pam:Role (just-in-time elevation)"},
};

export const PAM_ROTATABLE_TYPES = ["Linux", "MySQL", "PostgreSQL", "LDAP", "Active Directory"];

export const PAM_DEFAULT_PORTS: Record<string, number> = {
  "Linux": 22,
  "Windows": 3389,
  "MySQL": 3306,
  "PostgreSQL": 5432,
  "LDAP": 636,
  "Active Directory": 636,
};

export const PAM_PASSWORD_CHARSETS: EnumMap = {
  "Complex": {i18nKey: "pam:Letters, digits and symbols"},
  "Alphanumeric": {i18nKey: "pam:Letters and digits"},
};

export const PAM_ROTATION_STATES: EnumMap = {
  "Succeeded": {i18nKey: "webhook:Success", variant: "success"},
  "Failed": {i18nKey: "webhook:Failed", variant: "destructive"},
};

export const PAM_REQUEST_STATES: EnumMap = {
  "Pending": {i18nKey: "webhook:Pending", variant: "info"},
  "Approved": {i18nKey: "permission:Approved", variant: "warning"},
  "Rejected": {i18nKey: "pam:Rejected", variant: "destructive"},
  "Active": {i18nKey: "subscription:Active", variant: "success"},
  "Expired": {i18nKey: "subscription:Expired", variant: "secondary"},
  "Returned": {i18nKey: "pam:Returned", variant: "secondary"},
  "Cancelled": {i18nKey: "pam:Cancelled", variant: "secondary"},
};

export function isPamRotatable(type: string) {
  return PAM_ROTATABLE_TYPES.includes(type);
}
