import * as React from "react";
import i18next from "i18next";
import {X} from "lucide-react";
import {Button} from "@/components/ui/button";
import type {TourStep} from "@/lib/tour-config";

const PADDING = 6;
const CARD_WIDTH = 340;
const GAP = 12;

interface TourProps {
  open: boolean;
  steps: TourStep[];
  /** resolves the element a step highlights; returning null centres the card */
  getTarget: (step: TourStep, index: number) => Element | null;
  onClose: () => void;
  onFinish: () => void;
}

/**
 * The guided tour, replacing antd's `<Tour>`: a dimmed overlay with a hole
 * around the highlighted element and a card next to it. Same steps, same
 * "next page" behaviour — see src/lib/tour-config.ts.
 */
export function Tour({open, steps, getTarget, onClose, onFinish}: TourProps) {
  const [current, setCurrent] = React.useState(0);
  const [rect, setRect] = React.useState<DOMRect | null>(null);

  const step = steps[current];

  React.useEffect(() => {
    if (open) {
      setCurrent(0);
    }
  }, [open]);

  React.useEffect(() => {
    if (!open || !step) {
      return;
    }

    const measure = () => {
      const target = getTarget(step, current);
      if (!target) {
        setRect(null);
        return;
      }
      target.scrollIntoView({block: "nearest", behavior: "smooth"});
      setRect(target.getBoundingClientRect());
    };

    measure();
    // the highlighted element moves while the page scrolls or resizes
    window.addEventListener("resize", measure);
    window.addEventListener("scroll", measure, true);
    return () => {
      window.removeEventListener("resize", measure);
      window.removeEventListener("scroll", measure, true);
    };
  }, [open, step, current, getTarget]);

  if (!open || !step) {
    return null;
  }

  const isLast = current === steps.length - 1;

  // place the card under the hole, or centred when there is nothing to highlight
  const cardStyle: React.CSSProperties = rect
    ? {
      top: Math.min(rect.bottom + GAP, window.innerHeight - 220),
      left: Math.max(GAP, Math.min(rect.left, window.innerWidth - CARD_WIDTH - GAP)),
    }
    : {top: "50%", left: "50%", transform: "translate(-50%, -50%)"};

  return (
    <div className="fixed inset-0 z-[100]">
      {rect ? (
        <div
          className="pointer-events-none absolute rounded-md transition-all"
          style={{
            top: rect.top - PADDING,
            left: rect.left - PADDING,
            width: rect.width + PADDING * 2,
            height: rect.height + PADDING * 2,
            boxShadow: "0 0 0 9999px rgba(0, 0, 0, 0.45)",
          }}
        />
      ) : (
        <div className="absolute inset-0 bg-black/45" />
      )}

      <div
        className="absolute w-[340px] rounded-lg border bg-background p-4 shadow-xl"
        style={cardStyle}
      >
        <button
          type="button"
          onClick={onClose}
          className="absolute right-2 top-2 text-muted-foreground hover:text-foreground"
          aria-label={i18next.t("general:Close")}
        >
          <X className="h-4 w-4" />
        </button>

        {step.coverUrl ? (
          <img src={step.coverUrl} alt={step.title} className="mb-3 max-h-16 object-contain" />
        ) : null}
        <h3 className="pr-6 font-semibold">{step.title}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{step.description}</p>

        <div className="mt-4 flex items-center justify-between">
          <span className="text-xs text-muted-foreground">{current + 1} / {steps.length}</span>
          <div className="flex gap-2">
            {current > 0 ? (
              <Button variant="outline" size="sm" onClick={() => setCurrent(current - 1)}>
                Previous
              </Button>
            ) : null}
            <Button size="sm" onClick={() => (isLast ? onFinish() : setCurrent(current + 1))}>
              {isLast ? (step.nextButtonText ?? "Finish") : "Next"}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
