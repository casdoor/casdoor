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
import {AlertTriangle} from "lucide-react";
import {Badge} from "@/components/ui/badge";

/** One value of a claim, marked when an application role added it. */
function ValueChip({value, projected}: {value: string; projected: boolean}) {
  return (
    <Badge variant={projected ? "default" : "secondary"} className="font-mono text-xs font-normal" title={projected ? i18next.t("application:Added by an application role") : i18next.t("application:The user's own value")}>
      {value}
    </Badge>
  );
}

function ClaimRow({label, children}: {label: string; children: React.ReactNode}) {
  return (
    <div className="grid grid-cols-[9rem_1fr] gap-2 py-1 text-sm">
      <span className="text-muted-foreground">{label}</span>
      <div className="flex min-w-0 flex-wrap gap-1">{children}</div>
    </div>
  );
}

/**
 * What a user's sign-in to the application looks like: how access is given, which grants and roles
 * apply, and the values of the application's claims, the user's own ones apart from the projected
 * ones. Without a user it shows what a set of roles adds.
 */
export function AccessPreview({preview}: {preview: any}) {
  if (!preview) {
    return null;
  }
  const entitlements = preview.entitlements ?? {};
  const projectedGroups: string[] = entitlements.groups ?? [];
  const projectedRoles: string[] = entitlements.roleNames ?? [];
  const claims: Record<string, any> = entitlements.claims ?? {};
  const groups: string[] = preview.groups ?? projectedGroups;
  const roles: string[] = preview.roles ?? projectedRoles;
  const warnings: string[] = entitlements.warnings ?? [];
  const accessLabel: Record<string, string> = {
    Native: i18next.t("application:Signs in as a user of the application's organization"),
    Grant: i18next.t("application:Signs in through an application grant"),
    Denied: i18next.t("application:Can't sign in"),
    Pending: i18next.t("application:Gets access with this grant"),
  };

  return (
    <div className="space-y-2 rounded-md border p-3">
      {preview.user ? (
        <ClaimRow label={i18next.t("general:User")}>
          <span className="font-mono text-xs">{preview.user}</span>
          <Badge variant="outline">{i18next.t("application:Home organization")}: {preview.homeOrganization}</Badge>
        </ClaimRow>
      ) : null}
      {preview.access ? (
        <ClaimRow label={i18next.t("application:Access")}>
          <Badge variant={preview.access === "Denied" ? "destructive" : "secondary"}>{accessLabel[preview.access] ?? preview.access}</Badge>
          {preview.reason ? <span className="text-xs text-muted-foreground">{preview.reason}</span> : null}
        </ClaimRow>
      ) : null}
      {(preview.access === "Denied" || preview.access === "Pending") && (entitlements.roles ?? []).length > 0 ? (
        <p className="text-xs text-muted-foreground">{i18next.t("application:The values below apply once the user has access.")}</p>
      ) : null}
      {(entitlements.grants ?? []).length > 0 ? (
        <ClaimRow label={i18next.t("application:Grants")}>
          {entitlements.grants.map((grant: any) => (
            <Badge key={grant.name} variant="outline" className="font-mono text-xs font-normal">
              {grant.principalType === "Group" ? i18next.t("application:via group") + " " : ""}{grant.principal}
            </Badge>
          ))}
        </ClaimRow>
      ) : null}
      <ClaimRow label={i18next.t("application:Application Roles")}>
        {(entitlements.roles ?? []).length === 0
          ? <span className="text-xs text-muted-foreground">{i18next.t("application:None, sign-in only")}</span>
          : entitlements.roles.map((role: any) => (
            <Badge key={role.name} variant={role.isPrivileged ? "destructive" : "secondary"}>{role.displayName || role.name}</Badge>
          ))}
      </ClaimRow>
      {preview.preferredName ? (
        <ClaimRow label="preferred_username">
          <span className="font-mono text-xs">{preview.preferredName}</span>
        </ClaimRow>
      ) : null}
      <ClaimRow label="groups">
        {groups.length === 0 ? <span className="text-xs text-muted-foreground">-</span> : groups.map((value) => <ValueChip key={value} value={value} projected={projectedGroups.includes(value) && !(preview.ownGroups ?? []).includes(value)} />)}
        {preview.emitsGroups === false && projectedGroups.length > 0 ? (
          <span className="text-xs text-amber-600">{i18next.t("application:This application does not put Groups into its tokens, enable it in the token fields.")}</span>
        ) : null}
      </ClaimRow>
      <ClaimRow label="roles">
        {roles.length === 0 ? <span className="text-xs text-muted-foreground">-</span> : roles.map((value) => <ValueChip key={value} value={value} projected={projectedRoles.includes(value) && !(preview.ownRoles ?? []).includes(value)} />)}
        {preview.emitsRoles === false && projectedRoles.length > 0 ? (
          <span className="text-xs text-amber-600">{i18next.t("application:This application does not put Roles into its tokens, enable it in the token fields.")}</span>
        ) : null}
      </ClaimRow>
      {Object.keys(claims).sort().map((name) => (
        <ClaimRow key={name} label={name}>
          {(Array.isArray(claims[name]) ? claims[name] : [claims[name]]).map((value: string) => <ValueChip key={value} value={String(value)} projected />)}
        </ClaimRow>
      ))}
      {warnings.length > 0 ? (
        <div className="flex gap-2 rounded bg-amber-50 p-2 text-xs text-amber-800 dark:bg-amber-950 dark:text-amber-200">
          <AlertTriangle className="h-4 w-4 shrink-0" />
          <ul className="space-y-0.5">{warnings.map((warning) => <li key={warning}>{warning}</li>)}</ul>
        </div>
      ) : null}
      <p className="text-xs text-muted-foreground">
        {i18next.t("application:Highlighted values are added by application roles for this application only. The user's real organization, groups and roles are unchanged.")}
      </p>
    </div>
  );
}
