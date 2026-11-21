import i18next from "i18next";
import {HelpCircle} from "lucide-react";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";

/** The "Skip MFA" column title of the signin method and provider tables, with what skipping costs */
export function SkipMfaColumnTitle() {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="inline-flex items-center gap-1">
          {i18next.t("application:Skip MFA")}
          <HelpCircle className="size-3.5 text-muted-foreground" />
        </span>
      </TooltipTrigger>
      <TooltipContent className="max-w-xs">{i18next.t("application:Skip MFA - Tooltip")}</TooltipContent>
    </Tooltip>
  );
}
