import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams, useSearchParams} from "react-router-dom";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {Loading} from "@/components/common/Loading";
import {useAccount} from "@/hooks/use-account";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Setting from "@/lib/setting";

/**
 * CAS single logout, ported from web/src/auth/CasLogout.js: it keeps calling
 * /api/logout until no account is left, following each redirect the backend
 * hands back, then returns to the service or the CAS login page.
 */
export default function CasLogout() {
  const {owner = "", casApplicationName = ""} = useParams();
  const [search] = useSearchParams();
  const navigate = useNavigate();
  const {reload} = useAccount();

  React.useEffect(() => {
    let cancelled = false;

    const finish = (redirectUri?: string) => {
      Setting.showMessage("success", i18next.t("application:Logged out successfully"));
      reload();
      if (redirectUri) {
        Setting.goToLink(redirectUri);
      } else if (search.get("service")) {
        Setting.goToLink(search.get("service") as string);
      } else {
        navigate(`/cas/${owner}/${casApplicationName}/login`);
      }
    };

    const logoutLoop = (redirectUri?: string) => {
      window.setTimeout(() => {
        if (cancelled) {
          return;
        }
        AuthBackend.getAccount().then((accountRes: any) => {
          if (cancelled) {
            return;
          }
          if (accountRes.status !== "ok") {
            finish(redirectUri);
            return;
          }
          AuthBackend.logout().then((logoutRes: any) => {
            if (logoutRes.status === "ok") {
              logoutLoop(logoutRes.data2);
            } else {
              Setting.showMessage("error", `${i18next.t("general:Failed to log out")}: ${logoutRes.msg}`);
            }
          });
        });
      }, 100);
    };

    AuthBackend.logout().then((res: any) => {
      if (cancelled) {
        return;
      }
      if (res.status === "ok") {
        logoutLoop(res.data2);
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to log out")}: ${res.msg}`);
      }
    });

    return () => {
      cancelled = true;
    };
    // the logout runs once for this route
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <AuthLayout>
      <div className="space-y-3 text-center">
        <Loading />
        <p className="text-sm text-muted-foreground">{i18next.t("login:Logging out...")}</p>
      </div>
    </AuthLayout>
  );
}
