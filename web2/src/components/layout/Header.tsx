import * as React from "react";
import i18next from "i18next";
import * as Cookie from "cookie";
import {ChevronDown, LogOut, Menu, Settings} from "lucide-react";
import {useLocation, useNavigate} from "react-router-dom";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Button} from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {LanguageSelect} from "@/components/common/LanguageSelect";
import {OrganizationSelect} from "@/components/common/OrganizationSelect";
import {ThemeToggle} from "@/components/common/ThemeToggle";
import {useAccount} from "@/hooks/use-account";
import * as AuthBackend from "@/backend/AuthBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

export function Header({onOpenMobileNav}: {onOpenMobileNav: () => void}) {
  const {account, setAccount} = useAccount();
  const navigate = useNavigate();
  const location = useLocation();
  const [organization, setOrganizationState] = React.useState(() => Setting.getOrganization());

  if (!account) {
    return null;
  }

  const logout = () => {
    AuthBackend.logout().then((res: any) => {
      if (res.status === "ok") {
        const owner = account.owner;
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

  const exitImpersonation = () => {
    UserBackend.exitImpersonateUser(account.owner, account.name).then((res: any) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("account:Exit impersonation"));
        navigate("/");
        window.location.reload();
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  const impersonating = (() => {
    try {
      return Boolean(Cookie.parse(document.cookie)["impersonateUser"]);
    } catch {
      return false;
    }
  })();

  const avatarUrl = Setting.getEffectiveAvatarUrl(account);
  const showOrganizationSelect = Setting.isAdminUser(account) && !location.pathname.startsWith("/trees");

  return (
    <header className="sticky top-0 z-30 flex h-14 shrink-0 items-center gap-2 border-b bg-background/95 px-3 backdrop-blur supports-[backdrop-filter]:bg-background/80">
      <Button variant="ghost" size="iconSm" className="lg:hidden" onClick={onOpenMobileNav} aria-label="Menu">
        <Menu />
      </Button>

      <div className="ml-auto flex items-center gap-1.5">
        {showOrganizationSelect && (
          <OrganizationSelect
            withAll
            value={organization}
            className="hidden h-8 w-[190px] sm:flex"
            onChange={(value) => {
              setOrganizationState(value);
              Setting.setOrganization(value);
            }}
          />
        )}
        <LanguageSelect languages={account.organization?.languages} />
        <ThemeToggle />

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button type="button" className="ml-1 flex items-center gap-2 rounded-md p-1 hover:bg-accent">
              <Avatar className="h-8 w-8">
                {avatarUrl ? <AvatarImage src={avatarUrl} alt={account.name} /> : null}
                <AvatarFallback style={{backgroundColor: Setting.getAvatarColor(account.name), color: "#fff"}}>
                  {(account.name || "?").charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span className="hidden max-w-[160px] truncate text-sm md:inline">
                {account.displayName || account.name}
              </span>
              <ChevronDown className="h-3.5 w-3.5 opacity-60" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuLabel className="truncate font-normal">
              <div className="text-sm font-medium">{account.displayName || account.name}</div>
              <div className="truncate text-xs text-muted-foreground">
                {account.owner}/{account.name}
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={() => navigate("/account")}>
              <Settings />
              {i18next.t("account:My Account")}
            </DropdownMenuItem>
            {impersonating ? (
              <DropdownMenuItem onSelect={exitImpersonation}>
                <LogOut />
                {i18next.t("account:Exit impersonation")}
              </DropdownMenuItem>
            ) : (
              <DropdownMenuItem onSelect={logout}>
                <LogOut />
                {i18next.t("account:Logout")}
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
