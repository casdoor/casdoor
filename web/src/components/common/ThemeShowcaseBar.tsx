import * as React from "react";
import {createPortal} from "react-dom";
import i18next from "i18next";
import {Check, ChevronDown, ChevronUp, Loader2, PanelLeftClose, PanelLeftOpen, PanelRightClose, PanelRightOpen, Pause, Play} from "lucide-react";
import {Button} from "@/components/ui/button";
import {getLocalized} from "@/lib/integration";
import {ROTATE_MS, getThemeIds, isFramedShowcase, showcaseActions, useThemeShowcase} from "@/lib/theme-showcase";
import {cn} from "@/lib/utils";

const KEYFRAMES = "@keyframes casdoor-theme-showcase-x{from{transform:scaleX(0)}to{transform:scaleX(1)}}" +
  "@keyframes casdoor-theme-showcase-y{from{transform:scaleY(0)}to{transform:scaleY(1)}}";

/**
 * The demo site's theme switcher, a sidebar on the edge of the page: collapsed it steps through
 * the themes, expanded it lists every theme with its screenshot to pick from. It sits on the page
 * body rather than inside the themed page, so a theme's CSS never restyles it.
 */
export function ThemeShowcaseBar({side = "left"}: {side?: "left" | "right"}) {
  const showcase = useThemeShowcase();
  const listRef = React.useRef<HTMLDivElement>(null);
  const expanded = !!showcase?.expanded;
  const selected = showcase?.selected;

  // the theme on show stays in view as the themes take turns
  React.useEffect(() => {
    if (expanded && selected) {
      listRef.current?.querySelector(`[data-theme-id="${CSS.escape(selected)}"]`)?.scrollIntoView({block: "nearest"});
    }
  }, [expanded, selected]);

  if (!showcase || isFramedShowcase()) {
    return null;
  }

  const ids = getThemeIds(showcase.themes);
  const index = Math.max(0, ids.indexOf(showcase.selected));
  const getTheme = (id: string) => showcase.themes.find((item) => item.id === id);
  const getName = (id: string) => {
    const theme = getTheme(id);
    return theme ? getLocalized(theme.name) : i18next.t("general:Default");
  };
  const playLabel = showcase.playing ? i18next.t("theme:Pause themes") : i18next.t("theme:Play themes");
  const progressKey = `${showcase.shown}-${showcase.selected}`;
  const isLeft = side === "left";
  const OpenIcon = isLeft ? PanelLeftOpen : PanelRightOpen;
  const CloseIcon = isLeft ? PanelLeftClose : PanelRightClose;

  const playButton = (
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
  );

  const content = expanded ? (
    <div
      className={cn(
        "fixed inset-y-3 z-50 flex w-72 max-w-[calc(100vw-1.5rem)] flex-col overflow-hidden rounded-xl border bg-background/95 text-foreground shadow-2xl backdrop-blur",
        isLeft ? "left-3" : "right-3",
      )}
    >
      <div className="relative flex items-center gap-1 border-b py-2 pl-4 pr-2">
        <span className="flex-1 truncate font-semibold">
          {i18next.t("theme:All themes")}
          <span className="ml-2 font-normal text-muted-foreground">{showcase.themes.length}</span>
        </span>
        {playButton}
        <Button
          type="button"
          variant="ghost"
          size="iconSm"
          className="rounded-full"
          aria-label={i18next.t("general:Close")}
          title={i18next.t("general:Close")}
          onClick={() => showcaseActions.setExpanded(false)}
        >
          <CloseIcon />
        </Button>
        {showcase.playing ? (
          <span
            key={progressKey}
            className="pointer-events-none absolute inset-x-0 bottom-0 h-0.5 origin-left bg-primary"
            style={{animation: `casdoor-theme-showcase-x ${ROTATE_MS}ms linear`}}
          />
        ) : null}
      </div>
      <div ref={listRef} className="flex-1 overflow-y-auto p-3">
        <div className="grid gap-3">
          {ids.map((id) => {
            const theme = getTheme(id);
            const active = id === showcase.selected;
            const loading = active && showcase.selected !== showcase.shown;
            return (
              <button
                key={id}
                type="button"
                data-theme-id={id}
                aria-pressed={active}
                className={cn(
                  "rounded-lg border p-1.5 text-left transition-colors hover:bg-accent",
                  active && "border-primary ring-2 ring-primary",
                )}
                onClick={() => showcaseActions.pick(id)}
              >
                <div className="relative aspect-[16/10] overflow-hidden rounded-md bg-muted">
                  {theme?.screenshot ? (
                    <img src={theme.screenshot} alt="" loading="lazy" className="size-full object-cover object-top" />
                  ) : (
                    <div className="flex size-full items-center justify-center text-xs text-muted-foreground">{getName(id)}</div>
                  )}
                  {loading ? (
                    <div className="absolute inset-0 flex items-center justify-center bg-background/60">
                      <Loader2 className="size-5 animate-spin" />
                    </div>
                  ) : null}
                </div>
                <div className="flex items-center justify-between gap-2 px-1 pt-1.5 text-sm font-medium">
                  <span className="truncate">{getName(id)}</span>
                  {active ? <Check className="size-4 shrink-0 text-primary" /> : null}
                </div>
              </button>
            );
          })}
        </div>
      </div>
    </div>
  ) : (
    <div
      className={cn(
        "fixed bottom-3 z-50 sm:bottom-auto sm:top-1/2 sm:-translate-y-1/2",
        isLeft ? "left-3" : "right-3",
      )}
    >
      <div className="relative flex flex-col items-center gap-0.5 overflow-hidden rounded-full border bg-background/90 p-1 text-foreground shadow-lg backdrop-blur">
        <Button
          type="button"
          variant="ghost"
          size="iconSm"
          className="rounded-full"
          aria-label={i18next.t("theme:All themes")}
          title={i18next.t("theme:All themes")}
          onClick={() => showcaseActions.setExpanded(true)}
        >
          <OpenIcon />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="iconSm"
          className="rounded-full"
          aria-label={i18next.t("theme:Previous theme")}
          title={i18next.t("theme:Previous theme")}
          onClick={() => showcaseActions.pick(ids[(index - 1 + ids.length) % ids.length])}
        >
          <ChevronUp />
        </Button>
        <button
          type="button"
          className="max-h-32 truncate rounded-full px-1 py-2 text-xs font-medium hover:bg-accent"
          style={{writingMode: "vertical-rl"}}
          title={i18next.t("theme:All themes")}
          onClick={() => showcaseActions.setExpanded(true)}
        >
          {getName(showcase.selected)}
        </button>
        <Button
          type="button"
          variant="ghost"
          size="iconSm"
          className="rounded-full"
          aria-label={i18next.t("theme:Next theme")}
          title={i18next.t("theme:Next theme")}
          onClick={() => showcaseActions.pick(ids[(index + 1) % ids.length])}
        >
          <ChevronDown />
        </Button>
        {playButton}
        {showcase.playing ? (
          <span
            key={progressKey}
            className={cn("pointer-events-none absolute inset-y-4 w-0.5 origin-top bg-primary", isLeft ? "right-0" : "left-0")}
            style={{animation: `casdoor-theme-showcase-y ${ROTATE_MS}ms linear`}}
          />
        ) : null}
      </div>
    </div>
  );

  return createPortal(
    <div style={{fontFamily: "Inter, ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif", fontSize: 14}}>
      <style>{KEYFRAMES}</style>
      {content}
    </div>,
    document.body,
  );
}
