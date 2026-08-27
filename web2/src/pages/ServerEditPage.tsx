import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useApplicationOptions, useOrganizationOptions} from "@/hooks/use-options";
import * as ServerBackend from "@/backend/ServerBackend";
import * as Setting from "@/lib/setting";

export default function ServerEditPage() {
  const {organizationName = "", serverName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const applications = useApplicationOptions(organizationName);

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name"},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "text", name: "description", labelKey: "general:Description"},
    {type: "text", name: "url", labelKey: "general:URL"},
    {type: "select", name: "application", labelKey: "general:Application", options: () => applications},
  ];

  return (
    <SimpleEditPage
      titleKey="server:Edit Server"
      backTo="/servers"
      deps={[organizationName, serverName]}
      fields={fields}
      fetch={() => ServerBackend.getServer(organizationName, serverName)}
      add={(record) => ServerBackend.addServer(record)}
      update={(record) => ServerBackend.updateServer(organizationName, serverName, record)}
      editUrl={(record) => `/servers/${record.owner}/${record.name}`}
      extraActions={(ctx) => (
        <Button
          variant="outline"
          onClick={() => {
            ServerBackend.syncMcpTool(ctx.record.owner, ctx.record.name, ctx.record).then((res: any) => {
              if (res.status === "ok") {
                Setting.showMessage("success", i18next.t("general:Successfully saved"));
              } else {
                Setting.showMessage("error", res.msg);
              }
            });
          }}
        >
          {i18next.t("server:Sync tools")}
        </Button>
      )}
    />
  );
}
