import * as React from "react";
import i18next from "i18next";
import {Link, useNavigate, useParams} from "react-router-dom";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {Loading} from "@/components/common/Loading";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {TagsInput} from "@/components/common/TagsInput";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {EditableTable} from "@/components/crud/EditableTable";
import {FormRow} from "@/components/crud/FormRow";
import {useAccount} from "@/hooks/use-account";
import {useEditRecord} from "@/hooks/use-edit-record";
import {
  useApplicationOptions,
  useGroupOptions,
  useOrganizationOptions,
} from "@/hooks/use-options";
import {submitEdit} from "@/lib/crud";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as MfaBackend from "@/backend/MfaBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

const GENDERS = ["Male", "Female", "Other"];
const ID_CARD_TYPES = ["ID card", "Passport", "Driver license"];

export default function UserEditPage({self}: {self?: boolean} = {}) {
  const params = useParams();
  const navigate = useNavigate();
  const {account, reload: reloadAccount} = useAccount();
  // "/account" edits the signed-in user, every other route takes the name from the URL
  const organizationName = self ? account?.owner ?? "" : params.organizationName ?? "";
  const userName = self ? account?.name ?? "" : params.userName ?? "";
  const [saving, setSaving] = React.useState(false);
  const [organization, setOrganization] = React.useState<any>({});
  const [newPassword, setNewPassword] = React.useState("");
  const [mfaItems, setMfaItems] = React.useState<any[]>([]);
  const [removingMfa, setRemovingMfa] = React.useState(false);

  const organizations = useOrganizationOptions();
  const applications = useApplicationOptions(organizationName);
  const groups = useGroupOptions(organizationName);

  const {record: user, updateField, loading, mode, setMode} = useEditRecord<any>({
    fetch: () => UserBackend.getUser(organizationName, userName),
    deps: [organizationName, userName],
  });

  React.useEffect(() => {
    setMfaItems(user?.multiFactorAuths ?? []);
  }, [user]);

  React.useEffect(() => {
    OrganizationBackend.getOrganization("admin", organizationName).then((res: any) => {
      if (res.status === "ok") {
        setOrganization(res.data ?? {});
      }
    });
  }, [organizationName]);

  if (loading || user === null || (self && !account)) {
    return <Loading />;
  }

  const isAdmin = Setting.isLocalAdminUser(account);
  const isSelf = account?.owner === user.owner && account?.name === user.name;

  const save = async(exitAfterSave: boolean) => {
    setSaving(true);
    await submitEdit({
      mode,
      record: Setting.deepCopy(user),
      add: (record) => UserBackend.addUser(record),
      update: (record) => UserBackend.updateUser(organizationName, userName, record),
      onSaved: () => {
        setMode("edit");
        if (isSelf) {
          reloadAccount();
        }
        if (exitAfterSave) {
          navigate("/users");
        } else if (user.name !== userName) {
          navigate(`/users/${user.owner}/${user.name}`, {replace: true});
        }
      },
    });
    setSaving(false);
  };

  const setPreferredMfa = (mfaType: string) => {
    MfaBackend.SetPreferredMfa({owner: user.owner, name: user.name, mfaType}).then((res: any) => {
      if (res.status === "ok") {
        setMfaItems(res.data ?? []);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  const deleteMfa = () => {
    setRemovingMfa(true);
    return MfaBackend.DeleteMfa({owner: user.owner, name: user.name})
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          setMfaItems(res.data ?? []);
        } else {
          Setting.showMessage("error", i18next.t("general:Failed to delete"));
        }
      })
      .finally(() => setRemovingMfa(false));
  };

  const setPassword = () => {
    if (!newPassword) {
      return;
    }
    UserBackend.setPassword(user.owner, user.name, "", newPassword).then((res: any) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully saved"));
        setNewPassword("");
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  const avatarUrl = Setting.getEffectiveAvatarUrl(user);
  // `properties` is a plain map on the wire, the table edits it as rows
  const propertyRows = Object.entries(user.properties ?? {}).map(([key, value]) => ({key, value}));

  return (
    <EditPageShell
      title={`${i18next.t("user:Edit User")} - ${user.displayName || user.name}`}
      mode={mode}
      backTo={self ? "/" : "/users"}
      onSave={save}
      saving={saving}
    >
      <Tabs defaultValue="account">
        <TabsList className="mb-2 flex-wrap">
          <TabsTrigger value="account">{i18next.t("user:Account")}</TabsTrigger>
          <TabsTrigger value="profile">{i18next.t("user:User Profile")}</TabsTrigger>
          <TabsTrigger value="security">{i18next.t("user:Security")}</TabsTrigger>
          <TabsTrigger value="authorization">{i18next.t("general:Authorization")}</TabsTrigger>
        </TabsList>

        <TabsContent value="account">
          <FormRow labelKey="general:Organization">
            <SearchableSelect
              value={user.owner}
              disabled={!Setting.isAdminUser(account)}
              onChange={(v) => updateField("owner", v)}
              options={organizations}
            />
          </FormRow>
          <FormRow labelKey="general:Name">
            <Input value={user.name ?? ""} disabled={!isAdmin} onChange={(e) => updateField("name", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Display name">
            <Input value={user.displayName ?? ""} onChange={(e) => updateField("displayName", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:First name">
            <Input value={user.firstName ?? ""} onChange={(e) => updateField("firstName", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Last name">
            <Input value={user.lastName ?? ""} onChange={(e) => updateField("lastName", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Avatar">
            <div className="flex items-center gap-3">
              <Avatar className="h-16 w-16">
                {avatarUrl ? <AvatarImage src={avatarUrl} alt={user.name} /> : null}
                <AvatarFallback style={{backgroundColor: Setting.getAvatarColor(user.name ?? "?"), color: "#fff"}}>
                  {(user.name || "?").charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <Input value={user.avatar ?? ""} onChange={(e) => updateField("avatar", e.target.value)} />
            </div>
          </FormRow>
          <FormRow labelKey="general:Email">
            <Input type="email" value={user.email ?? ""} onChange={(e) => updateField("email", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Phone">
            <div className="flex gap-2">
              <div className="w-32 shrink-0">
                <SearchableSelect
                  value={user.countryCode ?? ""}
                  onChange={(v) => updateField("countryCode", v)}
                  options={Setting.getCountryCodeData(organization.countryCodes).map((country: any) => ({
                    value: country.code,
                    label: `+${country.phone}`,
                    keywords: `${country.name} ${country.code} ${country.phone}`,
                  }))}
                />
              </div>
              <Input value={user.phone ?? ""} onChange={(e) => updateField("phone", e.target.value)} />
            </div>
          </FormRow>
          <FormRow labelKey="general:User type">
            <SearchableSelect
              value={user.type ?? "normal-user"}
              onChange={(v) => updateField("type", v)}
              options={(organization.userTypes?.length > 0
                ? organization.userTypes
                : ["normal-user", "paid-user"]
              ).map((item: string) => ({value: item, label: item}))}
            />
          </FormRow>
          <FormRow labelKey="general:Tag">
            <Input value={user.tag ?? ""} onChange={(e) => updateField("tag", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Application">
            <SearchableSelect
              value={user.signupApplication ?? ""}
              onChange={(v) => updateField("signupApplication", v)}
              options={applications}
            />
          </FormRow>
          <FormRow labelKey="general:Groups">
            <MultiSelect
              value={user.groups ?? []}
              onChange={(v) => updateField("groups", v)}
              options={groups}
            />
          </FormRow>
        </TabsContent>

        <TabsContent value="profile">
          <FormRow labelKey="user:Country/Region">
            <Input value={user.region ?? ""} onChange={(e) => updateField("region", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:Location">
            <Input value={user.location ?? ""} onChange={(e) => updateField("location", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:Address">
            <TagsInput value={user.address ?? []} onChange={(v) => updateField("address", v)} />
          </FormRow>
          <FormRow labelKey="user:Addresses" block>
            <EditableTable
              rows={user.addresses ?? []}
              onChange={(rows) => updateField("addresses", rows)}
              newRow={() => ({name: "", address: "", isDefault: false})}
              columns={[
                {
                  key: "name",
                  title: i18next.t("general:Name"),
                  width: 200,
                  render: (row: any, _i, patch) => (
                    <Input value={row.name ?? ""} onChange={(e) => patch({name: e.target.value})} />
                  ),
                },
                {
                  key: "address",
                  title: i18next.t("user:Address"),
                  render: (row: any, _i, patch) => (
                    <Input value={row.address ?? ""} onChange={(e) => patch({address: e.target.value})} />
                  ),
                },
                {
                  key: "isDefault",
                  title: i18next.t("general:Default"),
                  width: 110,
                  render: (row: any, _i, patch) => (
                    <Switch checked={!!row.isDefault} onCheckedChange={(v) => patch({isDefault: v})} />
                  ),
                },
              ]}
            />
          </FormRow>
          <FormRow labelKey="user:Affiliation">
            <Input value={user.affiliation ?? ""} onChange={(e) => updateField("affiliation", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:Title">
            <Input value={user.title ?? ""} onChange={(e) => updateField("title", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:Homepage">
            <Input value={user.homepage ?? ""} onChange={(e) => updateField("homepage", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:Bio">
            <Input value={user.bio ?? ""} onChange={(e) => updateField("bio", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:Gender">
            <SearchableSelect
              value={user.gender ?? ""}
              onChange={(v) => updateField("gender", v)}
              options={GENDERS.map((item) => ({value: item, label: item}))}
            />
          </FormRow>
          <FormRow labelKey="user:Birthday">
            <Input value={user.birthday ?? ""} onChange={(e) => updateField("birthday", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:Education">
            <Input value={user.education ?? ""} onChange={(e) => updateField("education", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:ID card type">
            <SearchableSelect
              value={user.idCardType ?? ""}
              onChange={(v) => updateField("idCardType", v)}
              options={ID_CARD_TYPES.map((item) => ({value: item, label: item}))}
            />
          </FormRow>
          <FormRow labelKey="user:ID card">
            <Input value={user.idCard ?? ""} onChange={(e) => updateField("idCard", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Real name">
            <Input value={user.realName ?? ""} onChange={(e) => updateField("realName", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:Language">
            <Input value={user.language ?? ""} onChange={(e) => updateField("language", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:Score">
            <Input
              type="number"
              value={user.score ?? 0}
              onChange={(e) => updateField("score", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="user:UID number">
            <Input value={user.uidNumber ?? ""} onChange={(e) => updateField("uidNumber", e.target.value)} />
          </FormRow>
          <FormRow labelKey="user:Ranking">
            <Input
              type="number"
              value={user.ranking ?? 0}
              onChange={(e) => updateField("ranking", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="user:Karma">
            <Input
              type="number"
              value={user.karma ?? 0}
              onChange={(e) => updateField("karma", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="organization:Balance credit">
            <Input
              type="number"
              step="0.01"
              value={user.balanceCredit ?? 0}
              onChange={(e) => updateField("balanceCredit", Number(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="user:Properties" block>
            <EditableTable
              rows={propertyRows}
              onChange={(rows) =>
                updateField(
                  "properties",
                  Object.fromEntries(rows.filter((row: any) => row.key).map((row: any) => [row.key, row.value])),
                )
              }
              newRow={() => ({key: "", value: ""})}
              reorderable={false}
              columns={[
                {
                  key: "key",
                  title: i18next.t("general:Name"),
                  width: 240,
                  render: (row: any, _i, patch) => (
                    <Input value={row.key ?? ""} onChange={(e) => patch({key: e.target.value})} />
                  ),
                },
                {
                  key: "value",
                  title: i18next.t("webhook:Value"),
                  render: (row: any, _i, patch) => (
                    <Input value={row.value ?? ""} onChange={(e) => patch({value: e.target.value})} />
                  ),
                },
              ]}
            />
          </FormRow>
          <FormRow labelKey="user:Balance">
            <div className="flex gap-2">
              <Input
                type="number"
                step="0.01"
                value={user.balance ?? 0}
                onChange={(e) => updateField("balance", Number(e.target.value))}
              />
              <div className="w-32 shrink-0">
                <SearchableSelect
                  value={user.balanceCurrency ?? "USD"}
                  onChange={(v) => updateField("balanceCurrency", v)}
                  options={(Setting.CurrencyOptions as any[]).map((item) => ({value: item.id, label: item.name}))}
                />
              </div>
            </div>
          </FormRow>
        </TabsContent>

        <TabsContent value="security">
          <FormRow labelKey="general:Password">
            <div className="flex gap-2">
              <Input
                type="password"
                autoComplete="new-password"
                value={newPassword}
                placeholder={i18next.t("user:New Password")}
                onChange={(e) => setNewPassword(e.target.value)}
              />
              <Button variant="outline" className="shrink-0" disabled={!newPassword} onClick={setPassword}>
                {i18next.t("user:Modify password...")}
              </Button>
            </div>
          </FormRow>
          <FormRow labelKey="user:Need update password">
            <Switch
              checked={!!user.needUpdatePassword}
              onCheckedChange={(v) => updateField("needUpdatePassword", v)}
            />
          </FormRow>
          <FormRow labelKey="user:Is admin">
            <Switch
              checked={!!user.isAdmin}
              disabled={!isAdmin}
              onCheckedChange={(v) => updateField("isAdmin", v)}
            />
          </FormRow>
          <FormRow labelKey="user:Is forbidden">
            <Switch
              checked={!!user.isForbidden}
              disabled={!isAdmin}
              onCheckedChange={(v) => updateField("isForbidden", v)}
            />
          </FormRow>
          <FormRow labelKey="user:Is deleted">
            <Switch
              checked={!!user.isDeleted}
              disabled={!isAdmin}
              onCheckedChange={(v) => updateField("isDeleted", v)}
            />
          </FormRow>
          <FormRow labelKey="user:Is verified">
            <Switch checked={!!user.isVerified} onCheckedChange={(v) => updateField("isVerified", v)} />
          </FormRow>
          <FormRow labelKey="general:IP whitelist">
            <TagsInput value={user.ipWhitelist ?? []} onChange={(v) => updateField("ipWhitelist", v)} />
          </FormRow>
          <FormRow labelKey="user:Deleted time">
            <Input value={Setting.getFormattedDate(user.deletedTime) ?? ""} disabled />
          </FormRow>
          <FormRow labelKey="general:MFA items" block>
            <EditableTable
              rows={user.mfaItems ?? []}
              onChange={(rows) => updateField("mfaItems", rows)}
              newRow={() => ({name: "Email", rule: "Optional"})}
              columns={[
                {
                  key: "name",
                  title: i18next.t("general:Name"),
                  width: 220,
                  render: (row: any, _i, patch) => (
                    <SearchableSelect
                      value={row.name}
                      onChange={(v) => patch({name: v})}
                      options={["Email", "SMS", "TOTP"].map((item) => ({value: item, label: item}))}
                    />
                  ),
                },
                {
                  key: "rule",
                  title: i18next.t("application:Rule"),
                  width: 220,
                  render: (row: any, _i, patch) => (
                    <SearchableSelect
                      value={row.rule}
                      onChange={(v) => patch({rule: v})}
                      options={["Optional", "Prompted", "Required"].map((item) => ({
                        value: item,
                        label: i18next.t(`general:${item}`),
                      }))}
                    />
                  ),
                },
              ]}
            />
          </FormRow>
          <FormRow labelKey="user:Managed accounts" block>
            <EditableTable
              rows={user.managedAccounts ?? []}
              onChange={(rows) => updateField("managedAccounts", rows)}
              newRow={() => ({application: "", username: "", password: "", signinUrl: ""})}
              reorderable={false}
              columns={[
                {
                  key: "application",
                  title: i18next.t("general:Application"),
                  width: 200,
                  render: (row: any, _i, patch) => (
                    <SearchableSelect
                      value={row.application}
                      onChange={(v) => patch({application: v})}
                      options={applications}
                    />
                  ),
                },
                {
                  key: "username",
                  title: i18next.t("signup:Username"),
                  width: 180,
                  render: (row: any, _i, patch) => (
                    <Input value={row.username ?? ""} onChange={(e) => patch({username: e.target.value})} />
                  ),
                },
                {
                  key: "password",
                  title: i18next.t("general:Password"),
                  width: 180,
                  render: (row: any, _i, patch) => (
                    <Input
                      type="password"
                      value={row.password ?? ""}
                      onChange={(e) => patch({password: e.target.value})}
                    />
                  ),
                },
                {
                  key: "signinUrl",
                  title: i18next.t("general:Signin URL"),
                  render: (row: any, _i, patch) => (
                    <Input value={row.signinUrl ?? ""} onChange={(e) => patch({signinUrl: e.target.value})} />
                  ),
                },
              ]}
            />
          </FormRow>
          <FormRow labelKey="user:MFA accounts" block>
            <EditableTable
              rows={user.mfaAccounts ?? []}
              onChange={(rows) => updateField("mfaAccounts", rows)}
              newRow={() => ({accountName: "", issuer: "", secretKey: "", origin: ""})}
              reorderable={false}
              columns={[
                {
                  key: "accountName",
                  title: i18next.t("user:MFA account name"),
                  width: 180,
                  render: (row: any, _i, patch) => (
                    <Input value={row.accountName ?? ""} onChange={(e) => patch({accountName: e.target.value})} />
                  ),
                },
                {
                  key: "issuer",
                  title: i18next.t("user:MFA account issuer"),
                  width: 180,
                  render: (row: any, _i, patch) => (
                    <Input value={row.issuer ?? ""} onChange={(e) => patch({issuer: e.target.value})} />
                  ),
                },
                {
                  key: "origin",
                  title: i18next.t("general:URL"),
                  width: 180,
                  render: (row: any, _i, patch) => (
                    <Input value={row.origin ?? ""} onChange={(e) => patch({origin: e.target.value})} />
                  ),
                },
                {
                  key: "secretKey",
                  title: i18next.t("user:MFA account secret key"),
                  render: (row: any, _i, patch) => (
                    <Input value={row.secretKey ?? ""} onChange={(e) => patch({secretKey: e.target.value})} />
                  ),
                },
              ]}
            />
          </FormRow>
          <FormRow labelKey="user:WebAuthn credentials" block>
            <div className="divide-y rounded-lg border">
              {(user.webauthnCredentials ?? []).length === 0 ? (
                <div className="p-4 text-center text-sm text-muted-foreground">{i18next.t("general:No data")}</div>
              ) : (
                (user.webauthnCredentials ?? []).map((item: any, index: number) => (
                  <div key={index} className="flex items-center gap-2 p-3 text-sm">
                    <span className="truncate font-mono text-xs">{item.ID ?? item.id}</span>
                  </div>
                ))
              )}
            </div>
          </FormRow>
          <FormRow labelKey="user:Face IDs" block>
            <div className="divide-y rounded-lg border">
              {(user.faceIds ?? []).length === 0 ? (
                <div className="p-4 text-center text-sm text-muted-foreground">{i18next.t("general:No data")}</div>
              ) : (
                (user.faceIds ?? []).map((item: any, index: number) => (
                  <div key={index} className="flex items-center gap-2 p-3 text-sm">
                    <span className="truncate">{item.name}</span>
                  </div>
                ))
              )}
            </div>
          </FormRow>
          <FormRow labelKey="user:Last change password time">
            <Input value={Setting.getFormattedDate(user.lastChangePasswordTime) ?? ""} disabled />
          </FormRow>
          <FormRow labelKey="mfa:Multi-factor authentication" block>
            <div className="space-y-2">
              <div className="divide-y rounded-lg border">
                {mfaItems.length === 0 ? (
                  <div className="p-4 text-center text-sm text-muted-foreground">
                    {i18next.t("general:No data")}
                  </div>
                ) : (
                  mfaItems.map((item: any) => (
                    <div key={item.mfaType} className="flex flex-wrap items-center gap-2 p-3">
                      <span className="text-sm font-medium">{item.mfaType}</span>
                      {item.secret ? (
                        <span className="truncate text-xs text-muted-foreground">{item.secret}</span>
                      ) : null}
                      <span className="flex-1" />
                      {item.enabled ? (
                        <>
                          <Badge variant="success">{i18next.t("general:Enabled")}</Badge>
                          {item.isPreferred ? (
                            <Badge>{i18next.t("mfa:preferred")}</Badge>
                          ) : (
                            <Button size="sm" onClick={() => setPreferredMfa(item.mfaType)}>
                              {i18next.t("mfa:Set preferred")}
                            </Button>
                          )}
                        </>
                      ) : (
                        <Badge variant="secondary">{i18next.t("general:Disabled")}</Badge>
                      )}
                      {isSelf ? (
                        <Button variant="outline" size="sm" asChild>
                          <Link to={`/mfa/setup?mfaType=${item.mfaType}`}>{i18next.t("general:Edit")}</Link>
                        </Button>
                      ) : null}
                    </div>
                  ))
                )}
              </div>
              <div className="flex gap-2">
                {isSelf ? (
                  <Button variant="outline" size="sm" asChild>
                    <Link to="/mfa/setup">{i18next.t("general:Enable")}</Link>
                  </Button>
                ) : null}
                {mfaItems.some((item: any) => item.enabled) ? (
                  <ConfirmButton
                    variant="destructive"
                    size="sm"
                    loading={removingMfa}
                    description={i18next.t("mfa:Multi-factor authentication")}
                    onConfirm={deleteMfa}
                  >
                    {i18next.t("general:Delete")}
                  </ConfirmButton>
                ) : null}
              </div>
            </div>
          </FormRow>
        </TabsContent>

        <TabsContent value="authorization">
          <FormRow labelKey="general:Roles">
            <div className="flex flex-wrap gap-1">
              {(user.roles ?? []).length === 0 ? (
                <span className="text-sm text-muted-foreground">{i18next.t("general:No data")}</span>
              ) : (
                (user.roles ?? []).map((role: any) => (
                  <Badge key={`${role.owner}/${role.name}`} variant="secondary">
                    {role.name}
                  </Badge>
                ))
              )}
            </div>
          </FormRow>
          <FormRow labelKey="general:Permissions">
            <div className="flex flex-wrap gap-1">
              {(user.permissions ?? []).length === 0 ? (
                <span className="text-sm text-muted-foreground">{i18next.t("general:No data")}</span>
              ) : (
                (user.permissions ?? []).map((permission: any) => (
                  <Badge key={`${permission.owner}/${permission.name}`} variant="secondary">
                    {permission.name}
                  </Badge>
                ))
              )}
            </div>
          </FormRow>
          <FormRow labelKey="user:Register type">
            <Input value={user.registerType ?? ""} disabled />
          </FormRow>
          <FormRow labelKey="user:Register source">
            <Input value={user.registerSource ?? ""} disabled />
          </FormRow>
          <FormRow labelKey="user:Last signin time">
            <Input value={Setting.getFormattedDate(user.lastSigninTime) ?? ""} disabled />
          </FormRow>
          <FormRow labelKey="user:Last signin IP">
            <Input value={user.lastSigninIp ?? ""} disabled />
          </FormRow>
        </TabsContent>
      </Tabs>
    </EditPageShell>
  );
}
