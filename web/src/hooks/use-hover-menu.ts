import * as React from "react";

/** Opens a Radix dropdown on mouse hover; touch and keyboard keep the default click behaviour. */
export function useHoverMenu(closeDelay = 150) {
  const [open, setOpen] = React.useState(false);
  const timer = React.useRef<ReturnType<typeof setTimeout>>(undefined);

  React.useEffect(() => () => clearTimeout(timer.current), []);

  const onPointerEnter = (e: React.PointerEvent) => {
    if (e.pointerType !== "mouse") {
      return;
    }
    clearTimeout(timer.current);
    setOpen(true);
  };

  const onPointerLeave = (e: React.PointerEvent) => {
    if (e.pointerType !== "mouse") {
      return;
    }
    clearTimeout(timer.current);
    timer.current = setTimeout(() => setOpen(false), closeDelay);
  };

  const onTriggerPointerDown = (e: React.PointerEvent) => {
    if (e.pointerType === "mouse") {
      e.preventDefault();
    }
  };

  return {
    root: {open, onOpenChange: setOpen, modal: false},
    trigger: {onPointerEnter, onPointerLeave, onPointerDown: onTriggerPointerDown},
    content: {onPointerEnter, onPointerLeave},
  };
}
