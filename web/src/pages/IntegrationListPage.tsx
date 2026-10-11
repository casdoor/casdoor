import i18next from "i18next";
import {Link} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationFilter} from "@/hooks/use-organization";
import * as IntegrationBackend from "@/backend/IntegrationBackend";
import {IntegrationLogo, getIntegrationTypeLabel} from "@/lib/integration";
import * as Setting from "@/lib/setting";

export default function IntegrationListPage() {
  const {account} = useAccount();
  const organizationName = useOrganizationFilter();
  const readOnly = !Setting.isLocalAdminUser(account);

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/integrations/${r.owner}/${r.name}`}),
    organizationColumn(),
    {
      dataIndex: "displayName",
      title: i18next.t("general:Display name"),
      width: 200,
      sortable: true,
      searchable: true,
      render: (value, record) => (
        <span className="flex items-center gap-2">
          <IntegrationLogo logo={record.logo} name={value} className="size-5" />
          {value}
        </span>
      ),
    },
    {
      dataIndex: "type",
      title: i18next.t("general:Type"),
      width: 110,
      sortable: true,
      render: (value) => getIntegrationTypeLabel(value),
    },
    textColumn({dataIndex: "version", title: i18next.t("system:Version"), width: 100, mono: true}),
    textColumn({
      dataIndex: "application",
      title: i18next.t("general:Application"),
      width: 160,
      searchable: true,
      link: (value, record) => (value ? `/applications/${record.owner}/${value}` : undefined),
    }),
    textColumn({
      dataIndex: "provider",
      title: i18next.t("general:Provider"),
      width: 160,
      searchable: true,
      link: (value, record) => (value ? `/providers/${record.owner}/${value}` : undefined),
    }),
    dateColumn("createdTime"),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Integrations")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        IntegrationBackend.getIntegrations(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      toolbar={
        <Button asChild>
          <Link to="/marketplace">{i18next.t("general:Marketplace")}</Link>
        </Button>
      }
      readOnly={readOnly}
      editUrl={(r) => `/integrations/${r.owner}/${r.name}`}
      remove={(r) => IntegrationBackend.deleteIntegration(r)}
    />
  );
}
