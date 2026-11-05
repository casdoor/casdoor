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
import {Search, ShieldCheck, Trash2, UserPlus} from "lucide-react";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {SearchableSelect, type SearchableOption} from "@/components/common/SearchableSelect";
import {FormRow} from "@/components/crud/FormRow";
import * as ApplicationGrantBackend from "@/backend/ApplicationGrantBackend";
import * as Setting from "@/lib/setting";

interface ApplicationAccessTabProps {
  application: any;
  account: any;
  organizations: SearchableOption[];
}

/**
 * Grants that let existing users or groups of other organizations sign in to this
 * application through its own OAuth/OIDC client. A global admin may search the other
 * organizations; an admin of the application's organization names the principal exactly.
 */
export function ApplicationAccessTab({application, account, organizations}: ApplicationAccessTabProps) {
  const [grants, setGrants] = React.useState<any[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [principalType, setPrincipalType] = React.useState("User");
  const [principalOrganization, setPrincipalOrganization] = React.useState("");
  const [query, setQuery] = React.useState("");
  const [candidates, setCandidates] = React.useState<any[] | null>(null);
  const [principal, setPrincipal] = React.useState("");
  const [expireTime, setExpireTime] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [saving, setSaving] = React.useState(false);

  const isGlobalAdmin = Setting.isAdminUser(account);
  const organization = application.organization;
  const otherOrganizations = organizations.filter((item) => item.value !== organization);

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

  const resetCandidate = () => {
    setCandidates(null);
    setPrincipal("");
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
      });
  };

  const grant = () => {
    if (!principal) {
      return;
    }
    let expire = "";
    if (expireTime) {
      const date = new Date(expireTime);
      if (Number.isNaN(date.getTime())) {
        Setting.showMessage("error", i18next.t("application:Invalid expire time"));
        return;
      }
      expire = date.toISOString().replace(/\.\d{3}Z$/, "Z");
    }
    setSaving(true);
    ApplicationGrantBackend.addApplicationGrant({
      owner: organization,
      application: application.name,
      principalType,
      principal,
      expireTime: expire,
      description: description.trim(),
    })
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("application:Access granted"));
          resetCandidate();
          setQuery("");
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

  if (application.isShared) {
    return (
      <div className="col-span-full rounded-md border p-4 text-sm text-muted-foreground">
        {i18next.t("application:A shared application is available to every organization and does not take grants")}
      </div>
    );
  }

  const isExpired = (value: string) => !!value && new Date(value).getTime() <= Date.now();

  return (
    <div className="col-span-full space-y-6">
      <div className="flex gap-3 rounded-md border bg-muted/40 p-4 text-sm">
        <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
        <div className="space-y-1">
          <p className="font-medium">{i18next.t("application:Cross-Organization Access")}</p>
          <p className="text-muted-foreground">
            {i18next.t("application:Existing users or groups of other organizations can sign in to this application with their own account, password and MFA. The application keeps its organization and its client ID. A grant only allows the sign-in. What the user may do in the application is still decided by the application.")}
          </p>
          <p className="text-muted-foreground">
            {i18next.t("application:A granted user is named username@organization in the tokens and in the userinfo, and their groups, roles and permissions carry their organization. Revoking a grant stops new sign-ins, token refreshes and userinfo at once; tokens already issued expire at the latest after the application's token lifetime.")}
          </p>
        </div>
      </div>

      <div className="space-y-3">
        <h3 className="text-base font-semibold">{i18next.t("application:Grant Access")}</h3>
        <FormRow label={i18next.t("general:Type")}>
          <Select
            value={principalType}
            onValueChange={(value) => {
              setPrincipalType(value);
              resetCandidate();
            }}
          >
            <SelectTrigger className="w-48" aria-label={i18next.t("general:Type")}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="User">{i18next.t("general:User")}</SelectItem>
              <SelectItem value="Group">{i18next.t("subscription:Group")}</SelectItem>
            </SelectContent>
          </Select>
          {principalType === "Group" && (
            <p className="mt-2 text-sm text-muted-foreground">
              {i18next.t("application:Everyone the other organization puts into this group or one of its subgroups gets access, including through default groups of its sign-up or sync settings.")}
            </p>
          )}
        </FormRow>
        <FormRow label={i18next.t("general:Organization")}>
          {isGlobalAdmin ? (
            <SearchableSelect
              value={principalOrganization}
              onChange={(value) => {
                setPrincipalOrganization(value);
                resetCandidate();
              }}
              options={otherOrganizations}
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
        <FormRow
          label={principalType === "User" ? i18next.t("general:User") : i18next.t("subscription:Group")}
          tooltip={isGlobalAdmin ? undefined : i18next.t("application:Enter the exact name, members of other organizations are not listed")}
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
              placeholder={isGlobalAdmin ? i18next.t("application:Name or Display Name") : i18next.t("application:Exact Name")}
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
                  <input
                    type="radio"
                    name="grant-principal"
                    checked={principal === item.id}
                    onChange={() => setPrincipal(item.id)}
                  />
                  <span className="font-mono">{item.id}</span>
                  {item.displayName ? <span className="text-muted-foreground">{item.displayName}</span> : null}
                </label>
              ))}
            </div>
          )}
        </FormRow>
        <FormRow label={i18next.t("application:Expire Time")} tooltip={i18next.t("application:Optional, the grant stops working at this time")}>
          <Input className="w-64" type="datetime-local" value={expireTime} onChange={(e) => setExpireTime(e.target.value)} />
        </FormRow>
        <FormRow label={i18next.t("general:Description")}>
          <Input value={description} maxLength={255} onChange={(e) => setDescription(e.target.value)} placeholder={i18next.t("application:Reason or ticket reference")} />
        </FormRow>
        <div>
          <Button type="button" onClick={grant} disabled={!principal || saving}>
            <UserPlus className="mr-1 h-4 w-4" />
            {i18next.t("application:Grant Access")}
          </Button>
        </div>
      </div>

      <div className="space-y-2">
        <h3 className="text-base font-semibold">{i18next.t("application:Granted Access")}</h3>
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{i18next.t("general:Type")}</TableHead>
                <TableHead>{i18next.t("application:Principal")}</TableHead>
                <TableHead>{i18next.t("application:Expire Time")}</TableHead>
                <TableHead>{i18next.t("application:Granted By")}</TableHead>
                <TableHead>{i18next.t("general:Created time")}</TableHead>
                <TableHead>{i18next.t("general:Description")}</TableHead>
                <TableHead className="w-24">{i18next.t("general:Action")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {grants.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-sm text-muted-foreground">
                    {loading ? i18next.t("application:Loading") : i18next.t("application:No user or group of another organization has access")}
                  </TableCell>
                </TableRow>
              ) : grants.map((item: any) => (
                <TableRow key={item.name}>
                  <TableCell>
                    <Badge variant="secondary">{item.principalType === "Group" ? i18next.t("subscription:Group") : i18next.t("general:User")}</Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs">{item.principal}</TableCell>
                  <TableCell>
                    {item.expireTime ? (
                      <span className={isExpired(item.expireTime) ? "text-destructive" : undefined}>
                        {Setting.getFormattedDate(item.expireTime)}
                        {isExpired(item.expireTime) ? ` (${i18next.t("subscription:Expired")})` : ""}
                      </span>
                    ) : i18next.t("invitation:Never")}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{item.createdBy}</TableCell>
                  <TableCell>{Setting.getFormattedDate(item.createdTime)}</TableCell>
                  <TableCell className="max-w-xs truncate">{item.description}</TableCell>
                  <TableCell>
                    <ConfirmButton
                      size="sm"
                      variant="outline"
                      title={i18next.t("application:Revoke Access") + "?"}
                      description={item.principal}
                      confirmText={i18next.t("application:Revoke Access")}
                      onConfirm={() => revoke(item)}
                    >
                      <Trash2 className="mr-1 h-4 w-4" />
                      {i18next.t("application:Revoke")}
                    </ConfirmButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
