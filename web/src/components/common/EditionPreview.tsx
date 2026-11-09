import i18next from "i18next";
import {Check, ChevronDown, Sparkles} from "lucide-react";
import * as Conf from "@/Conf";
import {getEditionText} from "@/components/common/EditionNotice";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {Button} from "@/components/ui/button";
import {cn} from "@/lib/utils";

/** Only a demo site lets the visitor switch editions, see Conf.applyDemoEdition(). */
export function isEditionPreviewEnabled(): boolean {
  return Conf.ThemeShowcase;
}

/** With a demo site for each edition the switch goes there, otherwise only the look changes. */
export function getEditionPreviewHint(): string {
  return Object.keys(Conf.EditionDemoUrls).length > 0
    ? i18next.t("system:Each edition has a demo site of its own, switch to try its features")
    : i18next.t("system:Try the look of each edition on this demo site, paid features need a license");
}

/** The three editions side by side, for the sign-in pages and the license card. */
export function EditionSwitcher({className}: {className?: string}) {
  if (!isEditionPreviewEnabled()) {
    return null;
  }

  return (
    <div
      role="radiogroup"
      aria-label={i18next.t("system:Edition")}
      title={getEditionPreviewHint()}
      className={cn("edition-switcher inline-flex items-center gap-0.5 rounded-full border bg-background/80 p-0.5 text-xs shadow-sm backdrop-blur", className)}
    >
      <Sparkles className="ml-2 mr-1 size-3.5 shrink-0 text-muted-foreground" />
      {Conf.Editions.map((edition) => {
        const active = edition === Conf.Edition;
        return (
          <button
            key={edition}
            type="button"
            role="radio"
            aria-checked={active}
            className={cn(
              "whitespace-nowrap rounded-full px-3 py-1 font-medium transition-colors",
              active ? "bg-primary text-primary-foreground shadow-sm" : "text-muted-foreground hover:bg-accent hover:text-foreground",
            )}
            onClick={() => {
              if (!active) {
                Conf.setDemoEdition(edition);
              }
            }}
          >
            {getEditionText(edition)}
          </button>
        );
      })}
    </div>
  );
}

/** The console header's switcher, a menu to save the room three buttons would take. */
export function EditionMenu() {
  if (!isEditionPreviewEnabled()) {
    return null;
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="hidden h-8 gap-1.5 rounded-full px-3 text-xs md:inline-flex">
          <Sparkles className="size-3.5" />
          {getEditionText(Conf.Edition)}
          <ChevronDown className="size-3.5 opacity-60" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-64">
        <DropdownMenuLabel className="font-normal">
          <div className="text-sm font-medium">{i18next.t("system:Edition")}</div>
          <div className="mt-0.5 text-xs text-muted-foreground">
            {getEditionPreviewHint()}
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {Conf.Editions.map((edition) => (
          <DropdownMenuItem
            key={edition}
            onSelect={() => {
              if (edition !== Conf.Edition) {
                Conf.setDemoEdition(edition);
              }
            }}
          >
            <span className="flex-1">{getEditionText(edition)}</span>
            {edition === Conf.Edition ? <Check className="size-4" /> : null}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
