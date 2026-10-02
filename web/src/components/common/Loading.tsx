import * as React from "react";
import {Loader2} from "lucide-react";
import {cn} from "@/lib/utils";

function alignSpinToDocumentTimeline(element: Element | null) {
  element?.getAnimations?.().forEach((animation) => {
    animation.startTime = 0;
  });
}

export function Loading({className, size = 24}: {className?: string; size?: number}) {
  const spinnerRef = React.useRef<SVGSVGElement>(null);

  React.useLayoutEffect(() => alignSpinToDocumentTimeline(spinnerRef.current), []);

  return (
    <div className={cn("flex w-full items-center justify-center py-16", className)}>
      <Loader2
        ref={spinnerRef}
        className="animate-spin text-muted-foreground"
        style={{width: size, height: size}}
      />
    </div>
  );
}

export default Loading;
