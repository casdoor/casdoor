import * as React from "react";
import i18next from "i18next";
import {Link, useNavigate, useParams} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Card, CardContent} from "@/components/ui/card";
import {Checkbox} from "@/components/ui/checkbox";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
import * as LdapBackend from "@/backend/LdapBackend";
import * as Setting from "@/lib/setting";

/** Preview the users an LDAP server exposes and import the selected ones. */
export default function LdapSyncPage() {
  const {organizationName = "", ldapId = ""} = useParams();
  const navigate = useNavigate();
  const [users, setUsers] = React.useState<any[]>([]);
  const [existUuids, setExistUuids] = React.useState<string[]>([]);
  const [selected, setSelected] = React.useState<string[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [syncing, setSyncing] = React.useState(false);

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

  const syncable = users.filter((user) => !existUuids.includes(user.uuid));

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
        navigate(`/organizations/${organizationName}`);
      })
      .finally(() => setSyncing(false));
  };

  const allSelected = syncable.length > 0 && selected.length === syncable.length;

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
                    onCheckedChange={(v) => setSelected(v === true ? syncable.map((u) => u.uuid) : [])}
                  />
                </TableHead>
                <TableHead>{i18next.t("ldap:CN")}</TableHead>
                <TableHead>UidNumber / Uid</TableHead>
                <TableHead>{i18next.t("ldap:Group ID")}</TableHead>
                <TableHead>{i18next.t("general:Email")}</TableHead>
                <TableHead>{i18next.t("general:Phone")}</TableHead>
                <TableHead>{i18next.t("user:Address")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.length === 0 ? (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={7} className="h-24 text-center text-muted-foreground">
                    {i18next.t("general:No data")}
                  </TableCell>
                </TableRow>
              ) : (
                users.map((user) => {
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
                      <TableCell>{user.phone}</TableCell>
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
