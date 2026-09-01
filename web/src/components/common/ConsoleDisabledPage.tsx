import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {useLogout} from "@/hooks/use-logout";

/**
 * Shown in place of the whole console when the organization turns on "Disable
 * console": Casdoor is then only an identity provider for its regular users, so
 * signing out is the only thing left to offer them.
 */
export function ConsoleDisabledPage() {
  const logout = useLogout();

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-6 text-center">
      <div className="text-6xl font-semibold tracking-tight text-muted-foreground">403</div>
      <h1 className="text-xl font-semibold">{i18next.t("general:Unauthorized")}</h1>
      <p className="max-w-xl text-muted-foreground">
        {i18next.t("organization:The Casdoor console is disabled for the regular users of this organization")}
      </p>
      <Button onClick={logout}>{i18next.t("account:Logout")}</Button>
    </div>
  );
}
