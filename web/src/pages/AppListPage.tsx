import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {GridCards, type GridCardItem} from "@/components/common/GridCards";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
import {useAccount} from "@/hooks/use-account";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

export default function AppListPage() {
  const {account} = useAccount();
  const [applications, setApplications] = React.useState<any[] | null>(null);
  const [selectedTags, setSelectedTags] = React.useState<string[]>([]);

  React.useEffect(() => {
    if (!account) {
      return;
    }
    ApplicationBackend.getApplicationsByOrganization("admin", account.owner).then((res: any) => {
      const items = [...(res.data ?? [])].sort((a: any, b: any) => a.order - b.order);
      setApplications(items);
    });
  }, [account]);

  if (applications === null) {
    return <Loading />;
  }

  const allTags = Array.from(
    new Set(applications.flatMap((application: any) => (Array.isArray(application.tags) ? application.tags : []))),
  );

  const filtered =
    selectedTags.length === 0
      ? applications
      : applications.filter((application: any) =>
        selectedTags.every((tag) => Array.isArray(application.tags) && application.tags.includes(tag)),
      );

  const items: GridCardItem[] = filtered.map((application: any) => ({
    link: Setting.getLoginLink(application),
    name: application.displayName || application.name,
    description: application.description,
    logo: application.logo,
    createdTime: application.createdTime,
    isExternal: true,
  }));

  return (
    <div className="space-y-4">
      <PageHeader title={i18next.t("general:Apps")} />
      {allTags.length > 0 ? (
        <div className="flex flex-wrap gap-2">
          {allTags.map((tag) => {
            const active = selectedTags.includes(tag);
            return (
              <button
                key={tag}
                type="button"
                onClick={() =>
                  setSelectedTags((prev) => (active ? prev.filter((t) => t !== tag) : [...prev, tag]))
                }
              >
                <Badge variant={active ? "default" : "outline"} className={cn("cursor-pointer")}>
                  {tag}
                </Badge>
              </button>
            );
          })}
        </div>
      ) : null}
      <GridCards items={items} />
    </div>
  );
}
