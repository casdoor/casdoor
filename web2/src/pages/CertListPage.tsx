import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
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
    textColumn({dataIndex: "scope", title: i18next.t("provider:Scope"), width: 110}),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 110}),
    textColumn({dataIndex: "cryptoAlgorithm", title: i18next.t("cert:Crypto algorithm"), width: 150}),
    textColumn({dataIndex: "bitSize", title: i18next.t("cert:Bit size"), width: 110}),
    textColumn({dataIndex: "expireInYears", title: i18next.t("cert:Expire in years"), width: 140}),
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
    />
  );
}
