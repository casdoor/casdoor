import * as React from "react";
import i18next from "i18next";
import {Switch} from "@/components/ui/switch";
import {Label} from "@/components/ui/label";
import {Loading} from "@/components/common/Loading";
import {EditionNotice} from "@/components/common/EditionNotice";
import {PageHeader} from "@/components/crud/PageHeader";
import {AccessReviewItemTable} from "@/components/access-review/AccessReviewItemTable";
import {useAccount} from "@/hooks/use-account";
import * as AccessReviewBackend from "@/backend/AccessReviewBackend";
import * as Setting from "@/lib/setting";

/** The items of the running access reviews that the signed-in user reviews, to keep or revoke. */
export default function MyAccessReviewsPage() {
  const {account} = useAccount();
  const [items, setItems] = React.useState<any[] | null>(null);
  const [pendingOnly, setPendingOnly] = React.useState(true);
  const [version, setVersion] = React.useState(0);

  React.useEffect(() => {
    AccessReviewBackend.getMyAccessReviewItems(pendingOnly).then((res: any) => {
      if (res.status === "ok") {
        setItems(res.data ?? []);
      } else {
        setItems([]);
        Setting.showMessage("error", res.msg);
      }
    });
  }, [pendingOnly, version]);

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("enterprise:My access reviews")}
        description={i18next.t("enterprise:My access reviews - Tooltip")}
        actions={
          <div className="flex items-center gap-2">
            <Switch id="pending-only" checked={pendingOnly} onCheckedChange={setPendingOnly} />
            <Label htmlFor="pending-only">{i18next.t("enterprise:Only undecided")}</Label>
          </div>
        }
      />
      <EditionNotice edition="Enterprise" />
      {items === null ? (
        <Loading />
      ) : (
        <AccessReviewItemTable
          items={items}
          canDecide
          showReview
          selfName={account?.name}
          onDecided={() => setVersion((v) => v + 1)}
        />
      )}
    </div>
  );
}
