import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import * as ConsentBackend from "@/backend/ConsentBackend";
import * as Setting from "@/lib/setting";

interface ConsentTableProps {
  /** the user's `applicationScopes` */
  table: any[];
  onUpdateTable: () => void;
}

/**
 * The OAuth consents a user has granted, ported from web/src/table/ConsentTable.js.
 * A scope tag revokes just that scope, the Delete button revokes the whole consent.
 */
export function ConsentTable({table, onUpdateTable}: ConsentTableProps) {
  const rows = table ?? [];

  const revoke = (record: any, scopeToDelete?: string) => {
    return ConsentBackend.revokeConsent({
      application: record.application,
      grantedScopes: scopeToDelete ? [scopeToDelete] : record.grantedScopes,
    })
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully revoked"));
          onUpdateTable();
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  return (
    <div className="overflow-x-auto rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="w-[200px]">{i18next.t("general:Application")}</TableHead>
            <TableHead>{i18next.t("consent:Granted scopes")}</TableHead>
            <TableHead className="w-[120px]">{i18next.t("general:Action")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow className="hover:bg-transparent">
              <TableCell colSpan={3} className="h-20 text-center text-muted-foreground">
                {i18next.t("general:No data")}
              </TableCell>
            </TableRow>
          ) : (
            rows.map((record: any) => (
              <TableRow key={record.application}>
                <TableCell>{record.application}</TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1">
                    {(Array.isArray(record.grantedScopes) ? record.grantedScopes : []).map((scope: string) => (
                      <ConfirmButton
                        key={scope}
                        variant="ghost"
                        size="sm"
                        className="h-auto p-0"
                        title={`${i18next.t("consent:Are you sure you want to revoke scope")}: ${scope}?`}
                        onConfirm={() => revoke(record, scope)}
                      >
                        <Badge variant="secondary" className="cursor-pointer">{scope}</Badge>
                      </ConfirmButton>
                    ))}
                  </div>
                </TableCell>
                <TableCell>
                  <ConfirmButton
                    variant="destructive"
                    size="sm"
                    title={i18next.t("consent:Are you sure you want to revoke this consent?")}
                    onConfirm={() => revoke(record)}
                  >
                    {i18next.t("general:Delete")}
                  </ConfirmButton>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
