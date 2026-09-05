import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {GridCards, type GridCardItem} from "@/components/common/GridCards";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
import {useAccount} from "@/hooks/use-account";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import {cn} from "@/lib/utils";

/** the antd page's generateTagColor(): a djb2 hash into a fixed palette */
const TAG_COLORS = [
  "#ff4d4f", "#f5222d", "#ff7a45", "#fa541c",
  "#ffa940", "#fa8c16", "#ffc53d", "#faad14",
  "#ffec3d", "#fadb14", "#bae637", "#a0d911",
  "#73d13d", "#52c41a", "#36cfc9", "#13c2c2",
  "#40a9ff", "#1890ff", "#f759ab", "#eb2f96",
];

function tagColor(tag: string) {
  let hash = 5381;
  for (let i = 0; i < tag.length; i++) {
    hash = (hash * 33) ^ tag.charCodeAt(i);
  }
  return TAG_COLORS[Math.abs(hash) % TAG_COLORS.length];
}

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
    // the card opens the application itself, not Casdoor's sign-in page
    link: application.homepageUrl === "<custom-url>" ? account?.homepage ?? "" : application.homepageUrl,
    name: application.displayName || application.name,
    description: application.description,
    logo: application.logo,
    tags: (Array.isArray(application.tags) ? application.tags : []).map((tag: string) => ({
      name: tag,
      color: tagColor(tag),
    })),
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
          {selectedTags.length > 0 ? (
            <Button variant="ghost" size="sm" className="h-6" onClick={() => setSelectedTags([])}>
              {i18next.t("forget:Reset")}
            </Button>
          ) : null}
        </div>
      ) : null}
      <GridCards items={items} />
    </div>
  );
}
