import i18next from "i18next";
import dayjs from "dayjs";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {EditionNotice} from "@/components/common/EditionNotice";
import {enumColumn} from "@/lib/enum-labels";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationFilter} from "@/hooks/use-organization";
import * as PrivilegedAccountBackend from "@/backend/PrivilegedAccountBackend";
import {PAM_ACCOUNT_TYPES, PAM_ROTATION_STATES} from "@/lib/pam";
import * as Setting from "@/lib/setting";

function newPrivilegedAccount(account: any) {
  const randomName = Setting.getRandomName();
  return {
    owner: Setting.getRequestOrganization(account),
    name: `privileged_account_${randomName}`,
    createdTime: dayjs().format(),
    updatedTime: dayjs().format(),
    displayName: `New Privileged Account - ${randomName}`,
    description: "",
    type: "Linux",
    host: "",
    port: 22,
    database: "",
    username: "root",
    secret: "",
    adminUsername: "",
    adminSecret: "",
    hostKey: "",
    skipTlsVerify: false,
    role: "",
    users: [],
    approvers: [],
    requireApproval: true,
    maxCheckoutMinutes: 60,
    rotateAfterCheckin: false,
    rotateIntervalDays: 0,
    passwordLength: 24,
    passwordCharset: "Complex",
    isEnabled: true,
  };
}

export default function PrivilegedAccountListPage() {
  const {account} = useAccount();
  const organizationName = useOrganizationFilter();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/privileged-accounts/${r.owner}/${r.name}`, width: 170}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 180}),
    enumColumn({dataIndex: "type", title: i18next.t("general:Type"), map: PAM_ACCOUNT_TYPES, width: 150}),
    textColumn({dataIndex: "host", title: i18next.t("general:Host"), searchable: true, width: 150}),
    textColumn({dataIndex: "username", title: i18next.t("signup:Username"), searchable: true, width: 140}),
    boolColumn({dataIndex: "requireApproval", title: i18next.t("pam:Require approval")}),
    enumColumn({dataIndex: "rotationState", title: i18next.t("pam:Last rotation"), map: PAM_ROTATION_STATES, width: 120}),
    dateColumn("lastRotatedTime", i18next.t("pam:Last rotated time")),
    boolColumn({dataIndex: "isEnabled", title: i18next.t("general:Is enabled")}),
  ];

  return (
    <div className="space-y-2">
      <EditionNotice edition="Enterprise" message={i18next.t("pam:Privileged account management needs the enterprise edition")} />
      <CrudListPage
        title={i18next.t("pam:Privileged accounts")}
        description={i18next.t("pam:Privileged accounts - Tooltip")}
        columns={columns}
        deps={[organizationName]}
        fetch={(q) =>
          PrivilegedAccountBackend.getPrivilegedAccounts(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
        }
        newRecord={account ? () => newPrivilegedAccount(account) : undefined}
        editUrl={(r) => `/privileged-accounts/${r.owner}/${r.name}`}
        remove={(r) => PrivilegedAccountBackend.deletePrivilegedAccount(r)}
      />
    </div>
  );
}
