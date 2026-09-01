import * as React from "react";
import i18next from "i18next";
import {HelpCircle} from "lucide-react";
import {useLocation, useNavigate} from "react-router-dom";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {Tour} from "@/components/common/Tour";
import * as TourConfig from "@/lib/tour-config";
import * as Setting from "@/lib/setting";

/** Re-reads the flag whenever any part of the app flips it. */
function useTourVisible() {
  const [visible, setVisible] = React.useState(TourConfig.getTourVisible());

  React.useEffect(() => {
    const onChange = () => setVisible(TourConfig.getTourVisible());
    window.addEventListener("storageTourChanged", onChange);
    return () => window.removeEventListener("storageTourChanged", onChange);
  }, []);

  return visible;
}

/** The "?" button in the header, ported from web/src/common/OpenTour.js. */
export function OpenTour() {
  const location = useLocation();
  const enabled = TourConfig.canTour(location.pathname);

  if (Setting.isMobile()) {
    return null;
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          type="button"
          disabled={!enabled}
          onClick={() => TourConfig.setIsTourVisible(true)}
          className="flex h-9 w-9 items-center justify-center rounded-md text-muted-foreground hover:bg-accent disabled:cursor-not-allowed disabled:opacity-40"
          aria-label={i18next.t("general:Click to open tour")}
        >
          <HelpCircle className="h-5 w-5" />
        </button>
      </TooltipTrigger>
      <TooltipContent>{i18next.t("general:Click to open tour")}</TooltipContent>
    </Tooltip>
  );
}

/**
 * Drives the tour across the console, the way BaseListPage did in the antd
 * frontend: the steps come from the current path, the first one highlights the
 * table and the last one moves on to the next page in TourUrlList.
 */
export function ConsoleTour() {
  const location = useLocation();
  const navigate = useNavigate();
  const visible = useTourVisible();

  const steps = React.useMemo(() => {
    const nextPathName = TourConfig.getNextUrl(location.pathname);
    return TourConfig.getSteps(location.pathname).map((step, index, all) =>
      index === all.length - 1
        ? {...step, nextButtonText: TourConfig.getNextButtonChild(nextPathName)}
        : step,
    );
  }, [location.pathname]);

  const getTarget = React.useCallback((step: TourConfig.TourStep, index: number) => {
    if (step.id) {
      return document.getElementById(step.id);
    }
    // the antd tour anchors the first step on the page's table
    return index === 0 ? document.querySelector("[data-tour='table']") : null;
  }, []);

  if (Setting.isMobile() || steps.length === 0) {
    return null;
  }

  return (
    <Tour
      open={visible}
      steps={steps}
      getTarget={getTarget}
      onClose={() => TourConfig.setIsTourVisible(false)}
      onFinish={() => {
        const nextPathName = TourConfig.getNextUrl(location.pathname);
        if (nextPathName === "") {
          TourConfig.setIsTourVisible(false);
          return;
        }
        navigate(`/${nextPathName}`);
        TourConfig.setIsTourVisible(true);
      }}
    />
  );
}
