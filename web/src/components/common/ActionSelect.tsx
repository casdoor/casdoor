import * as React from "react";
import {SelectField} from "@/components/common/SelectField";
import * as IntegrationBackend from "@/backend/IntegrationBackend";
import {getActionKind, parseBundle} from "@/lib/integration";

/** Picks an action of one kind installed in the organization from the Marketplace. */
export function ActionSelect({owner, kind, value, onChange}: {owner: string; kind: string; value?: string; onChange: (value: string) => void}) {
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
      value={value ?? ""}
      onChange={onChange}
      options={actions.map((integration: any) => ({id: integration.name, name: `${integration.displayName} (${integration.name})`}))}
    />
  );
}
