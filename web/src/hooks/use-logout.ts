import i18next from "i18next";
import {useNavigate} from "react-router-dom";
import {useAccount} from "@/hooks/use-account";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Setting from "@/lib/setting";

/** Signs the current user out, honouring the application's post-logout redirect. */
export function useLogout() {
  const {account, setAccount} = useAccount();
  const navigate = useNavigate();

  return () => {
    AuthBackend.logout().then((res: any) => {
      if (res.status === "ok") {
        const owner = account?.owner;
        setAccount(null);
        Setting.showMessage("success", i18next.t("application:Logged out successfully"));
        const redirectUri = res.data2;
        if (redirectUri !== null && redirectUri !== undefined && redirectUri !== "") {
          Setting.goToLink(redirectUri);
        } else if (owner !== "built-in") {
          Setting.goToLink(`${window.location.origin}/login/${owner}`);
        } else {
          navigate("/");
        }
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to log out")}: ${res.msg}`);
      }
    });
  };
}
