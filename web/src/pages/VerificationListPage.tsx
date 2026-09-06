import i18next from "i18next";
import {Link} from "react-router-dom";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, clientIpColumn, dateColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as VerificationBackend from "@/backend/VerificationBackend";
import * as Setting from "@/lib/setting";

export default function VerificationListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  // GetVerifications() filters by "owner", an empty one means every organization
  const owner = account && Setting.isDefaultOrganizationSelected(account) ? "" : organizationName;

  const columns: ColumnDef<any>[] = [
    organizationColumn(140, "owner", undefined, "left"),
    textColumn({dataIndex: "name", title: i18next.t("general:Name"), width: 180, searchable: true, fixed: "left"}),
    dateColumn(),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 110, searchable: true}),
    {
      dataIndex: "user",
      title: i18next.t("general:User"),
      width: 140,
      sortable: true,
      searchable: true,
      render: (value, record) => (
        <Link to={`/users/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    textColumn({dataIndex: "provider", title: i18next.t("general:Provider"), width: 160, searchable: true, link: (v, r: any) => `/providers/${r.owner}/${v}`}),
    clientIpColumn({
      dataIndex: "remoteAddr",
      // the backend stores it as "1.2.3.4: " when the port is unknown
      normalize: (value) => (value.endsWith(": ") ? value.slice(0, -2) : value),
    }),
    textColumn({dataIndex: "receiver", title: i18next.t("verification:Receiver"), width: 180, searchable: true}),
    textColumn({dataIndex: "code", title: i18next.t("login:Verification code"), width: 130, mono: true, searchable: true}),
    boolColumn({dataIndex: "isUsed", title: i18next.t("verification:Is used")}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Verifications")}
      columns={columns}
      deps={[owner]}
      showActionColumn={false}
      fetch={(q) =>
        VerificationBackend.getVerifications(
          owner,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
    />
  );
}
