import * as React from "react";
import i18next from "i18next";
import {Link, useNavigate, useParams} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Card, CardContent} from "@/components/ui/card";
import {Checkbox} from "@/components/ui/checkbox";
import {DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger} from "@/components/ui/dropdown-menu";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {Filter, X} from "lucide-react";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
import * as LdapBackend from "@/backend/LdapBackend";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

/** Preview the users an LDAP server exposes and import the selected ones. */
export default function LdapSyncPage() {
  const {organizationName = "", ldapId = ""} = useParams();
  const navigate = useNavigate();
  const [users, setUsers] = React.useState<any[]>([]);
  const [existUuids, setExistUuids] = React.useState<string[]>([]);
  const [selected, setSelected] = React.useState<string[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [syncing, setSyncing] = React.useState(false);
  // the antd table offers a filter menu built from the group ids in the result
  const [groupFilter, setGroupFilter] = React.useState("");

  React.useEffect(() => {
    LdapBackend.getLdapUser(organizationName, ldapId)
      .then((res: any) => {
        if (res.status === "ok") {
          setUsers(res.data?.users ?? []);
          setExistUuids(res.data?.existUuids ?? []);
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setLoading(false));
  }, [organizationName, ldapId]);

  if (loading) {
    return <Loading />;
  }

  const sync = () => {
    if (selected.length === 0) {
      Setting.showMessage("error", i18next.t("general:Please select at least 1 user first"));
      return;
    }
    setSyncing(true);
    LdapBackend.syncUsers(
      organizationName,
      ldapId,
      users.filter((user) => selected.includes(user.uuid)),
    )
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", res.msg);
          return;
        }
        Setting.showMessage("success", i18next.t("general:Successfully synced"));
        // the backend reports the users it skipped and the ones it could not add
        const exist = res.data?.exist ?? [];
        const failed = res.data?.failed ?? [];
        if (exist.length > 0) {
          Setting.showMessage("error", `${i18next.t("general:User already exists")}: [${exist.map((u: any) => u.cn)}]`);
        }
        if (failed.length > 0) {
          Setting.showMessage("error", `${i18next.t("general:Failed to sync")}: [${failed.map((u: any) => u.cn)}]`);
        }
        // only a clean run leaves the page: the skipped and failed rows are worth
        // reading against the list that produced them
        if (exist.length === 0 && failed.length === 0) {
          navigate(`/organizations/${organizationName}/users`);
        }
      })
      .finally(() => setSyncing(false));
  };

  const groupIds = Array.from(new Set(users.map((user) => user.groupId).filter(Boolean))) as string[];
  const shownUsers = groupFilter === ""
    ? users
    : users.filter((user) => `${user.groupId ?? ""}`.indexOf(groupFilter) === 0);
  const shownSyncable = shownUsers.filter((user) => !existUuids.includes(user.uuid));
  const allSelected = shownSyncable.length > 0 && shownSyncable.every((user) => selected.includes(user.uuid));

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("general:Sync")}
        description={`${organizationName} / ${ldapId}`}
        actions={
          <Button loading={syncing} onClick={sync}>
            {i18next.t("general:Sync")} ({selected.length})
          </Button>
        }
      />
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="w-10">
                  <Checkbox
                    checked={allSelected}
                    onCheckedChange={(v) => setSelected(v === true ? shownSyncable.map((u) => u.uuid) : [])}
                  />
                </TableHead>
                <TableHead>{i18next.t("ldap:CN")}</TableHead>
                <TableHead>UidNumber / Uid</TableHead>
                <TableHead>
                  <span className="inline-flex items-center gap-1">
                    {i18next.t("ldap:Group ID")}
                    {groupIds.length > 0 ? (
                      <DropdownMenu>
                        <DropdownMenuTrigger
                          aria-label={i18next.t("general:Filter")}
                          className={cn(
                            "inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground/50 hover:bg-accent hover:text-foreground",
                            groupFilter !== "" && "bg-accent text-foreground",
                          )}
                        >
                          <Filter className="h-3.5 w-3.5" />
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="start">
                          {groupIds.map((groupId) => (
                            <DropdownMenuItem key={groupId} onSelect={() => setGroupFilter(groupId)}>
                              {groupId}
                            </DropdownMenuItem>
                          ))}
                          <DropdownMenuItem disabled={groupFilter === ""} onSelect={() => setGroupFilter("")}>
                            <X className="mr-1 h-3.5 w-3.5" />
                            {i18next.t("forget:Reset")}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    ) : null}
                  </span>
                </TableHead>
                <TableHead>{i18next.t("general:Email")}</TableHead>
                <TableHead>{i18next.t("general:Phone")}</TableHead>
                <TableHead>{i18next.t("user:Address")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {shownUsers.length === 0 ? (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={7} className="h-24 text-center text-muted-foreground">
                    {i18next.t("general:No data")}
                  </TableCell>
                </TableRow>
              ) : (
                shownUsers.map((user) => {
                  const exists = existUuids.includes(user.uuid);
                  return (
                    <TableRow key={user.uuid} className={exists ? "opacity-50" : undefined}>
                      <TableCell>
                        <Checkbox
                          disabled={exists}
                          checked={selected.includes(user.uuid)}
                          onCheckedChange={(v) =>
                            setSelected((prev) =>
                              v === true ? [...prev, user.uuid] : prev.filter((id) => id !== user.uuid),
                            )
                          }
                        />
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center justify-between gap-2">
                          <span>{user.cn}</span>
                          <Badge variant={exists ? "success" : "destructive"}>
                            {i18next.t(exists ? "ldap:synced" : "ldap:unsynced")}
                          </Badge>
                        </div>
                      </TableCell>
                      <TableCell>
                        {user.uidNumber} /{" "}
                        {exists ? (
                          <Link
                            to={`/users/${organizationName}/${user.uid}`}
                            className="underline-offset-4 hover:underline"
                          >
                            {user.uid}
                          </Link>
                        ) : (
                          user.uid
                        )}
                      </TableCell>
                      <TableCell>{user.groupId}</TableCell>
                      <TableCell>{user.email}</TableCell>
                      {/* the LDAP user's phone arrives as "mobile", see object/ldap_conn.go */}
                      <TableCell>{user.mobile}</TableCell>
                      <TableCell>{user.address}</TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
