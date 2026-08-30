// The default `tableColumns` Casdoor fills in when the syncer type changes, a
// port of `getSyncerTableColumns()` in web/src/SyncerEditPage.js. "Database"
// has no default mapping — the columns come from the user's own table.

export interface SyncerTableColumn {
  name: string;
  type: string;
  casdoorName: string;
  isHashed: boolean;
  values: string[];
}

const column = (name: string, casdoorName: string, type = "string"): SyncerTableColumn => ({
  name,
  type,
  casdoorName,
  isHashed: true,
  values: [],
});

const DEFAULT_COLUMNS: Record<string, SyncerTableColumn[]> = {
  "Keycloak": [
    column("ID", "Id"),
    column("USERNAME", "Name"),
    column("LAST_NAME+FIRST_NAME", "DisplayName"),
    column("EMAIL", "Email"),
    column("EMAIL_VERIFIED", "EmailVerified", "boolean"),
    column("FIRST_NAME", "FirstName"),
    column("LAST_NAME", "LastName"),
    column("CREATED_TIMESTAMP", "CreatedTime"),
    column("ENABLED", "IsForbidden", "boolean"),
  ],
  "WeCom": [
    column("userid", "Id"),
    column("name", "DisplayName"),
    column("email", "Email"),
    column("mobile", "Phone"),
    column("avatar", "Avatar"),
    column("position", "Title"),
    column("gender", "Gender"),
  ],
  "Azure AD": [
    column("id", "Id"),
    column("userPrincipalName", "Name"),
    column("displayName", "DisplayName"),
    column("givenName", "FirstName"),
    column("surname", "LastName"),
    column("mail", "Email"),
    column("mobilePhone", "Phone"),
    column("jobTitle", "Title"),
    column("officeLocation", "Location"),
    column("preferredLanguage", "Language"),
    column("accountEnabled", "IsForbidden", "boolean"),
  ],
  "Google Workspace": [
    column("id", "Id"),
    column("primaryEmail", "Name"),
    column("name.fullName", "DisplayName"),
    column("name.givenName", "FirstName"),
    column("name.familyName", "LastName"),
    column("suspended", "IsForbidden", "boolean"),
    column("isAdmin", "IsAdmin", "boolean"),
  ],
  "DingTalk": [
    column("userid", "Id"),
    column("unionid", "Name"),
    column("name", "DisplayName"),
    column("email", "Email"),
    column("mobile", "Phone"),
    column("avatar", "Avatar"),
    column("title", "Title"),
    column("active", "IsForbidden", "boolean"),
  ],
  "Active Directory": [
    column("objectGUID", "Id"),
    column("sAMAccountName", "Name"),
    column("displayName", "DisplayName"),
    column("givenName", "FirstName"),
    column("sn", "LastName"),
    column("mail", "Email"),
    column("mobile", "Phone"),
    column("title", "Title"),
    column("department", "Affiliation"),
    column("userAccountControl", "IsForbidden"),
  ],
  "Lark": [
    column("user_id", "Id"),
    column("name", "DisplayName"),
    column("email", "Email"),
    column("mobile", "Phone"),
    column("avatar", "Avatar"),
    column("job_title", "Title"),
    column("gender", "Gender", "number"),
  ],
  "Okta": [
    column("id", "Id"),
    column("profile.login", "Name"),
    column("profile.displayName", "DisplayName"),
    column("profile.firstName", "FirstName"),
    column("profile.lastName", "LastName"),
    column("profile.email", "Email"),
    column("profile.mobilePhone", "Phone"),
    column("profile.title", "Title"),
    column("profile.preferredLanguage", "Language"),
    column("status", "IsForbidden"),
  ],
  "SCIM": [
    column("id", "Id"),
    column("userName", "Name"),
    column("displayName", "DisplayName"),
    column("name.givenName", "FirstName"),
    column("name.familyName", "LastName"),
    column("emails", "Email"),
    column("phoneNumbers", "Phone"),
    column("title", "Title"),
    column("preferredLanguage", "Language"),
    column("active", "IsForbidden", "boolean"),
  ],
  "AWS IAM": [
    column("UserId", "Id"),
    column("UserName", "Name"),
    column("UserName", "DisplayName"),
    column("Tags.Email", "Email"),
    column("Tags.Phone", "Phone"),
    column("Tags.FirstName", "FirstName"),
    column("Tags.LastName", "LastName"),
    column("Tags.Title", "Title"),
    column("Tags.Department", "Affiliation"),
    column("CreateDate", "CreatedTime"),
  ],
};

export function getSyncerTableColumns(type: string): SyncerTableColumn[] {
  const columns = DEFAULT_COLUMNS[type];
  return columns ? columns.map((item) => ({...item, values: []})) : [];
}
