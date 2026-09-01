import * as React from "react";
import i18next from "i18next";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Setting from "@/lib/setting";
import * as TourConfig from "@/lib/tour-config";

export interface Account {
  owner: string;
  name: string;
  organization?: Record<string, any>;
  [key: string]: any;
}

interface AccountContextValue {
  /** undefined = still loading, null = not signed in */
  account: Account | null | undefined;
  accessToken: string | null;
  loading: boolean;
  setAccount: (account: Account | null) => void;
  reload: () => Promise<void>;
}

const AccountContext = React.createContext<AccountContextValue>({
  account: undefined,
  accessToken: null,
  loading: true,
  setAccount: () => undefined,
  reload: async() => undefined,
});

function getUrlWithoutQuery() {
  return window.location.toString().replace(window.location.search, "");
}

// Casdoor supports signing in through query params (accessToken / username+password),
// the same way the antd frontend did. The params are consumed once and then stripped
// from the address bar so that they are not leaked through the history or a bookmark.
function consumeQueryCredentials(): string {
  const params = new URLSearchParams(window.location.search);

  let query = "";
  const accessToken = params.get("access_token");
  if (accessToken !== null && accessToken !== undefined && accessToken !== "") {
    query = `?accessToken=${accessToken}`;
  } else {
    const username = params.get("username");
    const password = params.get("password");
    if (username !== null && password !== null) {
      query = `?username=${username}&password=${password}`;
    }
  }

  const language = params.get("language");
  if (language !== null && language !== "") {
    Setting.setLanguage(language);
    const url = window.location.toString().replace(new RegExp(`[?&]language=${language}`), "");
    window.history.replaceState({}, document.title, url);
  }

  if (query !== "") {
    const returnUrl = params.get("returnUrl");
    let newUrl: string;
    if (returnUrl) {
      const newParams = new URLSearchParams();
      newParams.set("returnUrl", returnUrl);
      newUrl = window.location.pathname + "?" + newParams.toString();
    } else {
      newUrl = getUrlWithoutQuery();
    }
    window.history.replaceState({}, document.title, newUrl);
  }

  return query;
}

export function AccountProvider({children}: {children: React.ReactNode}) {
  const [account, setAccountState] = React.useState<Account | null | undefined>(undefined);
  const [accessToken, setAccessToken] = React.useState<string | null>(null);
  const [loading, setLoading] = React.useState(true);

  const fetchAccount = React.useCallback(async(query = "") => {
    setLoading(true);
    try {
      const res = await AuthBackend.getAccount(query);
      if (res.status === "ok") {
        const next = res.data;
        next.organization = res.data2;
        setAccountState(next);
        setAccessToken(res.data.accessToken ?? null);
        if (!localStorage.getItem("language") && next.language) {
          Setting.setLanguage(next.language);
        }
        // the organization decides whether the product tour runs at all
        TourConfig.setTourLogo(next.organization?.logo ?? "");
        TourConfig.setOrgIsTourVisible(next.organization?.enableTour);
      } else {
        setAccountState(null);
        if (res.data !== "Please login first" && res.msg) {
          Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${res.msg}`);
        }
      }
    } catch (e: any) {
      setAccountState(null);
      Setting.showMessage("error", e?.message ?? String(e));
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    const query = consumeQueryCredentials();
    fetchAccount(query);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const value = React.useMemo<AccountContextValue>(
    () => ({
      account,
      accessToken,
      loading,
      setAccount: (next) => setAccountState(next),
      reload: () => fetchAccount(""),
    }),
    [account, accessToken, loading, fetchAccount],
  );

  return <AccountContext.Provider value={value}>{children}</AccountContext.Provider>;
}

export function useAccount() {
  return React.useContext(AccountContext);
}
