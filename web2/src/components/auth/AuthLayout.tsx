import * as React from "react";
import {Card, CardContent} from "@/components/ui/card";
import {LanguageSelect} from "@/components/common/LanguageSelect";
import {ThemeToggle} from "@/components/common/ThemeToggle";
import * as Setting from "@/lib/setting";
import * as Conf from "@/Conf";
import {cn} from "@/lib/utils";

interface AuthLayoutProps {
  application?: any;
  children: React.ReactNode;
  className?: string;
  /** widen the card for the signup form */
  wide?: boolean;
}

/** Centered panel shared by the sign-in, sign-up, forget-password and result pages. */
export function AuthLayout({application, children, className, wide}: AuthLayoutProps) {
  const logo = application?.logo || `${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256.png`;

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
          {application?.footerHtml ? (
            <div
              className="mt-6 text-center text-xs text-muted-foreground"
              dangerouslySetInnerHTML={{__html: application.footerHtml}}
            />
          ) : (
            <div className="mt-6 text-center text-xs text-muted-foreground">
              {Conf.CustomFooter ?? (
                <>
                  Powered by{" "}
                  <a href="https://casdoor.org" target="_blank" rel="noreferrer" className="hover:underline">
                    Casdoor
                  </a>
                </>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
