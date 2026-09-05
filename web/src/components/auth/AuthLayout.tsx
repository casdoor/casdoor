import * as React from "react";
import i18next from "i18next";
import {AlertCircle} from "lucide-react";
import {PoweredBy} from "@/components/layout/PoweredBy";
import {Button} from "@/components/ui/button";
import {Card, CardContent} from "@/components/ui/card";
import {LanguageSelect} from "@/components/common/LanguageSelect";
import {ThemeToggle} from "@/components/common/ThemeToggle";
import {CustomHtml, CustomStyle} from "@/components/common/CustomHtml";
import {
  getOrganizationCookieChrome,
  useApplicationHelmet,
  useApplicationTheme,
  useCustomHead,
} from "@/hooks/use-application-chrome";
import {useIsMobile} from "@/hooks/use-mobile";
import {useIsDark} from "@/hooks/use-theme";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

interface AuthLayoutProps {
  application?: any;
  children: React.ReactNode;
  className?: string;
  /** widen the card for the signup form */
  wide?: boolean;
  /** the sign-in page drives these two slots from its own signinItems */
  hideLogo?: boolean;
  hideLanguages?: boolean;
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
export function AuthLayout({application, children, className, wide, hideLogo, hideLanguages}: AuthLayoutProps) {
  useApplicationTheme(application);
  // an application can force the dark palette regardless of the visitor's own
  // preference, and the logo has to follow the palette that is actually painted
  const isDark = useIsDark();
  const isMobile = useIsMobile();
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

  // the customized form styling is meant for the standalone page, an embedded
  // one keeps the host page's own look, as the antd pages did
  const embedded = Setting.inIframe();
  const backgroundUrl = embedded
    ? undefined
    : (isMobile ? application?.formBackgroundUrlMobile : application?.formBackgroundUrl);
  // formOffset: 1 left, 2 center, 3 right, 4 center with the side panel
  const offset = embedded || isMobile ? 2 : (application?.formOffset ?? 2);
  const sidePanel = offset === 4 && application?.formSideHtml;

  return (
    <div
      className={cn("login-background flex min-h-screen flex-col bg-muted/30", backgroundUrl && "bg-cover bg-fixed bg-center")}
      style={backgroundUrl ? {backgroundImage: `url(${backgroundUrl})`} : undefined}
    >
      {embedded ? null : <CustomStyle css={isMobile ? application?.formCssMobile : application?.formCss} />}

      <div className="flex items-center justify-end gap-1 p-3">
        {hideLanguages ? null : <LanguageSelect languages={application?.organizationObj?.languages} />}
        <ThemeToggle />
      </div>

      <div
        className={cn(
          "login-content flex flex-1 items-start px-4 pb-10 pt-4 sm:items-center sm:pt-0",
          offset === 1 ? "justify-start sm:pl-[10%]" : offset === 3 ? "justify-end sm:pr-[10%]" : "justify-center",
        )}
      >
        <div className={cn("login-panel flex w-full items-center gap-8", sidePanel ? "max-w-4xl" : wide ? "max-w-xl" : "max-w-sm")}>
          {sidePanel ? (
            <CustomHtml html={application.formSideHtml} className="side-image hidden w-[420px] shrink-0 lg:block" />
          ) : null}
          <div className={cn("login-form w-full", className)}>
            {hideLogo ? null : (
              <div className="login-logo-box mb-6 flex justify-center">
                {application?.homepageUrl ? (
                  <a href={application.homepageUrl} target="_blank" rel="noreferrer">
                    <img src={logo} alt={application?.displayName ?? "Casdoor"} className="h-12 object-contain" />
                  </a>
                ) : (
                  <img src={logo} alt={application?.displayName ?? "Casdoor"} className="h-12 object-contain" />
                )}
              </div>
            )}
            <Card className="shadow-md">
              <CardContent className="p-6">{children}</CardContent>
            </Card>
          </div>
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
