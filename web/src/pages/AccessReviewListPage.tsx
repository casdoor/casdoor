import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as AccessReviewBackend from "@/backend/AccessReviewBackend";
import * as Setting from "@/lib/setting";

export function getAccessReviewStateText(state: string) {
  switch (state) {
  case "Active":
    return i18next.t("enterprise:In review");
  case "Completed":
    return i18next.t("enterprise:Completed");
  default:
    return i18next.t("enterprise:Draft");
  }
}

export function AccessReviewStateBadge({state}: {state: string}) {
  const variant = state === "Active" ? "info" : state === "Completed" ? "success" : "secondary";
  return <Badge variant={variant}>{getAccessReviewStateText(state)}</Badge>;
}

export function newAccessReview(organizationName: string) {
  const randomName = Setting.getRandomName();
  return {
    owner: organizationName,
    name: `access_review_${randomName}`,
    createdTime: new Date().toISOString(),
    displayName: `New Access Review - ${randomName}`,
    description: "",
    entitlementTypes: ["Role", "Group", "Permission", "Application"],
    roles: [],
    groups: [],
    permissions: [],
    reviewerType: "Manager",
    reviewers: [],
    durationDays: 14,
    defaultDecision: "Keep",
    autoApply: false,
    recurrenceDays: 90,
    state: "Draft",
    round: 0,
  };
}

/** The access review campaigns (定期权限复核) of the organization, an enterprise edition feature. */
export default function AccessReviewListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/access-reviews/${r.owner}/${r.name}`, width: 180}),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 200}),
    dateColumn(),
    {
      dataIndex: "state",
      title: i18next.t("general:State"),
      width: 110,
      render: (value: string) => <AccessReviewStateBadge state={value} />,
    },
    textColumn({dataIndex: "round", title: i18next.t("enterprise:Round"), width: 80}),
    {
      dataIndex: "progress",
      title: i18next.t("enterprise:Progress"),
      width: 150,
      render: (value: any) =>
        value && value.total > 0 ? `${value.total - value.pending} / ${value.total}` : "-",
    },
    {
      dataIndex: "endTime",
      title: i18next.t("enterprise:Deadline"),
      width: 165,
      render: (value: string, record: any) => (record.state === "Active" && value ? Setting.getFormattedDate(value) : "-"),
    },
    {
      dataIndex: "recurrenceDays",
      title: i18next.t("enterprise:Repeat every (days)"),
      width: 130,
      render: (value: number) => (value > 0 ? value : "-"),
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("edition:Access reviews")}
      description={i18next.t("enterprise:Access reviews - Tooltip")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        AccessReviewBackend.getAccessReviews(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newAccessReview(organizationName) : undefined}
      editUrl={(r) => `/access-reviews/${r.owner}/${r.name}`}
      remove={(r) => AccessReviewBackend.deleteAccessReview(r)}
      deleteDisabled={(r) => r.state === "Active"}
    />
  );
}
