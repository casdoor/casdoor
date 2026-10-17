import * as React from "react";
import {createPortal} from "react-dom";
import i18next from "i18next";
import {ChevronLeft, ChevronRight, Pause, Play} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import {getLocalized} from "@/lib/integration";
import {ROTATE_MS, getThemeIds, isFramedShowcase, showcaseActions, useThemeShowcase} from "@/lib/theme-showcase";
import {cn} from "@/lib/utils";

/**
 * The demo site's theme switcher. It sits on the page body rather than inside the themed
 * page, so a theme's CSS never restyles it.
 */
export function ThemeShowcaseBar({className}: {className?: string}) {
  const showcase = useThemeShowcase();
  if (!showcase || isFramedShowcase()) {
    return null;
  }

  const ids = getThemeIds(showcase.themes);
  const index = Math.max(0, ids.indexOf(showcase.selected));
  const getName = (id: string) => {
    const theme = showcase.themes.find((item) => item.id === id);
    return theme ? getLocalized(theme.name) : i18next.t("general:Default");
  };
  const playLabel = showcase.playing ? i18next.t("theme:Pause themes") : i18next.t("theme:Play themes");

  return createPortal(
    <div
      className={cn("fixed bottom-3 left-3 z-50 sm:bottom-4 sm:left-4", className)}
      style={{fontFamily: "Inter, ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif", fontSize: 14}}
    >
      <style>{"@keyframes casdoor-theme-showcase{from{transform:scaleX(0)}to{transform:scaleX(1)}}"}</style>
      <div className="relative flex items-center gap-0.5 overflow-hidden rounded-full border bg-background/90 p-1 text-foreground shadow-lg backdrop-blur">
        <Button
          type="button"
          variant="ghost"
          size="iconSm"
          className="rounded-full"
          aria-label={i18next.t("theme:Previous theme")}
          title={i18next.t("theme:Previous theme")}
          onClick={() => showcaseActions.pick(ids[(index - 1 + ids.length) % ids.length])}
        >
          <ChevronLeft />
        </Button>
        <Select
          value={ids[index]}
          onValueChange={showcaseActions.pick}
          onOpenChange={(open) => open && showcaseActions.setPlaying(false)}
        >
          <SelectTrigger className="h-8 w-36 rounded-full border-0 bg-transparent shadow-none focus:ring-0" aria-label={i18next.t("theme:Theme")}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent side="top" className="max-h-80">
            {ids.map((id) => (
              <SelectItem key={id} value={id}>{getName(id)}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button
          type="button"
          variant="ghost"
          size="iconSm"
          className="rounded-full"
          aria-label={i18next.t("theme:Next theme")}
          title={i18next.t("theme:Next theme")}
          onClick={() => showcaseActions.pick(ids[(index + 1) % ids.length])}
        >
          <ChevronRight />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="iconSm"
          className="rounded-full"
          aria-label={playLabel}
          title={playLabel}
          onClick={() => showcaseActions.setPlaying(!showcase.playing)}
        >
          {showcase.playing ? <Pause /> : <Play />}
        </Button>
        {showcase.playing ? (
          <span
            key={`${showcase.shown}-${showcase.selected}`}
            className="pointer-events-none absolute inset-x-4 bottom-0 h-0.5 origin-left bg-primary"
            style={{animation: `casdoor-theme-showcase ${ROTATE_MS}ms linear`}}
          />
        ) : null}
      </div>
    </div>,
    document.body,
  );
}
