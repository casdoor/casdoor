
import i18next from "i18next";
import {Input} from "@/components/ui/input";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SelectField} from "@/components/common/SelectField";
import {TagsInput} from "@/components/common/TagsInput";
import {EditableTable} from "@/components/crud/EditableTable";
import {FormRow} from "@/components/crud/FormRow";
import type {ApplicationTabProps} from "@/components/application/types";
import {enumSelectOptions, type EnumMap} from "@/lib/enum-labels";
import * as Setting from "@/lib/setting";

const GRANT_TYPES = [
  "authorization_code",
  "password",
  "client_credentials",
  "token",
  "id_token",
  "refresh_token",
  "device_code",
];
const TOKEN_FORMATS = ["JWT", "JWT-Empty", "JWT-Custom", "JWT-Standard"];
const TOKEN_SIGNING_METHODS = ["RS256", "RS512", "ES256", "ES512", "ES384"];
const TOKEN_ATTRIBUTE_TYPES = ["String", "Number", "Boolean"];
const TOKEN_ATTRIBUTE_CATEGORIES: EnumMap = {
  "Static Value": {i18nKey: "application:Static Value"},
  "Existing Field": {i18nKey: "application:Existing Field"},
};

/** the user fields an "Existing Field" token attribute may copy */
const TOKEN_ATTRIBUTE_USER_FIELDS = [
  "Owner", "Name", "Id", "DisplayName", "Avatar", "Email", "Phone",
  "Tag", "Roles", "Permissions", "permissionNames", "Groups",
];

/** The "OIDC/OAuth" tab: the client credentials, the token format, and what goes into a token. */
export function ApplicationOidcOauthTab({application, updateField}: ApplicationTabProps) {
  return (
    <>
      {/* both are editable so that an admin can rotate the pair, as in the antd page */}
      <FormRow labelKey="provider:Client ID">
        <Input value={application.clientId ?? ""} onChange={(e) => updateField("clientId", e.target.value)} />
      </FormRow>
      <FormRow labelKey="provider:Client secret">
        <Input
          value={application.clientSecret ?? ""}
          onChange={(e) => updateField("clientSecret", e.target.value)}
        />
      </FormRow>
      <FormRow block labelKey="application:Redirect URLs">
        <TagsInput value={application.redirectUris ?? []} onChange={(v) => updateField("redirectUris", v)} />
      </FormRow>
      <FormRow block labelKey="application:Forced redirect origin">
        <Input
          value={application.forcedRedirectOrigin ?? ""}
          onChange={(e) => updateField("forcedRedirectOrigin", e.target.value)}
        />
      </FormRow>
      <FormRow block labelKey="application:Backchannel logout URL">
        <Input
          value={application.backchannelLogoutUri ?? ""}
          onChange={(e) => updateField("backchannelLogoutUri", e.target.value)}
        />
      </FormRow>
      <FormRow block labelKey="application:Grant types">
        <MultiSelect
          value={application.grantTypes ?? []}
          onChange={(v) => updateField("grantTypes", v)}
          options={GRANT_TYPES.map((item) => ({value: item, label: item}))}
        />
      </FormRow>
      {/* scopes are ScopeItem objects (name / displayName / description), the
          same three columns antd's ScopeTable edits */}
      <FormRow labelKey="general:Scopes" block>
        <EditableTable
          rows={application.scopes ?? []}
          onChange={(rows) => updateField("scopes", rows)}
          newRow={() => ({name: "", displayName: "", description: ""})}
          columns={[
            {
              key: "name",
              title: i18next.t("general:Name"),
              width: "25%",
              render: (row: any, index, update) => (
                <Input
                  value={row.name ?? ""}
                  placeholder="e.g., files:read"
                  onChange={(e) => update({name: e.target.value})}
                />
              ),
            },
            {
              key: "displayName",
              title: i18next.t("general:Display name"),
              width: "25%",
              render: (row: any, index, update) => (
                <Input
                  value={row.displayName ?? ""}
                  placeholder="e.g., Read Files"
                  onChange={(e) => update({displayName: e.target.value})}
                />
              ),
            },
            {
              key: "description",
              title: i18next.t("general:Description"),
              width: "40%",
              render: (row: any, index, update) => (
                <Input
                  value={row.description ?? ""}
                  placeholder="e.g., Allow reading your files and documents"
                  onChange={(e) => update({description: e.target.value})}
                />
              ),
            },
          ]}
        />
      </FormRow>
      <FormRow labelKey="general:Custom scopes" block>
        <EditableTable
          rows={application.customScopes ?? []}
          onChange={(rows) => updateField("customScopes", rows)}
          newRow={() => ({scope: "", displayName: "", description: ""})}
          columns={[
            {
              key: "scope",
              title: i18next.t("general:Name"),
              width: 200,
              render: (row: any, _i, patch) => (
                <Input value={row.scope ?? ""} onChange={(e) => patch({scope: e.target.value})} />
              ),
            },
            {
              key: "displayName",
              title: i18next.t("general:Display name"),
              width: 200,
              render: (row: any, _i, patch) => (
                <Input value={row.displayName ?? ""} onChange={(e) => patch({displayName: e.target.value})} />
              ),
            },
            {
              key: "description",
              title: i18next.t("general:Description"),
              render: (row: any, _i, patch) => (
                <Input value={row.description ?? ""} onChange={(e) => patch({description: e.target.value})} />
              ),
            },
          ]}
        />
      </FormRow>
      <FormRow block labelKey="application:Token format">
        <SelectField
          value={application.tokenFormat ?? "JWT"}
          onChange={(v) => updateField("tokenFormat", v)}
          options={TOKEN_FORMATS.map((item) => ({id: item, name: item}))}
        />
      </FormRow>
      <FormRow block labelKey="application:Token signing method">
        <SelectField
          value={application.tokenSigningMethod ?? "RS256"}
          onChange={(v) => updateField("tokenSigningMethod", v)}
          options={TOKEN_SIGNING_METHODS.map((item) => ({id: item, name: item}))}
        />
      </FormRow>
      <FormRow block labelKey="application:Token fields">
        <MultiSelect
          value={application.tokenFields ?? []}
          onChange={(v) => updateField("tokenFields", v)}
          creatable
          options={Setting.UserFields.map((item: string) => ({value: item, label: item}))}
        />
      </FormRow>
      {application.tokenFormat === "JWT-Custom" ? (
        <FormRow labelKey="general:Token attributes" block>
          <EditableTable
            rows={application.tokenAttributes ?? []}
            onChange={(rows) => updateField("tokenAttributes", rows)}
            newRow={() => ({name: "", category: "Static Value", value: "", type: "Array"})}
            columns={[
              {
                key: "name",
                title: i18next.t("general:Name"),
                width: 180,
                render: (row: any, _i, patch) => (
                  <Input value={row.name ?? ""} onChange={(e) => patch({name: e.target.value})} />
                ),
              },
              {
                key: "category",
                title: i18next.t("general:Category"),
                width: 160,
                render: (row: any, _i, patch) => (
                  <SelectField
                    value={row.category ?? "Static Value"}
                    onChange={(v) => patch({category: v, value: ""})}
                    options={enumSelectOptions(TOKEN_ATTRIBUTE_CATEGORIES)}
                  />
                ),
              },
              {
                key: "value",
                title: i18next.t("webhook:Value"),
                // an "Existing Field" attribute copies one of the user's own
                // fields, so the value is picked rather than typed
                render: (row: any, _i, patch) =>
                  row.category === "Existing Field" ? (
                    <SelectField
                      value={row.value ?? ""}
                      onChange={(v) => patch({value: v})}
                      options={TOKEN_ATTRIBUTE_USER_FIELDS.map((field) => ({id: field, name: field}))}
                    />
                  ) : (
                    <Input value={row.value ?? ""} onChange={(e) => patch({value: e.target.value})} />
                  ),
              },
              {
                key: "type",
                title: i18next.t("general:Type"),
                width: 140,
                render: (row: any, _i, patch) => (
                  <SelectField
                    value={row.type}
                    onChange={(v) => patch({type: v})}
                    options={TOKEN_ATTRIBUTE_TYPES.map((item) => ({id: item, name: item}))}
                  />
                ),
              },
            ]}
          />
        </FormRow>
      ) : null}
      <FormRow labelKey="application:Token expire">
        <Input
          type="number"
          value={application.expireInHours ?? 168}
          onChange={(e) => updateField("expireInHours", Setting.myParseInt(e.target.value))}
        />
      </FormRow>
      <FormRow labelKey="application:Refresh token expire">
        <Input
          type="number"
          value={application.refreshExpireInHours ?? 168}
          onChange={(e) => updateField("refreshExpireInHours", Setting.myParseInt(e.target.value))}
        />
      </FormRow>
    </>
  );
}
