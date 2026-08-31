import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as CertBackend from "@/backend/CertBackend";
import * as Setting from "@/lib/setting";
import {newCert} from "@/pages/defaults";

export default function CertListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  const isGlobal = account ? Setting.isDefaultOrganizationSelected(account) : false;

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/certs/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "scope", title: i18next.t("provider:Scope"), width: 110, filters: valueFilters(["JWT"])}),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 110, filters: valueFilters(["x509", "Payment"])}),
    textColumn({dataIndex: "cryptoAlgorithm", title: i18next.t("cert:Crypto algorithm"), width: 150, filters: valueFilters(["RS256"])}),
    textColumn({dataIndex: "bitSize", title: i18next.t("cert:Bit size"), width: 110, searchable: true}),
    textColumn({dataIndex: "expireInYears", title: i18next.t("cert:Expire in years"), width: 140, searchable: true}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Certs")}
      columns={columns}
      deps={[organizationName, isGlobal]}
      fetch={(q) =>
        isGlobal
          ? CertBackend.getGlobalCerts(q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
          : CertBackend.getCerts(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newCert(account, organizationName) : undefined}
      editUrl={(r) => `/certs/${r.owner}/${r.name}`}
      remove={(r) => CertBackend.deleteCert(r)}
      actionColumnWidth={260}
      rowActions={(record, _index, {refresh}) => (
        <Button
          variant="outline"
          size="sm"
          onClick={() => {
            CertBackend.refreshDomainExpire(record.owner, record.name)
              .then((res: any) => {
                if (res.status === "error") {
                  Setting.showMessage("error", `Failed to refresh domain expire: ${res.msg}`);
                } else {
                  Setting.showMessage("success", "Domain expire refreshed successfully");
                  refresh();
                }
              })
              .catch((error) => Setting.showMessage("error", `Domain expire failed to refresh: ${error}`));
          }}
        >
          {i18next.t("general:Refresh")}
        </Button>
      )}
    />
  );
}
