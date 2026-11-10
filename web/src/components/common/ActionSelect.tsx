import * as React from "react";
import i18next from "i18next";
import {SelectField} from "@/components/common/SelectField";
import * as IntegrationBackend from "@/backend/IntegrationBackend";
import {getActionKind, parseBundle} from "@/lib/integration";

const NONE = "__none__";

/** Picks an action of one kind installed in the organization from the Marketplace, or none when allowNone is set. */
export function ActionSelect({owner, kind, value, onChange, allowNone}: {owner: string; kind: string; value?: string; onChange: (value: string) => void; allowNone?: boolean}) {
  const [actions, setActions] = React.useState<any[]>([]);
  React.useEffect(() => {
    if (!owner) {
      return;
    }
    IntegrationBackend.getIntegrations(owner).then((res: any) => {
      if (res.status === "ok") {
        setActions((res.data ?? []).filter((integration: any) => integration.type === "action" &&
          (integration.kind || getActionKind(parseBundle(integration.bundle))) === kind));
      }
    });
  }, [owner, kind]);

  return (
    <SelectField
      value={allowNone && !value ? NONE : value ?? ""}
      onChange={(v) => onChange(v === NONE ? "" : v)}
      options={[
        ...(allowNone ? [{id: NONE, name: i18next.t("general:None")}] : []),
        ...actions.map((integration: any) => ({id: integration.name, name: `${integration.displayName} (${integration.name})`})),
      ]}
    />
  );
}
