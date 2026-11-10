import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Input} from "@/components/ui/input";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import * as AccessReviewBackend from "@/backend/AccessReviewBackend";
import * as Setting from "@/lib/setting";

export function getEntitlementTypeText(type: string) {
  switch (type) {
  case "Role":
    return i18next.t("general:Roles");
  case "Group":
    return i18next.t("general:Groups");
  case "Permission":
    return i18next.t("general:Permissions");
  default:
    return i18next.t("enterprise:Application access");
  }
}

export function DecisionBadge({decision}: {decision: string}) {
  if (decision === "Keep") {
    return <Badge variant="success">{i18next.t("enterprise:Keep")}</Badge>;
  }
  if (decision === "Revoke") {
    return <Badge variant="destructive">{i18next.t("enterprise:Revoke")}</Badge>;
  }
  return <Badge variant="warning">{i18next.t("enterprise:Undecided")}</Badge>;
}

/**
 * The items of an access review. A reviewer, or an admin of a running campaign, keeps or revokes them one by one or
 * the selected ones at once, with an optional comment.
 */
export function AccessReviewItemTable({
  items,
  canDecide,
  showReview = false,
  showApplied = false,
  selfName,
  onDecided,
}: {
  items: any[];
  canDecide: boolean;
  showReview?: boolean;
  showApplied?: boolean;
  /** nobody decides on their own access */
  selfName?: string;
  onDecided: () => void;
}) {
  const [selected, setSelected] = React.useState<string[]>([]);
  const [comment, setComment] = React.useState("");
  const [busy, setBusy] = React.useState(false);

  const decidable = (item: any) => canDecide && !item.applied && item.user !== selfName;
  const decidableNames = items.filter(decidable).map((item) => item.name);
  const allSelected = decidableNames.length > 0 && decidableNames.every((name) => selected.includes(name));

  React.useEffect(() => {
    setSelected((current) => current.filter((name) => items.some((item) => item.name === name)));
  }, [items]);

  const decide = (names: string[], decision: string) => {
    if (names.length === 0) {
      return;
    }
    setBusy(true);
    AccessReviewBackend.decideAccessReviewItems(names, decision, comment)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          setSelected([]);
          onDecided();
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error) => Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`))
      .finally(() => setBusy(false));
  };

  if (items.length === 0) {
    return <p className="py-6 text-center text-sm text-muted-foreground">{i18next.t("enterprise:Nothing to review")}</p>;
  }

  return (
    <div className="space-y-3">
      {canDecide ? (
        <div className="flex flex-wrap items-center gap-2">
          <Input
            className="w-72"
            value={comment}
            placeholder={i18next.t("enterprise:Comment (optional)")}
            onChange={(e) => setComment(e.target.value)}
          />
          <Button size="sm" variant="outline" disabled={busy || selected.length === 0} onClick={() => decide(selected, "Keep")}>
            {i18next.t("enterprise:Keep selected")} ({selected.length})
          </Button>
          <Button size="sm" variant="outline" disabled={busy || selected.length === 0} onClick={() => decide(selected, "Revoke")}>
            {i18next.t("enterprise:Revoke selected")} ({selected.length})
          </Button>
        </div>
      ) : null}
      <div className="overflow-x-auto rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {canDecide ? (
                <TableHead className="w-8">
                  <Checkbox
                    checked={allSelected}
                    disabled={decidableNames.length === 0}
                    onCheckedChange={(v) => setSelected(v ? decidableNames : [])}
                    aria-label={i18next.t("general:All")}
                  />
                </TableHead>
              ) : null}
              {showReview ? <TableHead>{i18next.t("edition:Access reviews")}</TableHead> : null}
              <TableHead>{i18next.t("general:User")}</TableHead>
              <TableHead>{i18next.t("enterprise:Access")}</TableHead>
              <TableHead>{i18next.t("enterprise:Reviewer")}</TableHead>
              <TableHead>{i18next.t("entry:Decision")}</TableHead>
              <TableHead>{i18next.t("enterprise:Comment")}</TableHead>
              {showApplied ? <TableHead>{i18next.t("enterprise:Applied")}</TableHead> : null}
              {canDecide ? <TableHead className="w-40" /> : null}
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.name}>
                {canDecide ? (
                  <TableCell>
                    <Checkbox
                      checked={selected.includes(item.name)}
                      disabled={!decidable(item)}
                      onCheckedChange={(v) =>
                        setSelected((current) => (v ? [...current, item.name] : current.filter((name) => name !== item.name)))
                      }
                    />
                  </TableCell>
                ) : null}
                {showReview ? (
                  <TableCell className="whitespace-nowrap">
                    {item.reviewDisplayName}
                    {item.endTime ? (
                      <div className="text-xs text-muted-foreground">
                        {i18next.t("enterprise:Due")}: {Setting.getFormattedDate(item.endTime)}
                      </div>
                    ) : null}
                  </TableCell>
                ) : null}
                <TableCell className="whitespace-nowrap">
                  {item.userDisplayName || item.user}
                  <div className="text-xs text-muted-foreground">{item.user}</div>
                </TableCell>
                <TableCell>
                  <span className="text-xs text-muted-foreground">{getEntitlementTypeText(item.entitlementType)}</span>
                  <div>{item.entitlementDisplayName || item.entitlement}</div>
                </TableCell>
                <TableCell>{item.reviewer || <span className="text-muted-foreground">{i18next.t("enterprise:Organization admins")}</span>}</TableCell>
                <TableCell className="whitespace-nowrap">
                  <DecisionBadge decision={item.decision} />
                  {item.decidedBy ? <div className="text-xs text-muted-foreground">{item.decidedBy === "(deadline)" ? i18next.t("enterprise:By the deadline") : item.decidedBy}</div> : null}
                </TableCell>
                <TableCell className="max-w-64 truncate" title={item.comment}>{item.comment}</TableCell>
                {showApplied ? (
                  <TableCell className="whitespace-nowrap">
                    {item.applied ? <Badge variant="secondary">{i18next.t("enterprise:Applied")}</Badge> : item.applyResult ? <span className="text-xs text-destructive">{item.applyResult}</span> : null}
                  </TableCell>
                ) : null}
                {canDecide ? (
                  <TableCell className="whitespace-nowrap">
                    {decidable(item) ? (
                      <div className="flex gap-1">
                        <Button size="sm" variant="outline" disabled={busy} onClick={() => decide([item.name], "Keep")}>
                          {i18next.t("enterprise:Keep")}
                        </Button>
                        <Button size="sm" variant="outline" className="text-destructive" disabled={busy} onClick={() => decide([item.name], "Revoke")}>
                          {i18next.t("enterprise:Revoke")}
                        </Button>
                      </div>
                    ) : null}
                  </TableCell>
                ) : null}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
