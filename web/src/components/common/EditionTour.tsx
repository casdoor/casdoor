import * as React from "react";
import i18next from "i18next";
import {ArrowLeft, ArrowRight, ExternalLink, Gem, LayoutList, Sparkles, X} from "lucide-react";
import {useLocation, useNavigate} from "react-router-dom";
import * as Conf from "@/Conf";
import {Button} from "@/components/ui/button";
import {getEditionText, useEdition} from "@/components/common/EditionNotice";
import {useAccount} from "@/hooks/use-account";
import {type EditionFeature, getEditionFeature, getObjectListPath, isEditionCovered, resolveObjectPath} from "@/lib/edition-features";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

const STORAGE_KEY = "casdoorEditionTour";
const CHANGE_EVENT = "editionTourChanged";
const PADDING = 6;
const GAP = 12;
const CARD_WIDTH = 360;
const FIND_TIMEOUT = 6000;

interface TourState {
  features: string[];
  index: number;
  // when the tour was started, so that starting it again on the same step still goes there
  started: number;
}

function readState(): TourState | null {
  try {
    const state = JSON.parse(sessionStorage.getItem(STORAGE_KEY) ?? "null");
    return state && Array.isArray(state.features) && state.features.length > 0 ? state : null;
  } catch {
    return null;
  }
}

function writeState(state: TourState | null) {
  try {
    if (state) {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    } else {
      sessionStorage.removeItem(STORAGE_KEY);
    }
  } catch {
    // the tour then only lasts until the next reload
  }
  memoryState = state;
  window.dispatchEvent(new Event(CHANGE_EVENT));
}

let memoryState: TourState | null = readState();

/** Starts the tour over the given features, from the one at index. */
export function startEditionTour(features: string[], index = 0) {
  const known = features.filter((id) => getEditionFeature(id)?.stop);
  if (known.length > 0) {
    writeState({features: known, index: Math.min(Math.max(index, 0), known.length - 1), started: Date.now()});
  }
}

export function stopEditionTour() {
  writeState(null);
}

function useTourState() {
  const [state, setState] = React.useState<TourState | null>(memoryState);
  React.useEffect(() => {
    const onChange = () => setState(memoryState);
    window.addEventListener(CHANGE_EVENT, onChange);
    return () => window.removeEventListener(CHANGE_EVENT, onChange);
  }, []);
  return state;
}

/** The demo site of the edition, when this is a demo site that has one for it, to try a feature this license lacks. */
export function getEditionDemoUrl(edition: string, path: string): string | undefined {
  const demoUrl = Conf.EditionDemoUrls[edition];
  if (!demoUrl || new URL(demoUrl).host === window.location.host) {
    return undefined;
  }
  const target = new URL(path, demoUrl);
  target.searchParams.set("edition", edition.toLowerCase());
  return target.toString();
}

function findTab(value: string): HTMLElement | null {
  return document.querySelector<HTMLElement>(`[role="tab"][id$="-trigger-${CSS.escape(value)}"]`);
}

/** Radix tabs switch on mousedown, not on click */
function openTab(tab: HTMLElement) {
  if (tab.getAttribute("data-state") === "active") {
    return;
  }
  tab.dispatchEvent(new MouseEvent("mousedown", {bubbles: true, cancelable: true, button: 0}));
  tab.click();
}

/** Whether the step had no object to open, so it went to the list of them instead. */
function isOnObjectList(feature: EditionFeature): boolean {
  const object = feature.stop?.object;
  return !!object && object !== "organization" && window.location.pathname === getObjectListPath(object);
}

/** The element the step points at, and whether it is the setting itself or only a stand-in until the page has loaded it. */
function findTarget(feature: EditionFeature): {element: Element | null; final: boolean} {
  const stop = feature.stop;
  if (!stop) {
    return {element: null, final: true};
  }
  if (isOnObjectList(feature)) {
    const table = document.querySelector("[data-tour='table']");
    return {element: table, final: table !== null};
  }
  if (stop.label) {
    const row = document.querySelector(`[data-label-key="${CSS.escape(stop.label)}"]`);
    if (row) {
      return {element: row, final: true};
    }
  }
  if (stop.target) {
    const element = document.querySelector(stop.target);
    if (element) {
      return {element, final: true};
    }
  }
  if (stop.label || stop.target) {
    return {element: stop.tab ? findTab(stop.tab) : null, final: false};
  }
  if (stop.tab) {
    const tab = findTab(stop.tab);
    return {element: tab, final: tab !== null};
  }
  const table = document.querySelector("[data-tour='table']");
  return {element: table, final: table !== null};
}

/**
 * The tour of the commercial and enterprise features, started from the editions page: each step opens the
 * page the feature is on, switches to its tab and points at its setting, across pages and reloads.
 * The page stays usable under the tour, so the setting can be tried right away.
 */
export function EditionTour() {
  const state = useTourState();
  const location = useLocation();
  const navigate = useNavigate();
  const {account} = useAccount();
  const edition = useEdition();
  const [rect, setRect] = React.useState<DOMRect | null>(null);
  const [searching, setSearching] = React.useState(false);
  const pathCache = React.useRef<Record<string, string>>({});
  const cardRef = React.useRef<HTMLDivElement>(null);

  const feature = state ? getEditionFeature(state.features[state.index]) : undefined;
  const stop = feature?.stop;
  const stepKey = state ? `${state.started}:${state.index}` : "";
  // the step whose page was gone to: a tour restored after a reload stays on the page it is on
  const handledKey = React.useRef(stepKey);

  const goToStep = React.useCallback(async() => {
    if (!stop || !account) {
      return;
    }
    let path = stop.path;
    if (!path) {
      const object = stop.object!;
      if (!pathCache.current[object]) {
        pathCache.current[object] = await resolveObjectPath(object, account);
      }
      path = pathCache.current[object];
    }
    const url = new URL(path, window.location.origin);
    if (url.pathname !== window.location.pathname || (url.search && url.search !== window.location.search)) {
      navigate(url.pathname + url.search);
    }
  }, [stop, account, navigate]);

  // go to the page of the step once, when the step changes
  const hasAccount = !!account;
  React.useEffect(() => {
    if (!stepKey || !hasAccount || stepKey === handledKey.current) {
      return;
    }
    handledKey.current = stepKey;
    goToStep();
    // goToStep is left out on purpose: it changes with the account, which must not move the user again
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stepKey, hasAccount]);

  // follow the setting for as long as the step is open: the page loads it late, re-renders it and scrolls
  React.useEffect(() => {
    setRect(null);
    if (!feature || !stop) {
      setSearching(false);
      return;
    }
    setSearching(true);
    let target: Element | null = null;
    let found = false;
    let scrolled = false;
    const started = Date.now();

    const measure = () => {
      const next = target && target.isConnected ? target.getBoundingClientRect() : null;
      setRect((prev) => (
        prev && next && prev.top === next.top && prev.left === next.left && prev.width === next.width && prev.height === next.height ? prev : next
      ));
    };
    const track = () => {
      // switch to the tab of the setting until it shows up, then leave the tabs to the user
      if (stop.tab && !found) {
        const tab = findTab(stop.tab);
        if (tab) {
          openTab(tab);
        }
      }
      const {element, final} = findTarget(feature);
      target = element;
      if (element && final) {
        found = true;
        if (!scrolled) {
          // a large target, such as a list, is shown from its top
          const tall = element.getBoundingClientRect().height > window.innerHeight * 0.5;
          element.scrollIntoView({block: tall ? "start" : "center", behavior: "smooth"});
          scrolled = true;
        }
      }
      measure();
      if (found || Date.now() - started > FIND_TIMEOUT) {
        setSearching(false);
      }
    };

    track();
    const timer = window.setInterval(track, 200);
    window.addEventListener("resize", measure);
    window.addEventListener("scroll", measure, true);
    return () => {
      window.clearInterval(timer);
      window.removeEventListener("resize", measure);
      window.removeEventListener("scroll", measure, true);
    };
  }, [feature, stop, location.pathname, location.search]);

  if (!state || !feature || Setting.isMobile()) {
    return null;
  }

  const total = state.features.length;
  const isLast = state.index === total - 1;
  const covered = edition === null || isEditionCovered(edition, feature.edition);
  const demoUrl = covered ? undefined : getEditionDemoUrl(feature.edition, `/editions?tour=${feature.id}`);
  const go = (index: number) => writeState({...state, index});

  // under the setting when it fits, above it otherwise, and at the bottom right when neither does
  const cardHeight = cardRef.current?.offsetHeight ?? 200;
  const cardStyle: React.CSSProperties = rect
    ? (() => {
      const left = Math.max(GAP, Math.min(rect.left, window.innerWidth - CARD_WIDTH - GAP));
      if (rect.bottom + PADDING + GAP + cardHeight < window.innerHeight) {
        return {top: rect.bottom + PADDING + GAP, left};
      }
      if (rect.top - PADDING - GAP - cardHeight > 0) {
        return {top: rect.top - PADDING - GAP - cardHeight, left};
      }
      return {bottom: GAP, right: GAP};
    })()
    : {top: "50%", left: "50%", transform: "translate(-50%, -50%)"};

  return (
    <div className="pointer-events-none fixed inset-0 z-[100]" data-edition-tour>
      {rect ? (
        <div
          className="absolute rounded-md ring-2 ring-primary transition-all"
          style={{
            top: rect.top - PADDING,
            left: rect.left - PADDING,
            width: rect.width + PADDING * 2,
            height: rect.height + PADDING * 2,
            boxShadow: "0 0 0 9999px rgba(0, 0, 0, 0.35)",
          }}
        />
      ) : null}

      <div ref={cardRef} className="pointer-events-auto absolute w-[360px] rounded-lg border bg-background p-4 shadow-xl" style={cardStyle}>
        <button
          type="button"
          onClick={stopEditionTour}
          className="absolute right-2 top-2 text-muted-foreground hover:text-foreground"
          aria-label={i18next.t("general:Close")}
        >
          <X className="size-4" />
        </button>

        <span
          className={cn(
            "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-semibold",
            feature.edition === "Enterprise" ? "bg-[#16294f] text-[#e9cf8f]" : "bg-indigo-600/10 text-indigo-700 dark:bg-indigo-400/15 dark:text-indigo-300",
          )}
        >
          {feature.edition === "Enterprise" ? <Gem className="size-3" /> : <Sparkles className="size-3" />}
          {getEditionText(feature.edition)}
        </span>
        <h3 className="mt-2 pr-6 font-semibold">{i18next.t(feature.name)}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{i18next.t(feature.description)}</p>

        {rect && isOnObjectList(feature) ? (
          <p className="mt-2 text-xs text-muted-foreground">{i18next.t("edition:There is none in this organization yet, add one to try the setting")}</p>
        ) : null}

        {!rect && !searching ? (
          <p className="mt-2 text-xs text-muted-foreground">
            {i18next.t("edition:This setting isn't on this page, it may need another object or permission")}{" "}
            <button type="button" className="font-medium text-foreground underline underline-offset-2" onClick={() => goToStep()}>
              {i18next.t("edition:Go to its page")}
            </button>
          </p>
        ) : null}

        {!covered ? (
          <p className="mt-2 rounded-md bg-warning/10 px-2.5 py-1.5 text-xs text-warning">
            {i18next.t("edition:Not in your license - Tip")}
            {demoUrl ? (
              <a href={demoUrl} className="mt-1 flex items-center gap-1 font-medium underline underline-offset-2">
                {i18next.t("edition:Try it on the %s demo site").replace("%s", getEditionText(feature.edition))}
                <ExternalLink className="size-3" />
              </a>
            ) : null}
          </p>
        ) : null}

        <div className="mt-4 flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span className="text-xs tabular-nums text-muted-foreground">{state.index + 1} / {total}</span>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 px-2 text-xs"
              onClick={() => {
                stopEditionTour();
                navigate(`/editions#${feature.id}`);
              }}
            >
              <LayoutList className="size-3.5" />
              {i18next.t("edition:All features")}
            </Button>
          </div>
          <div className="flex gap-2">
            {state.index > 0 ? (
              <Button variant="outline" size="sm" onClick={() => go(state.index - 1)}>
                <ArrowLeft className="size-3.5" />
                {i18next.t("edition:Previous")}
              </Button>
            ) : null}
            <Button size="sm" onClick={() => (isLast ? stopEditionTour() : go(state.index + 1))}>
              {isLast ? i18next.t("edition:Done") : i18next.t("edition:Next")}
              {isLast ? null : <ArrowRight className="size-3.5" />}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
