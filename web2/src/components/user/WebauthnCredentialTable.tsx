import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import * as UserWebauthnBackend from "@/backend/UserWebauthnBackend";
import * as Setting from "@/lib/setting";

/**
 * The user page's WebAuthn credential list, ported from
 * `web/src/table/WebauthnCredentialTable.js`.
 *
 * Registering runs the browser's WebAuthn ceremony and is persisted by the backend
 * right away, so the page is reloaded afterwards; removing a row only edits the
 * user object, exactly as the antd table did, and takes effect on Save.
 */
export function WebauthnCredentialTable({
  table,
  isSelf,
  onUpdateTable,
  refresh,
}: {
  table: any[];
  isSelf: boolean;
  onUpdateTable: (rows: any[]) => void;
  refresh: () => void;
}) {
  const [registering, setRegistering] = React.useState(false);

  const register = () => {
    setRegistering(true);
    UserWebauthnBackend.registerWebauthnCredential()
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", "Successfully added webauthn credentials");
        } else {
          Setting.showMessage("error", res.msg);
        }
        refresh();
      })
      .catch((error) => Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`))
      .finally(() => setRegistering(false));
  };

  return (
    <div className="space-y-2">
      <div className="divide-y rounded-lg border">
        {(table ?? []).length === 0 ? (
          <div className="p-4 text-center text-sm text-muted-foreground">{i18next.t("general:No data")}</div>
        ) : (
          (table ?? []).map((item: any, index: number) => (
            <div key={item.id ?? item.ID ?? index} className="flex items-center justify-between gap-2 p-3">
              {/* the credential ID is stored as standard base64; show it as base64url
                  so it matches what the browser reports during the challenge */}
              <span className="truncate font-mono text-xs">
                {String(item.id ?? item.ID ?? "").replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "")}
              </span>
              <ConfirmButton
                variant="destructive"
                size="sm"
                description={String(item.id ?? item.ID ?? "")}
                onConfirm={() => onUpdateTable((table ?? []).filter((_, i) => i !== index))}
              >
                {i18next.t("general:Delete")}
              </ConfirmButton>
            </div>
          ))
        )}
      </div>
      <Button variant="outline" size="sm" disabled={!isSelf} loading={registering} onClick={register}>
        {i18next.t("general:Add")}
      </Button>
    </div>
  );
}
