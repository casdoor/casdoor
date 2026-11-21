import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import * as MySessionBackend from "@/backend/MySessionBackend";
import * as Setting from "@/lib/setting";

/** "Chrome on Windows" out of a user agent, good enough to tell one's own devices apart */
function describeUserAgent(userAgent: string): string {
  if (!userAgent) {
    return "";
  }

  const browsers: [RegExp, string][] = [[/Edg\//, "Edge"], [/OPR\//, "Opera"], [/Firefox\//, "Firefox"], [/Chrome\//, "Chrome"], [/Safari\//, "Safari"]];
  const systems: [RegExp, string][] = [[/Windows/, "Windows"], [/iPhone|iPad/, "iOS"], [/Android/, "Android"], [/Mac OS X/, "macOS"], [/Linux/, "Linux"]];
  const browser = browsers.find(([re]) => re.test(userAgent))?.[1];
  const system = systems.find(([re]) => re.test(userAgent))?.[1];
  if (!browser && !system) {
    return userAgent;
  }
  return browser && system ? `${browser} on ${system}` : (browser ?? system ?? "");
}

/** The browsers the current user is signed in on, each of which can be signed out from here. */
export function MySessionTable() {
  const [sessions, setSessions] = React.useState<any[] | null>(null);

  const load = React.useCallback(() => {
    MySessionBackend.getMySessions().then((res: any) => {
      if (res.status === "ok") {
        setSessions(res.data ?? []);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  }, []);

  React.useEffect(load, [load]);

  const signOut = (sessionId: string) => {
    return MySessionBackend.deleteMySession(sessionId).then((res: any) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("user:Signed out of the session"));
        load();
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  const rows = sessions ?? [];
  return (
    <div className="overflow-x-auto rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead>{i18next.t("entry:Device")}</TableHead>
            <TableHead>{i18next.t("general:Applications")}</TableHead>
            <TableHead>{i18next.t("general:Last active")}</TableHead>
            <TableHead className="w-[140px]">{i18next.t("general:Action")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow className="hover:bg-transparent">
              <TableCell colSpan={4} className="h-20 text-center text-muted-foreground">
                {sessions === null ? i18next.t("general:Loading") : i18next.t("general:No data")}
              </TableCell>
            </TableRow>
          ) : (
            rows.map((session: any) => (
              <TableRow key={session.sessionId}>
                <TableCell>
                  <div className="flex flex-col gap-0.5">
                    <span className="flex items-center gap-2">
                      {describeUserAgent(session.userAgent) || "-"}
                      {session.isCurrent ? <Badge variant="secondary">{i18next.t("user:This device")}</Badge> : null}
                    </span>
                    <span className="text-xs text-muted-foreground">{session.ip}</span>
                  </div>
                </TableCell>
                <TableCell>{(session.applications ?? []).join(", ")}</TableCell>
                <TableCell>{Setting.getFormattedDate(session.lastActiveTime || session.createdTime) ?? ""}</TableCell>
                <TableCell>
                  {session.isCurrent ? null : (
                    <ConfirmButton
                      variant="outline"
                      size="sm"
                      title={i18next.t("user:Sign out of this session") + "?"}
                      description={i18next.t("user:The applications signed in on that device are signed out too.")}
                      onConfirm={() => signOut(session.sessionId)}
                    >
                      {i18next.t("account:Logout")}
                    </ConfirmButton>
                  )}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
