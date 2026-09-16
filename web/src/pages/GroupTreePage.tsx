import * as React from "react";
import dayjs from "dayjs";
import i18next from "i18next";
import {ChevronRight, Folder, Pencil, Plus, Trash2, Users} from "lucide-react";
import {Link, useLocation, useNavigate, useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Card, CardContent} from "@/components/ui/card";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {Loading} from "@/components/common/Loading";
import {OrganizationSelect} from "@/components/common/OrganizationSelect";
import {PageHeader} from "@/components/crud/PageHeader";
import UserListPage from "@/pages/UserListPage";
import * as GroupBackend from "@/backend/GroupBackend";
import {useAccount} from "@/hooks/use-account";
import {submitDelete} from "@/lib/crud";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

interface GroupNode {
  key: string;
  title: string;
  owner: string;
  name: string;
  type: string;
  children?: GroupNode[];
}

function NodeAction({label, children, ...props}: {label: string} & React.ComponentProps<typeof Button>) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button variant="ghost" size="iconSm" className="h-6 w-6" aria-label={label} {...props}>
          {children}
        </Button>
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}

function TreeNode({
  node,
  selected,
  onSelect,
  actions,
  depth = 0,
}: {
  node: GroupNode;
  selected: string;
  onSelect: (node: GroupNode) => void;
  actions: (node: GroupNode) => React.ReactNode;
  depth?: number;
}) {
  const [open, setOpen] = React.useState(true);
  const hasChildren = (node.children?.length ?? 0) > 0;
  const isSelected = selected === node.key;
  const Icon = node.type === "Physical" ? Users : Folder;

  return (
    <li>
      <div
        className={cn(
          "flex cursor-pointer items-center gap-1 rounded-md px-2 py-1.5 text-sm hover:bg-accent",
          isSelected && "bg-accent font-medium",
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
        <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
        <span className="truncate">{node.title}</span>
        {isSelected ? (
          <span className="ml-auto flex shrink-0 items-center gap-0.5" onClick={(e) => e.stopPropagation()}>
            {actions(node)}
          </span>
        ) : null}
      </div>
      {hasChildren && open ? (
        <ul>
          {node.children!.map((child) => (
            <TreeNode key={child.key} node={child} selected={selected} onSelect={onSelect} actions={actions} depth={depth + 1} />
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
  const location = useLocation();
  const {account} = useAccount();
  const [tree, setTree] = React.useState<GroupNode[] | null>(null);
  const [selected, setSelected] = React.useState<GroupNode | null>(null);
  const [version, setVersion] = React.useState(0);

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
            type: group.type,
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
  }, [organizationName, groupName, version]);

  // the edit page comes back here on "Save & Exit" / "Cancel" instead of to the group list
  const backTo = location.pathname;

  const addGroup = (isRoot: boolean) => {
    const randomName = Setting.getRandomName();
    const group = {
      owner: organizationName,
      name: `group_${randomName}`,
      createdTime: dayjs().format(),
      updatedTime: dayjs().format(),
      displayName: `New Group - ${randomName}`,
      type: "Virtual",
      parentId: isRoot || selected === null ? organizationName : selected.name,
      isTopGroup: isRoot || selected === null,
      isEnabled: true,
    };
    navigate(`/groups/${group.owner}/${group.name}`, {state: {mode: "add", record: group, backTo}});
  };

  const deleteGroup = async(node: GroupNode) => {
    await submitDelete({
      record: {owner: node.owner, name: node.name},
      remove: (record) => GroupBackend.deleteGroup(record),
      onDeleted: () => {
        setSelected(null);
        navigate(`/trees/${organizationName}`, {replace: true});
        setVersion((v) => v + 1);
      },
    });
  };

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
            {Setting.isAdminUser(account) ? (
              <OrganizationSelect
                value={organizationName}
                className="w-[190px]"
                onChange={(value) => {
                  setSelected(null);
                  navigate(`/trees/${value}`);
                }}
              />
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
            <Button onClick={() => addGroup(true)}>
              <Plus />
              {i18next.t("general:Add")}
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
                    actions={(current) => (
                      <>
                        <NodeAction label={i18next.t("general:Add")} onClick={() => addGroup(false)}>
                          <Plus />
                        </NodeAction>
                        <NodeAction
                          label={i18next.t("general:Edit")}
                          onClick={() => navigate(`/groups/${current.owner}/${current.name}`, {state: {backTo}})}
                        >
                          <Pencil />
                        </NodeAction>
                        {/* a group with subgroups cannot be deleted, as on the group list */}
                        {(current.children?.length ?? 0) === 0 ? (
                          <ConfirmButton
                            variant="ghost"
                            size="iconSm"
                            className="h-6 w-6 text-destructive hover:text-destructive"
                            aria-label={i18next.t("general:Delete")}
                            title={`${i18next.t("general:Sure to delete")}: ${current.title} ?`}
                            onConfirm={() => deleteGroup(current)}
                          >
                            <Trash2 />
                          </ConfirmButton>
                        ) : null}
                      </>
                    )}
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
