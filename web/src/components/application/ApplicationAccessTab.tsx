// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import * as React from "react";
import i18next from "i18next";
import {AlertTriangle, Eye, Pencil, Plus, Search, ShieldAlert, ShieldCheck, Trash2, UserPlus} from "lucide-react";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardDescription, CardHeader, CardTitle} from "@/components/ui/card";
import {Checkbox} from "@/components/ui/checkbox";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {RadioGroup, RadioGroupItem} from "@/components/ui/radio-group";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {SearchableSelect, type SearchableOption} from "@/components/common/SearchableSelect";
import {FormRow} from "@/components/crud/FormRow";
import {Link} from "react-router-dom";
import {AccessPreview} from "@/components/application/AccessPreview";
import {type AppRole, AppRoleDialog} from "@/components/application/AppRoleDialog";
import {useGroupOptions, useRoleOptions} from "@/hooks/use-options";
import * as ApplicationGrantBackend from "@/backend/ApplicationGrantBackend";
import * as Setting from "@/lib/setting";

interface ApplicationAccessTabProps {
  application: any;
  account: any;
  organizations: SearchableOption[];
  updateField: (field: string, value: any) => void;
}

const toRfc3339 = (value: string): string | null => {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date.toISOString().replace(/\.\d{3}Z$/, "Z");
};

const toLocalInput = (value: string): string => {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
};

const isExpired = (value: string) => !!value && new Date(value).getTime() <= Date.now();

/** The application roles a grant assigns, privileged ones highlighted, missing ones flagged. */
function RoleChips({names, appRoles}: {names: string[]; appRoles: AppRole[]}) {
  if (!names?.length) {
    return <span className="text-sm text-muted-foreground">{i18next.t("application:Sign-in only")}</span>;
  }
  return (
    <div className="flex flex-wrap gap-1">
      {names.map((name) => {
        const role = appRoles.find((item) => item.name === name);
        if (!role) {
          return <Badge key={name} variant="destructive" title={i18next.t("application:This role no longer exists, the grant adds nothing for it")}>{name}?</Badge>;
        }
        return (
          <Badge key={name} variant={role.isPrivileged ? "destructive" : "secondary"} className="gap-1">
            {role.isPrivileged ? <ShieldAlert className="h-3 w-3" /> : null}
            {role.displayName || role.name}
          </Badge>
        );
      })}
    </div>
  );
}

/** Choosing application roles; privileged ones are marked and explained. */
function RolePicker({appRoles, value, onChange}: {appRoles: AppRole[]; value: string[]; onChange: (value: string[]) => void}) {
  if (appRoles.length === 0) {
    return <p className="text-sm text-muted-foreground">{i18next.t("application:This application has no roles yet. Add one under Application Roles.")}</p>;
  }
  return (
    <div className="space-y-1 rounded-md border p-2">
      {appRoles.map((role) => (
        <label key={role.name} className="flex cursor-pointer items-start gap-2 rounded p-1 text-sm hover:bg-muted/50">
          <Checkbox
            className="mt-0.5"
            checked={value.includes(role.name)}
            onCheckedChange={(checked) => onChange(checked === true ? [...value, role.name] : value.filter((name) => name !== role.name))}
          />
          <span className="min-w-0">
            <span className="font-medium">{role.displayName || role.name}</span>
            {role.isPrivileged ? <Badge variant="destructive" className="ml-2 gap-1"><ShieldAlert className="h-3 w-3" />{i18next.t("application:Privileged")}</Badge> : null}
            {role.description ? <span className="block text-muted-foreground">{role.description}</span> : null}
          </span>
        </label>
      ))}
    </div>
  );
}

function PrivilegedWarning({roles, appRoles}: {roles: string[]; appRoles: AppRole[]}) {
  const privileged = appRoles.filter((role) => role.isPrivileged && roles.includes(role.name));
  if (privileged.length === 0) {
    return null;
  }
  return (
    <div className="flex gap-2 rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm">
      <AlertTriangle className="h-4 w-4 shrink-0 text-destructive" />
      <span>
        {i18next.t("application:Privileged roles give administrative rights in the application")}
        {" ("}{privileged.map((role) => role.displayName || role.name).join(", ")}{"). "}
        {i18next.t("application:Prefer a user over a group and set an expire time.")}
      </span>
    </div>
  );
}

/** Edits what a grant gives: its roles, expiry and description. Whom it is for can't change. */
function GrantEditDialog({grant, appRoles, organization, onClose, onSaved}: {grant: any; appRoles: AppRole[]; organization: string; onClose: () => void; onSaved: () => void}) {
  const [roles, setRoles] = React.useState<string[]>([]);
  const [expireTime, setExpireTime] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [saving, setSaving] = React.useState(false);

  React.useEffect(() => {
    if (grant) {
      setRoles(grant.roles ?? []);
      setExpireTime(toLocalInput(grant.expireTime));
      setDescription(grant.description ?? "");
    }
  }, [grant]);

  if (!grant) {
    return null;
  }
  const isInternal = grant.principal?.split("/")[0] === organization;

  const save = () => {
    const expire = toRfc3339(expireTime);
    if (expire === null) {
      Setting.showMessage("error", i18next.t("application:Invalid expire time"));
      return;
    }
    setSaving(true);
    ApplicationGrantBackend.updateApplicationGrant({...grant, roles, expireTime: expire, description: description.trim()})
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          onSaved();
          onClose();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .finally(() => setSaving(false));
  };

  return (
    <Dialog open={!!grant} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{i18next.t("application:Edit Grant")}</DialogTitle>
          <DialogDescription className="font-mono">{grant.principal}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <FormRow label={i18next.t("application:Application Roles")}>
            <RolePicker appRoles={appRoles} value={roles} onChange={setRoles} />
            {isInternal && roles.length === 0 ? <p className="mt-1 text-xs text-destructive">{i18next.t("application:A grant for the application's own organization needs at least one role")}</p> : null}
          </FormRow>
          <PrivilegedWarning roles={roles} appRoles={appRoles} />
          <FormRow label={i18next.t("application:Expire Time")}>
            <Input className="w-64" type="datetime-local" value={expireTime} onChange={(e) => setExpireTime(e.target.value)} />
          </FormRow>
          <FormRow label={i18next.t("general:Description")}>
            <Input value={description} maxLength={255} onChange={(e) => setDescription(e.target.value)} />
          </FormRow>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>{i18next.t("general:Cancel")}</Button>
          <Button type="button" onClick={save} disabled={saving || (isInternal && roles.length === 0)}>{i18next.t("general:Save")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

/**
 * Who may sign in to this application from other organizations, and what the application is told
 * about them. Application roles are reusable entitlements of this application; grants give access
 * to users or groups of other organizations, optionally with roles, and give roles to users or
 * groups of the application's own organization.
 */
export function ApplicationAccessTab({application, account, organizations, updateField}: ApplicationAccessTabProps) {
  const organization = application.organization;
  const appRoles: AppRole[] = application.appRoles ?? [];
  const isGlobalAdmin = Setting.isAdminUser(account);

  const groupOptions = useGroupOptions(organization);
  const roleOptions = useRoleOptions(organization);

  const [grants, setGrants] = React.useState<any[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [filter, setFilter] = React.useState("");

  const [roleDialog, setRoleDialog] = React.useState<{open: boolean; role: AppRole | null}>({open: false, role: null});

  const [principalType, setPrincipalType] = React.useState("User");
  const [principalOrganization, setPrincipalOrganization] = React.useState("");
  const [query, setQuery] = React.useState("");
  const [candidates, setCandidates] = React.useState<any[] | null>(null);
  const [principal, setPrincipal] = React.useState("");
  const [accessMode, setAccessMode] = React.useState<"signin" | "roles">("signin");
  const [selectedRoles, setSelectedRoles] = React.useState<string[]>([]);
  const [expireTime, setExpireTime] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [draftPreview, setDraftPreview] = React.useState<any>(null);
  const [saving, setSaving] = React.useState(false);

  const [editing, setEditing] = React.useState<any>(null);
  const [explained, setExplained] = React.useState<{user: string; preview: any} | null>(null);
  const [checkUser, setCheckUser] = React.useState("");

  const isInternalPrincipal = principalOrganization === organization;
  const grantRoles = isInternalPrincipal || accessMode === "roles" ? selectedRoles : [];

  const reload = React.useCallback(() => {
    setLoading(true);
    ApplicationGrantBackend.getApplicationGrants(organization, application.name)
      .then((res: any) => {
        if (res.status === "ok") {
          setGrants(res.data ?? []);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      })
      .finally(() => setLoading(false));
  }, [organization, application.name]);

  React.useEffect(() => {
    reload();
  }, [reload]);

  const usage = React.useMemo(() => {
    const res: Record<string, number> = {};
    grants.forEach((grant) => (grant.roles ?? []).forEach((name: string) => (res[name] = (res[name] ?? 0) + 1)));
    return res;
  }, [grants]);

  const saveRoles = async(next: AppRole[]): Promise<boolean> => {
    const res: any = await ApplicationGrantBackend.updateApplicationRoles(organization, application.name, next);
    if (res.status !== "ok") {
      Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
      return false;
    }
    // keep the edit page's copy current, so its own Save does not write the old roles back
    updateField("appRoles", next);
    Setting.showMessage("success", i18next.t("general:Successfully saved"));
    return true;
  };

  const saveRole = (role: AppRole) => {
    const next = roleDialog.role
      ? appRoles.map((item) => (item.name === roleDialog.role?.name ? role : item))
      : [...appRoles, role];
    return saveRoles(next);
  };

  const resetCandidate = () => {
    setCandidates(null);
    setPrincipal("");
    setDraftPreview(null);
  };

  const search = () => {
    if (!principalOrganization || !query.trim()) {
      return;
    }
    ApplicationGrantBackend.getApplicationGrantPrincipals(organization, application.name, principalType, principalOrganization, query.trim())
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", res.msg);
          return;
        }
        const items = res.data ?? [];
        setCandidates(items);
        setPrincipal(items.length === 1 ? items[0].id : "");
        setDraftPreview(null);
      });
  };

  const previewDraft = () => {
    const user = principalType === "User" ? principal : "";
    ApplicationGrantBackend.getApplicationAccessPreview(organization, application.name, user, grantRoles, true).then((res: any) => {
      if (res.status === "ok") {
        setDraftPreview(res.data);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  const grant = () => {
    if (!principal) {
      return;
    }
    const expire = toRfc3339(expireTime);
    if (expire === null) {
      Setting.showMessage("error", i18next.t("application:Invalid expire time"));
      return;
    }
    setSaving(true);
    ApplicationGrantBackend.addApplicationGrant({
      owner: organization,
      application: application.name,
      principalType,
      principal,
      roles: grantRoles,
      expireTime: expire,
      description: description.trim(),
    })
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("application:Access granted"));
          resetCandidate();
          setQuery("");
          setSelectedRoles([]);
          setAccessMode("signin");
          setExpireTime("");
          setDescription("");
          reload();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .finally(() => setSaving(false));
  };

  const revoke = async(item: any) => {
    const res: any = await ApplicationGrantBackend.deleteApplicationGrant(item);
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("application:Access revoked"));
      reload();
    } else {
      Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
    }
  };

  const explain = (user: string) => {
    if (!user.includes("/")) {
      Setting.showMessage("error", i18next.t("application:Enter the user as organization/name"));
      return;
    }
    ApplicationGrantBackend.getApplicationAccessPreview(organization, application.name, user).then((res: any) => {
      if (res.status === "ok") {
        setExplained({user, preview: res.data});
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  if (application.isShared) {
    return (
      <div className="col-span-full rounded-md border p-4 text-sm text-muted-foreground">
        {i18next.t("application:A shared application is available to every organization and does not take grants")}
      </div>
    );
  }

  const organizationOptions = organizations.some((item) => item.value === organization)
    ? organizations
    : [{value: organization, label: organization}, ...organizations];
  const shownGrants = grants.filter((item) => !filter || `${item.principal} ${item.description ?? ""} ${(item.roles ?? []).join(" ")}`.toLowerCase().includes(filter.toLowerCase()));
  const grantButton = (
    <>
      <UserPlus className="mr-1 h-4 w-4" />
      {i18next.t("application:Grant Access")}
    </>
  );
  const privilegedSelected = appRoles.some((role) => role.isPrivileged && grantRoles.includes(role.name));

  return (
    <div className="col-span-full space-y-6">
      <div className="flex gap-3 rounded-md border bg-muted/40 p-4 text-sm">
        <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
        <div className="space-y-1">
          <p className="text-muted-foreground">
            {i18next.t("application:Existing users or groups of other organizations can sign in to this application with their own account, password and MFA. The application keeps its organization and its client ID. A grant only allows the sign-in. What the user may do in the application is still decided by the application.")}
          </p>
          <p className="text-muted-foreground">
            {i18next.t("application:Application roles optionally tell the application what a granted user may do, by adding groups, roles or claims to this application's tokens only. They never make anyone a member of a group, a role or an organization.")}
          </p>
        </div>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-start justify-between gap-4 space-y-0">
          <div>
            <CardTitle className="text-base">{i18next.t("application:Application Roles")}</CardTitle>
            <CardDescription>{i18next.t("application:Reusable entitlements of this application, assigned through grants.")}</CardDescription>
          </div>
          <Button type="button" size="sm" variant="outline" onClick={() => setRoleDialog({open: true, role: null})}>
            <Plus className="mr-1 h-4 w-4" />
            {i18next.t("application:Add Role")}
          </Button>
        </CardHeader>
        <CardContent>
          {appRoles.length === 0 ? (
            <p className="text-sm text-muted-foreground">{i18next.t("application:No application roles. Grants only allow the sign-in.")}</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{i18next.t("general:Name")}</TableHead>
                  <TableHead>{i18next.t("application:Adds to the tokens of this application")}</TableHead>
                  <TableHead>{i18next.t("application:Grants")}</TableHead>
                  <TableHead className="w-28">{i18next.t("general:Action")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {appRoles.map((role) => (
                  <TableRow key={role.name}>
                    <TableCell>
                      <div className="font-medium">{role.displayName || role.name}</div>
                      <div className="font-mono text-xs text-muted-foreground">{role.name}</div>
                      {role.isPrivileged ? <Badge variant="destructive" className="mt-1 gap-1"><ShieldAlert className="h-3 w-3" />{i18next.t("application:Privileged")}</Badge> : null}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1 text-xs">
                        {(role.groups ?? []).map((id) => <Badge key={id} variant="outline" className="font-mono font-normal">groups ← {id}</Badge>)}
                        {(role.roles ?? []).map((id) => <Badge key={id} variant="outline" className="font-mono font-normal">roles ← {id}</Badge>)}
                        {(role.claims ?? []).map((claim) => <Badge key={claim.name + claim.value} variant="outline" className="font-mono font-normal">{claim.name} = {claim.value}{claim.type === "Array" ? " []" : ""}</Badge>)}
                      </div>
                    </TableCell>
                    <TableCell>{usage[role.name] ?? 0}</TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button type="button" size="icon" variant="ghost" aria-label={i18next.t("general:Edit")} onClick={() => setRoleDialog({open: true, role})}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <ConfirmButton
                          size="icon"
                          variant="ghost"
                          aria-label={i18next.t("general:Delete")}
                          disabled={(usage[role.name] ?? 0) > 0}
                          title={(usage[role.name] ?? 0) > 0 ? i18next.t("application:Still assigned by grants") : undefined}
                          onConfirm={async() => {
                            await saveRoles(appRoles.filter((item) => item.name !== role.name));
                          }}
                        >
                          <Trash2 className="h-4 w-4" />
                        </ConfirmButton>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{i18next.t("application:Grant Access")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-[12rem_1fr]">
            <FormRow label={i18next.t("general:Type")}>
              <Select
                value={principalType}
                onValueChange={(value) => {
                  setPrincipalType(value);
                  resetCandidate();
                }}
              >
                <SelectTrigger aria-label={i18next.t("general:Type")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="User">{i18next.t("general:User")}</SelectItem>
                  <SelectItem value="Group">{i18next.t("subscription:Group")}</SelectItem>
                </SelectContent>
              </Select>
            </FormRow>
            <FormRow label={i18next.t("general:Organization")}>
              {isGlobalAdmin ? (
                <SearchableSelect
                  value={principalOrganization}
                  onChange={(value) => {
                    setPrincipalOrganization(value);
                    resetCandidate();
                  }}
                  options={organizationOptions}
                  placeholder={i18next.t("application:Organization of the user or group")}
                />
              ) : (
                <Input
                  value={principalOrganization}
                  onChange={(e) => {
                    setPrincipalOrganization(e.target.value.trim());
                    resetCandidate();
                  }}
                  placeholder={i18next.t("application:Organization of the user or group")}
                />
              )}
            </FormRow>
          </div>
          {principalType === "Group" && !isInternalPrincipal ? (
            <p className="text-sm text-muted-foreground">
              {i18next.t("application:Everyone the other organization puts into this group or one of its subgroups gets access, including through default groups of its sign-up or sync settings.")}
            </p>
          ) : null}

          <FormRow
            label={principalType === "User" ? i18next.t("general:User") : i18next.t("subscription:Group")}
            tooltip={isGlobalAdmin || isInternalPrincipal ? undefined : i18next.t("application:Enter the exact name, members of other organizations are not listed")}
          >
            <div className="flex flex-wrap gap-2">
              <Input
                className="min-w-[14rem] flex-1"
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  resetCandidate();
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    search();
                  }
                }}
                placeholder={isGlobalAdmin || isInternalPrincipal ? i18next.t("application:Name or Display Name") : i18next.t("application:Exact Name")}
              />
              <Button type="button" variant="outline" onClick={search} disabled={!principalOrganization || !query.trim()}>
                <Search className="mr-1 h-4 w-4" />
                {i18next.t("general:Search")}
              </Button>
            </div>
            {candidates !== null && (
              <div className="mt-2 rounded-md border">
                {candidates.length === 0 ? (
                  <p className="p-3 text-sm text-muted-foreground">{i18next.t("application:No matching user or group")}</p>
                ) : candidates.map((item: any) => (
                  <label key={item.id} className="flex cursor-pointer items-center gap-2 border-b p-2 text-sm last:border-b-0">
                    <input type="radio" name="grant-principal" checked={principal === item.id} onChange={() => {
                      setPrincipal(item.id);
                      setDraftPreview(null);
                    }} />
                    <span className="font-mono">{item.id}</span>
                    {item.displayName ? <span className="text-muted-foreground">{item.displayName}</span> : null}
                    <Badge variant="outline" className="ml-auto">
                      {isInternalPrincipal ? i18next.t("application:This application's organization") : `${i18next.t("application:Home organization")}: ${principalOrganization}`}
                    </Badge>
                  </label>
                ))}
              </div>
            )}
          </FormRow>

          <FormRow label={i18next.t("application:Access")}>
            {isInternalPrincipal ? (
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">{i18next.t("application:Users of this application's organization already sign in. Choose the application roles to give them.")}</p>
                <RolePicker appRoles={appRoles} value={selectedRoles} onChange={setSelectedRoles} />
              </div>
            ) : (
              <div className="space-y-2">
                <RadioGroup value={accessMode} onValueChange={(value) => setAccessMode(value as "signin" | "roles")} className="gap-2">
                  <label className="flex items-start gap-2 text-sm">
                    <RadioGroupItem value="signin" className="mt-0.5" />
                    <span><span className="font-medium">{i18next.t("application:Sign-in only")}</span><span className="block text-muted-foreground">{i18next.t("application:The application decides on its own what the user may do. Nothing is added to the tokens.")}</span></span>
                  </label>
                  <label className="flex items-start gap-2 text-sm">
                    <RadioGroupItem value="roles" className="mt-0.5" />
                    <span><span className="font-medium">{i18next.t("application:Sign-in and application roles")}</span><span className="block text-muted-foreground">{i18next.t("application:The application also receives the groups, roles or claims of the chosen roles.")}</span></span>
                  </label>
                </RadioGroup>
                {accessMode === "roles" ? <RolePicker appRoles={appRoles} value={selectedRoles} onChange={setSelectedRoles} /> : null}
              </div>
            )}
          </FormRow>
          <PrivilegedWarning roles={grantRoles} appRoles={appRoles} />

          <div className="grid gap-4 md:grid-cols-[16rem_1fr]">
            <FormRow label={i18next.t("application:Expire Time")} tooltip={i18next.t("application:Optional, the grant stops working at this time")}>
              <Input type="datetime-local" value={expireTime} onChange={(e) => setExpireTime(e.target.value)} />
            </FormRow>
            <FormRow label={i18next.t("general:Description")}>
              <Input value={description} maxLength={255} onChange={(e) => setDescription(e.target.value)} placeholder={i18next.t("application:Reason or ticket reference")} />
            </FormRow>
          </div>

          {draftPreview ? <AccessPreview preview={draftPreview} /> : null}

          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="outline" onClick={previewDraft} disabled={!principal || (isInternalPrincipal && grantRoles.length === 0)}>
              <Eye className="mr-1 h-4 w-4" />
              {i18next.t("application:Preview Effective Access")}
            </Button>
            {privilegedSelected ? (
              <ConfirmButton
                destructive
                disabled={!principal || saving}
                title={i18next.t("application:Grant privileged access") + "?"}
                description={`${principal}: ${i18next.t("application:Privileged roles give administrative rights in the application")}`}
                confirmText={i18next.t("application:Grant Access")}
                onConfirm={async() => grant()}
              >
                {grantButton}
              </ConfirmButton>
            ) : (
              <Button type="button" onClick={grant} disabled={!principal || saving || (isInternalPrincipal && grantRoles.length === 0) || (accessMode === "roles" && !isInternalPrincipal && grantRoles.length === 0)}>
                {grantButton}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-start justify-between gap-4 space-y-0">
          <div>
            <CardTitle className="text-base">{i18next.t("application:Granted Access")}</CardTitle>
            <CardDescription>
              {i18next.t("application:Every change is recorded in the audit log with who made it and when.")}{" "}
              <Link className="underline" to="/records">{i18next.t("application:Open the audit log")}</Link>
            </CardDescription>
          </div>
          <Input className="w-56" value={filter} onChange={(e) => setFilter(e.target.value)} placeholder={i18next.t("general:Search")} />
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{i18next.t("application:Principal")}</TableHead>
                  <TableHead>{i18next.t("application:Access")}</TableHead>
                  <TableHead>{i18next.t("application:Expire Time")}</TableHead>
                  <TableHead>{i18next.t("application:Granted By")}</TableHead>
                  <TableHead>{i18next.t("general:Description")}</TableHead>
                  <TableHead className="w-32">{i18next.t("general:Action")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {shownGrants.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center text-sm text-muted-foreground">
                      {loading ? i18next.t("application:Loading") : i18next.t("application:No user or group of another organization has access")}
                    </TableCell>
                  </TableRow>
                ) : shownGrants.map((item: any) => {
                  const home = item.principal.split("/")[0];
                  return (
                    <TableRow key={item.name}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Badge variant="secondary">{item.principalType === "Group" ? i18next.t("subscription:Group") : i18next.t("general:User")}</Badge>
                          <span className="font-mono text-xs">{item.principal}</span>
                        </div>
                        <div className="mt-1 text-xs text-muted-foreground">
                          {home === organization ? i18next.t("application:This application's organization") : `${i18next.t("application:Home organization")}: ${home}`}
                        </div>
                      </TableCell>
                      <TableCell><RoleChips names={item.roles ?? []} appRoles={appRoles} /></TableCell>
                      <TableCell>
                        {item.expireTime ? (
                          <span className={isExpired(item.expireTime) ? "text-destructive" : undefined}>
                            {Setting.getFormattedDate(item.expireTime)}
                            {isExpired(item.expireTime) ? ` (${i18next.t("subscription:Expired")})` : ""}
                          </span>
                        ) : i18next.t("invitation:Never")}
                      </TableCell>
                      <TableCell className="text-xs">
                        <div className="font-mono">{item.createdBy}</div>
                        <div className="text-muted-foreground">{Setting.getFormattedDate(item.createdTime)}</div>
                        {item.updatedTime ? (
                          <div className="text-muted-foreground">{i18next.t("application:Changed")} {Setting.getFormattedDate(item.updatedTime)} ({item.updatedBy})</div>
                        ) : null}
                      </TableCell>
                      <TableCell className="max-w-xs truncate">{item.description}</TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          {item.principalType === "User" ? (
                            <Button type="button" size="icon" variant="ghost" aria-label={i18next.t("application:Explain Access")} title={i18next.t("application:Explain Access")} onClick={() => explain(item.principal)}>
                              <Eye className="h-4 w-4" />
                            </Button>
                          ) : null}
                          <Button type="button" size="icon" variant="ghost" aria-label={i18next.t("general:Edit")} title={i18next.t("general:Edit")} onClick={() => setEditing(item)}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <ConfirmButton
                            size="icon"
                            variant="ghost"
                            aria-label={i18next.t("application:Revoke Access")}
                            title={i18next.t("application:Revoke Access") + "?"}
                            description={item.principal}
                            confirmText={i18next.t("application:Revoke Access")}
                            onConfirm={() => revoke(item)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </ConfirmButton>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{i18next.t("application:Check Access")}</CardTitle>
          <CardDescription>{i18next.t("application:Explain how a user signs in to this application and what its tokens say about them.")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-wrap gap-2">
            <Input className="min-w-[14rem] flex-1 font-mono" value={checkUser} onChange={(e) => setCheckUser(e.target.value.trim())} placeholder="organization/username" onKeyDown={(e) => e.key === "Enter" && explain(checkUser)} />
            <Button type="button" variant="outline" onClick={() => explain(checkUser)} disabled={!checkUser}>
              <Eye className="mr-1 h-4 w-4" />
              {i18next.t("application:Explain Access")}
            </Button>
          </div>
          {explained ? <AccessPreview preview={explained.preview} /> : null}
        </CardContent>
      </Card>

      <AppRoleDialog
        open={roleDialog.open}
        onOpenChange={(open) => setRoleDialog((current) => ({...current, open}))}
        role={roleDialog.role}
        organization={organization}
        groupOptions={groupOptions}
        roleOptions={roleOptions}
        existingNames={appRoles.map((role) => role.name)}
        onSave={saveRole}
      />
      <GrantEditDialog grant={editing} appRoles={appRoles} organization={organization} onClose={() => setEditing(null)} onSaved={reload} />
    </div>
  );
}
