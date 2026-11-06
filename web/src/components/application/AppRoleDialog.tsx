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
import {Plus, Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import {MultiSelect, type MultiSelectOption} from "@/components/common/MultiSelect";
import {FormRow} from "@/components/crud/FormRow";

export interface AppRole {
  name: string;
  displayName?: string;
  description?: string;
  isPrivileged?: boolean;
  groups?: string[];
  roles?: string[];
  claims?: {name: string; value: string; type: string; category?: string}[];
}

interface AppRoleDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** the role to edit, null to add one */
  role: AppRole | null;
  organization: string;
  groupOptions: MultiSelectOption[];
  roleOptions: MultiSelectOption[];
  existingNames: string[];
  onSave: (role: AppRole) => Promise<boolean>;
}

const emptyRole = (): AppRole => ({name: "", displayName: "", description: "", isPrivileged: false, groups: [], roles: [], claims: []});

/**
 * Adds or edits an application role: the groups and roles of the application's organization whose
 * claim values it hands to users of a grant, and literal custom claims. Its name identifies it in
 * grants, so it can't change once the role exists.
 */
export function AppRoleDialog({open, onOpenChange, role, organization, groupOptions, roleOptions, existingNames, onSave}: AppRoleDialogProps) {
  const [draft, setDraft] = React.useState<AppRole>(emptyRole());
  const [saving, setSaving] = React.useState(false);
  const isNew = role === null;

  React.useEffect(() => {
    if (open) {
      setDraft(role ? JSON.parse(JSON.stringify(role)) : emptyRole());
    }
  }, [open, role]);

  const update = (patch: Partial<AppRole>) => setDraft((current) => ({...current, ...patch}));
  const claims = draft.claims ?? [];
  const nameTaken = isNew && existingNames.includes(draft.name);
  const nameValid = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/.test(draft.name);
  const isEmpty = !(draft.groups?.length || draft.roles?.length || claims.length);

  const save = async() => {
    setSaving(true);
    const cleaned: AppRole = {
      ...draft,
      claims: claims.map((claim) => ({name: claim.name.trim(), value: claim.value, type: claim.type || "String", category: "Static Value"})),
    };
    const ok = await onSave(cleaned);
    setSaving(false);
    if (ok) {
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{isNew ? i18next.t("application:Add Application Role") : i18next.t("application:Edit Application Role")}</DialogTitle>
          <DialogDescription>
            {i18next.t("application:An application role only changes what this application's tokens and userinfo say. Nobody joins a group or role of the organization through it.")}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <FormRow label={i18next.t("general:Name")} error={draft.name && (!nameValid || nameTaken) ? (nameTaken ? i18next.t("application:This name is already used") : i18next.t("application:Letters, digits, dots, underscores and hyphens only")) : undefined}>
            <Input value={draft.name} disabled={!isNew} onChange={(e) => update({name: e.target.value.trim()})} placeholder="pve-admin" />
          </FormRow>
          <FormRow label={i18next.t("general:Display name")}>
            <Input value={draft.displayName ?? ""} onChange={(e) => update({displayName: e.target.value})} />
          </FormRow>
          <FormRow label={i18next.t("general:Description")}>
            <Input value={draft.description ?? ""} onChange={(e) => update({description: e.target.value})} />
          </FormRow>
          <label className="flex items-start gap-2 text-sm">
            <Checkbox checked={!!draft.isPrivileged} onCheckedChange={(checked) => update({isPrivileged: checked === true})} className="mt-0.5" />
            <span>
              <span className="font-medium">{i18next.t("application:Privileged")}</span>
              <span className="block text-muted-foreground">
                {i18next.t("application:Mark roles that give administrative rights in the application. Assigning them asks for confirmation and they are highlighted everywhere.")}
              </span>
            </span>
          </label>

          <FormRow
            label={i18next.t("general:Groups")}
            tooltip={i18next.t("application:The users get the value these groups have in this application's groups claim, exactly like their members, in the application's token group format.")}
          >
            <MultiSelect
              value={draft.groups ?? []}
              onChange={(value) => update({groups: value})}
              options={groupOptions}
              placeholder={i18next.t("application:Groups of the organization") + " " + organization}
            />
          </FormRow>
          <FormRow
            label={i18next.t("general:Roles")}
            tooltip={i18next.t("application:The names of these roles are added to this application's roles claim.")}
          >
            <MultiSelect
              value={draft.roles ?? []}
              onChange={(value) => update({roles: value})}
              options={roleOptions}
              placeholder={i18next.t("application:Roles of the organization") + " " + organization}
            />
          </FormRow>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">{i18next.t("application:Custom Claims")}</span>
              <Button type="button" size="sm" variant="outline" onClick={() => update({claims: [...claims, {name: "", value: "", type: "String"}]})}>
                <Plus className="mr-1 h-4 w-4" />
                {i18next.t("general:Add")}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              {i18next.t("application:For applications that authorize on their own claim names. Identity and protocol claims, groups, roles and the application's own token attributes can't be set here.")}
            </p>
            {claims.map((claim, index) => (
              <div key={index} className="flex flex-wrap items-center gap-2">
                <Input
                  className="w-44"
                  value={claim.name}
                  placeholder={i18next.t("application:Claim name")}
                  onChange={(e) => update({claims: claims.map((c, i) => (i === index ? {...c, name: e.target.value} : c))})}
                />
                <Input
                  className="min-w-[10rem] flex-1"
                  value={claim.value}
                  placeholder={i18next.t("webhook:Value")}
                  onChange={(e) => update({claims: claims.map((c, i) => (i === index ? {...c, value: e.target.value} : c))})}
                />
                <Select value={claim.type || "String"} onValueChange={(value) => update({claims: claims.map((c, i) => (i === index ? {...c, type: value} : c))})}>
                  <SelectTrigger className="w-28" aria-label={i18next.t("general:Type")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="String">String</SelectItem>
                    <SelectItem value="Array">Array</SelectItem>
                  </SelectContent>
                </Select>
                <Button type="button" size="icon" variant="ghost" aria-label={i18next.t("general:Delete")} onClick={() => update({claims: claims.filter((_, i) => i !== index)})}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>
          {isEmpty && (
            <p className="text-sm text-muted-foreground">{i18next.t("application:A role without groups, roles or claims adds nothing.")}</p>
          )}
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>{i18next.t("general:Cancel")}</Button>
          <Button type="button" onClick={save} disabled={saving || !nameValid || nameTaken}>{i18next.t("general:Save")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
