import i18next from "i18next";
import {Link} from "react-router-dom";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as TokenBackend from "@/backend/TokenBackend";
import * as Setting from "@/lib/setting";
import {newToken} from "@/pages/defaults";

export default function TokenListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/tokens/${r.name}`, width: 170}),
    dateColumn(),
    {
      dataIndex: "application",
      title: i18next.t("general:Application"),
      width: 150,
      sortable: true,
      searchable: true,
      render: (value, record) => (
        <Link to={`/applications/${record.organization}/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    {
      dataIndex: "organization",
      title: i18next.t("general:Organization"),
      width: 140,
      sortable: true,
      searchable: true,
      render: (value) => (
        <Link to={`/organizations/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    {
      dataIndex: "user",
      title: i18next.t("general:User"),
      width: 130,
      sortable: true,
      searchable: true,
      render: (value, record) => (
        <Link to={`/users/${record.organization}/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    {
      dataIndex: "code",
      sortable: true,
      searchable: true,
      title: i18next.t("token:Authorization code"),
      width: 160,
      render: (value) => (value ? Setting.getClickable(Setting.getShortText(value)) : null),
    },
    {
      dataIndex: "accessToken",
      sortable: true,
      searchable: true,
      title: i18next.t("token:Access token"),
      width: 160,
      render: (value) => (value ? Setting.getClickable(Setting.getShortText(value)) : null),
    },
    textColumn({dataIndex: "expiresIn", title: i18next.t("token:Expires in"), width: 120, searchable: true}),
    textColumn({dataIndex: "scope", title: i18next.t("provider:Scope"), width: 110, searchable: true}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Tokens")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        TokenBackend.getTokens(
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
      newRecord={account ? () => newToken(account) : undefined}
      editUrl={(r) => `/tokens/${r.name}`}
      remove={(r) => TokenBackend.deleteToken(r)}
    />
  );
}
