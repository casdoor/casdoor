import * as React from "react";
import i18next from "i18next";
import {AlertCircle} from "lucide-react";
import {PoweredBy} from "@/components/layout/PoweredBy";
import {Button} from "@/components/ui/button";
import {Card, CardContent} from "@/components/ui/card";
import {LanguageSelect} from "@/components/common/LanguageSelect";
import {ThemeToggle} from "@/components/common/ThemeToggle";
import {
  getOrganizationCookieChrome,
  useApplicationHelmet,
  useApplicationTheme,
  useCustomHead,
} from "@/hooks/use-application-chrome";
import {useIsDark} from "@/hooks/use-theme";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

interface AuthLayoutProps {
  application?: any;
  children: React.ReactNode;
  className?: string;
  /** widen the card for the signup form */
  wide?: boolean;
}

/** The "we cannot sign you in" panel, port of auth/Util.js renderMessageLarge(). */
function BlockedMessage({message}: {message: string}) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 px-4 text-center">
      <AlertCircle className="h-12 w-12 text-destructive" />
      <h1 className="text-xl font-semibold">{i18next.t("general:There was a problem signing you in..")}</h1>
      <p className="max-w-lg text-sm text-muted-foreground">{message}</p>
      <Button onClick={() => window.history.go(-2)}>{i18next.t("general:Back")}</Button>
    </div>
  );
}

/** Centered panel shared by the sign-in, sign-up, forget-password and result pages. */
export function AuthLayout({application, children, className, wide}: AuthLayoutProps) {
  useApplicationTheme(application);
  // an application can force the dark palette regardless of the visitor's own
  // preference, and the logo has to follow the palette that is actually painted
  const isDark = useIsDark();
  useApplicationHelmet(application);
  // headerHtml is the organization/application chrome, pageHtml is the per-page one
  useCustomHead(application?.headerHtml, "header");
  useCustomHead(application?.pageHtml, "page");

  // the backend hands us the organization's branding in cookies so the first
  // paint is already branded, before /api/get-application has come back
  const cookieChrome = React.useMemo(() => getOrganizationCookieChrome(), []);

  // an IP restriction on either the application or its organization blocks the
  // whole sign-in surface, the same way EntryPage does in the antd frontend
  const ipRestriction = application?.ipRestriction || application?.organizationObj?.ipRestriction;
  if (ipRestriction) {
    return <BlockedMessage message={ipRestriction} />;
  }

  const logo = Setting.getThemedLogo(
    application?.logo || cookieChrome.logo,
    application?.logoDark || application?.organizationObj?.logoDark,
    [isDark ? "dark" : "light"],
  );
  const footerHtml = application?.footerHtml || cookieChrome.footerHtml;

  return (
    <div className="flex min-h-screen flex-col bg-muted/30">
      <div className="flex items-center justify-end gap-1 p-3">
        <LanguageSelect languages={application?.organizationObj?.languages} />
        <ThemeToggle />
      </div>

      <div className="flex flex-1 items-start justify-center px-4 pb-10 pt-4 sm:items-center sm:pt-0">
        <div className={cn("w-full", wide ? "max-w-xl" : "max-w-sm", className)}>
          <div className="mb-6 flex justify-center">
            {application?.homepageUrl ? (
              <a href={application.homepageUrl} target="_blank" rel="noreferrer">
                <img src={logo} alt={application?.displayName ?? "Casdoor"} className="h-12 object-contain" />
              </a>
            ) : (
              <img src={logo} alt={application?.displayName ?? "Casdoor"} className="h-12 object-contain" />
            )}
          </div>
          <Card className="shadow-md">
            <CardContent className="p-6">{children}</CardContent>
          </Card>
        </div>
      </div>

      {/* below the centred card and at the bottom of the viewport, as antd's
          Layout.Footer is a sibling of the Content it follows */}
      {footerHtml ? (
        <footer
          id="footer"
          className="shrink-0 py-4 text-center text-xs text-muted-foreground"
          dangerouslySetInnerHTML={{__html: footerHtml}}
        />
      ) : (
        <footer id="footer" className="shrink-0 py-4 text-center text-xs text-muted-foreground">
          <PoweredBy />
        </footer>
      )}
    </div>
  );
}
