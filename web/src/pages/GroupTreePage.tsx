import * as React from "react";
import i18next from "i18next";
import {ChevronRight, Folder} from "lucide-react";
import {Link, useNavigate, useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Card, CardContent} from "@/components/ui/card";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
import UserListPage from "@/pages/UserListPage";
import * as GroupBackend from "@/backend/GroupBackend";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

interface GroupNode {
  key: string;
  title: string;
  owner: string;
  name: string;
  children?: GroupNode[];
}

function TreeNode({
  node,
  selected,
  onSelect,
  depth = 0,
}: {
  node: GroupNode;
  selected: string;
  onSelect: (node: GroupNode) => void;
  depth?: number;
}) {
  const [open, setOpen] = React.useState(true);
  const hasChildren = (node.children?.length ?? 0) > 0;

  return (
    <li>
      <div
        className={cn(
          "flex cursor-pointer items-center gap-1 rounded-md px-2 py-1.5 text-sm hover:bg-accent",
          selected === node.key && "bg-accent font-medium",
        )}
        style={{paddingLeft: 8 + depth * 16}}
        onClick={() => onSelect(node)}
      >
        {hasChildren ? (
          <button
            type="button"
            className="shrink-0"
            onClick={(e) => {
              e.stopPropagation();
              setOpen((v) => !v);
            }}
          >
            <ChevronRight className={cn("h-3.5 w-3.5 transition-transform", open && "rotate-90")} />
          </button>
        ) : (
          <span className="w-3.5" />
        )}
        <Folder className="h-4 w-4 shrink-0 text-muted-foreground" />
        <span className="truncate">{node.title}</span>
      </div>
      {hasChildren && open ? (
        <ul>
          {node.children!.map((child) => (
            <TreeNode key={child.key} node={child} selected={selected} onSelect={onSelect} depth={depth + 1} />
          ))}
        </ul>
      ) : null}
    </li>
  );
}

/** Group hierarchy of an organization, with a shortcut to each group's users. */
export default function GroupTreePage() {
  const {organizationName = "", groupName} = useParams();
  const navigate = useNavigate();
  const [tree, setTree] = React.useState<GroupNode[] | null>(null);
  const [selected, setSelected] = React.useState<GroupNode | null>(null);

  React.useEffect(() => {
    GroupBackend.getGroups(organizationName, true).then((res: any) => {
      if (res.status === "ok") {
        const toNode = (group: any): GroupNode => {
          // The tree endpoint returns the group name in "key" and the display name in "title".
          const name = group.name || group.key || "";
          return {
            key: `${group.owner}/${name}`,
            title: group.displayName || group.title || name,
            owner: group.owner,
            name: name,
            children: (group.children ?? []).map(toNode),
          };
        };
        const nodes = (res.data ?? []).map(toNode);
        setTree(nodes);
        // no group in the route means "Show all", which lists the whole organization
        const initial = groupName
          ? nodes.flatMap(function flatten(n: GroupNode): GroupNode[] {
            return [n, ...(n.children ?? []).flatMap(flatten)];
          }).find((n: GroupNode) => n.name === groupName)
          : null;
        setSelected(initial ?? null);
      } else {
        Setting.showMessage("error", res.msg);
        setTree([]);
      }
    });
  }, [organizationName, groupName]);

  if (tree === null) {
    return <Loading />;
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("general:Groups")}
        description={organizationName}
        actions={
          <>
            {selected !== null ? (
              <Button variant="outline" asChild>
                <Link to={`/groups/${selected.owner}/${selected.name}`}>{i18next.t("general:Edit")}</Link>
              </Button>
            ) : null}
            <Button
              variant="outline"
              onClick={() => {
                setSelected(null);
                navigate(`/trees/${organizationName}`);
              }}
            >
              {i18next.t("group:Show all")}
            </Button>
            <Button variant="outline" asChild>
              <Link to="/groups">{i18next.t("general:Groups")}</Link>
            </Button>
          </>
        }
      />
      <div className="grid gap-4 lg:grid-cols-[320px_minmax(0,1fr)]">
        <Card>
          <CardContent className="p-2">
            {tree.length === 0 ? (
              <p className="p-4 text-center text-sm text-muted-foreground">{i18next.t("general:No data")}</p>
            ) : (
              <ul>
                {tree.map((node) => (
                  <TreeNode
                    key={node.key}
                    node={node}
                    selected={selected?.key ?? ""}
                    onSelect={(next) => {
                      setSelected(next);
                      navigate(`/trees/${organizationName}/${next.name}`, {replace: true});
                    }}
                  />
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        {/* the users of the selected group, or of the whole organization under "Show all" */}
        <div className="min-w-0">
          <UserListPage organizationName={organizationName} groupName={selected?.name ?? ""} />
        </div>
      </div>
    </div>
  );
}
