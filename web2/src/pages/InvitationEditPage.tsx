import * as React from "react";
import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useApplicationOptions, useGroupOptions, useOrganizationOptions} from "@/hooks/use-options";
import {InvitationSend} from "@/components/user/InvitationSend";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as InvitationBackend from "@/backend/InvitationBackend";
import * as Setting from "@/lib/setting";

const STATES = ["Active", "Suspended"];

export default function InvitationEditPage() {
  const {organizationName = "", invitationName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const applications = useApplicationOptions(organizationName);
  const groups = useGroupOptions(organizationName);
  // copySignupLink needs each organization's defaultApplication, which the option
  // list does not carry, so the raw objects are fetched alongside it
  const [organizationObjs, setOrganizationObjs] = React.useState<any[]>([]);

  React.useEffect(() => {
    if (!Setting.isAdminUser(account)) {
      return;
    }
    OrganizationBackend.getOrganizations("admin").then((res: any) => {
      if (res.status === "ok") {
        setOrganizationObjs(res.data ?? []);
      }
    });
  }, [account]);

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
    {type: "text", name: "code", labelKey: "invitation:Code"},
    {type: "text", name: "defaultCode", labelKey: "invitation:Default code"},
    {type: "number", name: "quota", labelKey: "invitation:Quota"},
    {type: "number", name: "usedCount", labelKey: "invitation:Used count", disabled: () => true},
    {
      type: "select",
      name: "application",
      labelKey: "general:Application",
      options: () => [{value: "All", label: i18next.t("general:All")}, ...applications],
    },
    {type: "select", name: "signupGroup", labelKey: "provider:Signup group", options: () => groups},
    {type: "text", name: "username", labelKey: "signup:Username"},
    {type: "email", name: "email", labelKey: "general:Email"},
    {type: "text", name: "phone", labelKey: "general:Phone"},
    {
      type: "custom",
      name: "send",
      labelKey: "general:Send",
      block: true,
      render: (ctx) => (
        <InvitationSend invitation={ctx.record} organizations={organizationObjs} isAdd={ctx.mode === "add"} />
      ),
    },
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => STATES.map((item) => ({value: item, label: item})),
    },
  ];

  return (
    <SimpleEditPage
      titleKey="invitation:Edit Invitation"
      backTo="/invitations"
      deps={[organizationName, invitationName]}
      fields={fields}
      fetch={() => InvitationBackend.getInvitation(organizationName, invitationName)}
      add={(record) => InvitationBackend.addInvitation(record)}
      update={(record) => InvitationBackend.updateInvitation(organizationName, invitationName, record)}
      editUrl={(record) => `/invitations/${record.owner}/${record.name}`}
    />
  );
}
