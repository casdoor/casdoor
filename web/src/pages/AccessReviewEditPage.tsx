import * as React from "react";
import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {EditionNotice} from "@/components/common/EditionNotice";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {AccessReviewItemTable} from "@/components/access-review/AccessReviewItemTable";
import {AccessReviewStateBadge} from "@/pages/AccessReviewListPage";
import {useAccount} from "@/hooks/use-account";
import {useGroupNameOptions, useOrganizationOptions, useRoleNameOptions, useUserNameOptions} from "@/hooks/use-options";
import * as AccessReviewBackend from "@/backend/AccessReviewBackend";
import * as PermissionBackend from "@/backend/PermissionBackend";
import * as Setting from "@/lib/setting";

const ENTITLEMENT_TYPES = ["Role", "Group", "Permission", "Application"];

function usePermissionNameOptions(organizationName: string) {
  const [options, setOptions] = React.useState<{value: string; label: string}[]>([]);
  React.useEffect(() => {
    if (!organizationName) {
      return;
    }
    PermissionBackend.getPermissions(organizationName, 1, 1000).then((res: any) => {
      if (res.status === "ok") {
        setOptions((res.data ?? []).map((item: any) => ({value: item.name, label: item.displayName ? `${item.displayName} (${item.name})` : item.name})));
      }
    });
  }, [organizationName]);
  return options;
}

function getEntitlementTypeLabel(type: string) {
  switch (type) {
  case "Role":
    return i18next.t("enterprise:Role members");
  case "Group":
    return i18next.t("enterprise:Group members");
  case "Permission":
    return i18next.t("enterprise:Users listed in permissions");
  default:
    return i18next.t("enterprise:Application access");
  }
}

/** One round of a campaign: its progress, the actions on it and its items. */
function RoundPanel({review, onChanged}: {review: any; onChanged: () => void}) {
  const {account} = useAccount();
  const [items, setItems] = React.useState<any[]>([]);
  const [version, setVersion] = React.useState(0);

  React.useEffect(() => {
    if (review.round <= 0) {
      return;
    }
    AccessReviewBackend.getAccessReviewItems(review.owner, review.name, review.round).then((res: any) => {
      if (res.status === "ok") {
        setItems(res.data ?? []);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  }, [review.owner, review.name, review.round, review.state, version]);

  if (review.round <= 0) {
    return null;
  }

  const progress = review.progress ?? {};
  const isActive = review.state === "Active";

  return (
    <Card className="mt-6">
      <CardHeader className="flex flex-row flex-wrap items-center gap-3 space-y-0">
        <CardTitle className="text-base">
          {i18next.t("enterprise:Round")} {review.round}
        </CardTitle>
        <AccessReviewStateBadge state={review.state} />
        <span className="text-sm text-muted-foreground">
          {Setting.getFormattedDate(review.startTime)} - {Setting.getFormattedDate(review.endTime)}
        </span>
        <span className="text-sm">
          {i18next.t("enterprise:Undecided")}: {progress.pending ?? 0}, {i18next.t("enterprise:Keep")}: {progress.keep ?? 0}, {i18next.t("enterprise:Revoke")}: {progress.revoke ?? 0}, {i18next.t("enterprise:Applied")}: {progress.applied ?? 0}
        </span>
      </CardHeader>
      <CardContent>
        <AccessReviewItemTable
          items={items}
          canDecide={isActive}
          showApplied
          selfName={account?.owner === review.owner ? account?.name : undefined}
          onDecided={() => {
            setVersion((v) => v + 1);
            onChanged();
          }}
        />
      </CardContent>
    </Card>
  );
}

/** Settings of an access review campaign, and its current round once started. */
export default function AccessReviewEditPage() {
  const {organizationName = "", reviewName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const users = useUserNameOptions(organizationName);
  const roles = useRoleNameOptions(organizationName);
  const groups = useGroupNameOptions(organizationName);
  const permissions = usePermissionNameOptions(organizationName);

  const isRunning = (ctx: any) => ctx.record.state === "Active";

  const fields: EditField[] = [
    {type: "select", name: "owner", labelKey: "general:Organization", options: () => organizations, disabled: () => !Setting.isAdminUser(account)},
    {type: "text", name: "name", labelKey: "general:Name", required: true, disabled: (ctx) => ctx.record.round > 0},
    {type: "text", name: "displayName", labelKey: "general:Display name", disabled: isRunning},
    {type: "textarea", name: "description", labelKey: "general:Description", disabled: isRunning},
    {
      type: "multiselect",
      name: "entitlementTypes",
      labelKey: "enterprise:What to review",
      required: true,
      disabled: isRunning,
      options: () => ENTITLEMENT_TYPES.map((type) => ({value: type, label: getEntitlementTypeLabel(type)})),
    },
    {type: "multiselect", name: "roles", labelKey: "enterprise:Only these roles", disabled: isRunning, options: () => roles},
    {type: "multiselect", name: "groups", labelKey: "enterprise:Only these groups", disabled: isRunning, options: () => groups},
    {type: "multiselect", name: "permissions", labelKey: "enterprise:Only these permissions", disabled: isRunning, options: () => permissions},
    {
      type: "select",
      name: "reviewerType",
      labelKey: "enterprise:Who reviews",
      disabled: isRunning,
      options: () => [
        {value: "Manager", label: i18next.t("enterprise:The manager of the user's group, otherwise the reviewers below")},
        {value: "Reviewers", label: i18next.t("enterprise:The reviewers below, in turn")},
      ],
    },
    {type: "multiselect", name: "reviewers", labelKey: "enterprise:Reviewers", disabled: isRunning, options: () => users},
    {type: "number", name: "durationDays", labelKey: "enterprise:Review period (days)", min: 1, max: 365, disabled: isRunning},
    {
      type: "select",
      name: "defaultDecision",
      labelKey: "enterprise:Undecided at the deadline",
      disabled: isRunning,
      options: () => [
        {value: "Keep", label: i18next.t("enterprise:Keep")},
        {value: "Revoke", label: i18next.t("enterprise:Revoke")},
      ],
    },
    {type: "switch", name: "autoApply", labelKey: "enterprise:Apply revocations when the round ends", disabled: isRunning},
    {type: "number", name: "recurrenceDays", labelKey: "enterprise:Repeat every (days)", min: 0, disabled: isRunning},
  ];

  const run = (promise: Promise<any>, ctx: any, success: (data: any) => string) =>
    promise
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", success(res.data));
          ctx.reload();
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error) => Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`));

  const applyMessage = (data: any) =>
    i18next.t("enterprise:Applied %s revocations, %s failed").replace("%s", `${data?.applied ?? 0}`).replace("%s", `${data?.failed ?? 0}`);

  return (
    <SimpleEditPage
      titleKey="enterprise:Edit Access Review"
      backTo="/access-reviews"
      deps={[organizationName, reviewName]}
      fields={fields}
      fetch={() => AccessReviewBackend.getAccessReview(organizationName, reviewName)}
      add={(record) => AccessReviewBackend.addAccessReview(record)}
      update={(record) => AccessReviewBackend.updateAccessReview(organizationName, reviewName, record)}
      editUrl={(record) => `/access-reviews/${record.owner}/${record.name}`}
      extraActions={(ctx) =>
        ctx.mode === "add" ? null : (
          <>
            {ctx.record.state !== "Active" ? (
              <ConfirmButton
                destructive={false}
                title={i18next.t("enterprise:Start a review round now?")}
                description={i18next.t("enterprise:Save your changes first, the round uses the saved settings")}
                onConfirm={() =>
                  run(AccessReviewBackend.startAccessReview(ctx.record.owner, ctx.record.name), ctx, (count) =>
                    i18next.t("enterprise:Started with %s items to review").replace("%s", `${count}`))
                }
              >
                {i18next.t("enterprise:Start round")}
              </ConfirmButton>
            ) : (
              <>
                <Button
                  variant="outline"
                  onClick={() =>
                    run(AccessReviewBackend.remindAccessReview(ctx.record.owner, ctx.record.name), ctx, (data) =>
                      i18next.t("enterprise:Reminded %s reviewers").replace("%s", `${data?.sent ?? 0}`) +
                      (data?.failed?.length ? `, ${i18next.t("enterprise:not reached")}: ${data.failed.join(", ")}` : ""))
                  }
                >
                  {i18next.t("enterprise:Remind reviewers")}
                </Button>
                <ConfirmButton
                  variant="outline"
                  title={i18next.t("enterprise:End the round now?")}
                  description={i18next.t("enterprise:The undecided items get the decision for the deadline")}
                  onConfirm={() => run(AccessReviewBackend.closeAccessReview(ctx.record.owner, ctx.record.name), ctx, applyMessage)}
                >
                  {i18next.t("enterprise:End round")}
                </ConfirmButton>
              </>
            )}
            {ctx.record.round > 0 ? (
              <>
                <ConfirmButton
                  variant="outline"
                  title={i18next.t("enterprise:Apply the revocations now?")}
                  description={i18next.t("enterprise:The users are removed from the revoked roles, groups and permissions, each change is written to the records")}
                  onConfirm={() => run(AccessReviewBackend.applyAccessReview(ctx.record.owner, ctx.record.name), ctx, applyMessage)}
                >
                  {i18next.t("enterprise:Apply revocations")}
                </ConfirmButton>
                <Button variant="outline" asChild>
                  <a href={AccessReviewBackend.getExportUrl(ctx.record.owner, ctx.record.name)} download>
                    {i18next.t("enterprise:Export CSV")}
                  </a>
                </Button>
              </>
            ) : null}
          </>
        )
      }
    >
      {(ctx) => (
        <>
          <EditionNotice edition="Enterprise" />
          {ctx.mode === "add" ? null : <RoundPanel review={ctx.record} onChanged={ctx.reload} />}
        </>
      )}
    </SimpleEditPage>
  );
}
