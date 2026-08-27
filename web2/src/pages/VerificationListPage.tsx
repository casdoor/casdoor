import i18next from "i18next";
import {Link} from "react-router-dom";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as VerificationBackend from "@/backend/VerificationBackend";

export default function VerificationListPage() {
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    organizationColumn(),
    textColumn({dataIndex: "name", title: i18next.t("general:Name"), width: 180, searchable: true}),
    dateColumn(),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 110}),
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
    textColumn({dataIndex: "provider", title: i18next.t("general:Provider"), width: 160}),
    textColumn({dataIndex: "remoteAddr", title: i18next.t("general:Client IP"), width: 140}),
    textColumn({dataIndex: "receiver", title: i18next.t("verification:Receiver"), width: 180, searchable: true}),
    textColumn({dataIndex: "code", title: i18next.t("login:Verification code"), width: 130, mono: true}),
    boolColumn({dataIndex: "isUsed", title: i18next.t("verification:Is used")}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Verifications")}
      columns={columns}
      deps={[organizationName]}
      showActionColumn={false}
      fetch={(q) =>
        VerificationBackend.getVerifications(
          "admin",
          organizationName,
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
