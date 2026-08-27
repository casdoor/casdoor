import * as React from "react";
import i18next from "i18next";
import {ChevronRight, Folder, Users} from "lucide-react";
import {Link, useNavigate, useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Card, CardContent} from "@/components/ui/card";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
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
        const toNode = (group: any): GroupNode => ({
          key: `${group.owner}/${group.name}`,
          title: group.displayName || group.name,
          owner: group.owner,
          name: group.name,
          children: (group.children ?? []).map(toNode),
        });
        const nodes = (res.data ?? []).map(toNode);
        setTree(nodes);
        const initial = groupName
          ? nodes.flatMap(function flatten(n: GroupNode): GroupNode[] {
            return [n, ...(n.children ?? []).flatMap(flatten)];
          }).find((n: GroupNode) => n.name === groupName)
          : nodes[0];
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
          <Button variant="outline" asChild>
            <Link to="/groups">{i18next.t("general:Groups")}</Link>
          </Button>
        }
      />
      <div className="grid gap-4 lg:grid-cols-[320px_1fr]">
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

        <Card>
          <CardContent className="space-y-3 p-6">
            {selected === null ? (
              <p className="text-sm text-muted-foreground">{i18next.t("general:No data")}</p>
            ) : (
              <>
                <h2 className="text-lg font-semibold">{selected.title}</h2>
                <p className="text-sm text-muted-foreground">{selected.key}</p>
                <div className="flex gap-2">
                  <Button variant="outline" asChild>
                    <Link to={`/groups/${selected.owner}/${selected.name}`}>{i18next.t("general:Edit")}</Link>
                  </Button>
                  <Button asChild>
                    <Link to={`/organizations/${selected.owner}/users?groupName=${selected.name}`}>
                      <Users />
                      {i18next.t("general:Users")}
                    </Link>
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
