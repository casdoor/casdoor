import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
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
    setSyncing(true);
    LdapBackend.syncUsers(
      organizationName,
      ldapId,
      users.filter((user) => selected.includes(user.uuid)),
    )
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully synced"));
          navigate(`/organizations/${organizationName}`);
        } else {
          Setting.showMessage("error", res.msg);
        }
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
          <Button loading={syncing} disabled={selected.length === 0} onClick={sync}>
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
                <TableHead>{i18next.t("general:Email")}</TableHead>
                <TableHead>{i18next.t("general:Phone")}</TableHead>
                <TableHead>{i18next.t("user:Address")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.length === 0 ? (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
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
                      <TableCell>{user.cn}</TableCell>
                      <TableCell>
                        {user.uidNumber} / {user.uid}
                      </TableCell>
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
