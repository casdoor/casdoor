import i18next from "i18next";
import dayjs from "dayjs";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, linkColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as Setting from "@/lib/setting";
import {newOrganization} from "@/pages/organization-defaults";

export default function OrganizationListPage() {
  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (record) => `/organizations/${record.name}`, width: 140}),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 160}),
    {
      dataIndex: "favicon",
      title: i18next.t("general:Favicon"),
      width: 70,
      align: "center",
      render: (value) =>
        value ? (
          <a href={value} target="_blank" rel="noreferrer">
            <img src={value} alt="favicon" className="mx-auto h-8 w-8 object-contain" />
          </a>
        ) : null,
    },
    {
      dataIndex: "websiteUrl",
      title: i18next.t("organization:Website URL"),
      width: 200,
      sortable: true,
      searchable: true,
      render: (value) =>
        value ? (
          <a href={value} target="_blank" rel="noreferrer" className="underline-offset-4 hover:underline">
            {value}
          </a>
        ) : null,
    },
    textColumn({dataIndex: "passwordType", title: i18next.t("general:Password type"), width: 140}),
    textColumn({dataIndex: "passwordSalt", title: i18next.t("general:Password salt"), width: 130}),
    {
      dataIndex: "defaultAvatar",
      title: i18next.t("general:Default avatar"),
      width: 120,
      align: "center",
      render: (value) =>
        value ? (
          <a href={value} target="_blank" rel="noreferrer">
            <img src={value} alt="avatar" className="mx-auto h-8 w-8 rounded-full object-cover" />
          </a>
        ) : null,
    },
    {
      dataIndex: "orgBalance",
      title: i18next.t("organization:Org balance"),
      width: 130,
      sortable: true,
      render: (value, record) => (
        <Badge variant="secondary">{Setting.getPriceDisplay(value ?? 0, record.balanceCurrency)}</Badge>
      ),
    },
    {
      dataIndex: "userBalance",
      title: i18next.t("organization:User balance"),
      width: 130,
      sortable: true,
      render: (value, record) => Setting.getPriceDisplay(value ?? 0, record.balanceCurrency),
    },
    {
      dataIndex: "balanceCredit",
      title: i18next.t("organization:Balance credit"),
      width: 130,
      sortable: true,
      render: (value, record) => Setting.getPriceDisplay(value ?? 0, record.balanceCurrency),
    },
    textColumn({dataIndex: "balanceCurrency", title: i18next.t("organization:Balance currency"), width: 130}),
    boolColumn({dataIndex: "enableSoftDeletion", title: i18next.t("organization:Soft deletion"), width: 120}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Organizations")}
      columns={columns}
      formType="organizations"
      rowKey={(row) => row.name}
      fetch={(query) =>
        OrganizationBackend.getOrganizations(
          "admin",
          "",
          query.page,
          query.pageSize,
          query.searchedColumn,
          query.searchText,
          query.sortField,
          query.sortOrder,
        )
      }
      newRecord={() => newOrganization(dayjs().format())}
      editUrl={(record) => `/organizations/${record.name}`}
      remove={(record) =>
        OrganizationBackend.deleteOrganization(record).then((res: any) => {
          if (res.status === "ok") {
            window.dispatchEvent(new Event("storageOrganizationsChanged"));
          }
          return res;
        })
      }
    />
  );
}
