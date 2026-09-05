import * as React from "react";
import i18next from "i18next";
import {MoreHorizontal} from "lucide-react";
import {Tabs, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger} from "@/components/ui/dropdown-menu";

export type SigninMethod = {value: string; label: string};

const MORE_WIDTH = 28 + 4;
const GAP = 4;
const PADDING = 8;

/**
 * The sign-in method strip: it shows as many methods as fit on one row and
 * moves the rest into a "More" menu.
 */
export function SigninMethodTabs({methods, value, onChange, className}: {
  methods: SigninMethod[];
  value?: string;
  onChange: (value: string) => void;
  className?: string;
}) {
  const stripRef = React.useRef<HTMLDivElement>(null);
  const measureRef = React.useRef<HTMLDivElement>(null);
  const [visibleCount, setVisibleCount] = React.useState(methods.length);

  const labels = methods.map((item) => item.label).join("|");

  React.useLayoutEffect(() => {
    const strip = stripRef.current;
    const measure = measureRef.current;
    if (!strip || !measure) {
      return;
    }

    const fit = () => {
      const available = strip.clientWidth - PADDING;
      const widths = Array.from(measure.children).map((child) => (child as HTMLElement).offsetWidth);
      let used = 0;
      let count = 0;
      for (let i = 0; i < widths.length; i++) {
        const next = used + widths[i] + (i > 0 ? GAP : 0);
        // the "More" button only needs room while methods are left over
        const reserved = i === widths.length - 1 ? 0 : MORE_WIDTH;
        if (next + reserved > available) {
          break;
        }
        used = next;
        count = i + 1;
      }
      setVisibleCount(Math.max(1, count));
    };

    fit();
    const observer = new ResizeObserver(fit);
    observer.observe(strip);
    // the first pass can land on the narrower fallback font
    let cancelled = false;
    document.fonts?.ready.then(() => {
      if (!cancelled) {
        fit();
      }
    });
    return () => {
      cancelled = true;
      observer.disconnect();
    };
  }, [labels]);

  const visible = methods.slice(0, visibleCount);
  // the chosen method always keeps a seat, taking the last one if needed
  if (visible.length < methods.length && !visible.some((item) => item.value === value)) {
    const active = methods.find((item) => item.value === value);
    if (active) {
      visible[visible.length - 1] = active;
    }
  }
  const overflow = methods.filter((item) => !visible.includes(item));

  return (
    <Tabs value={value} onValueChange={onChange} className={className}>
      {/* the track is drawn here so the "More" button can share it with the tabs */}
      <div ref={stripRef} className="relative flex h-9 w-full items-center gap-1 overflow-hidden rounded-lg bg-muted p-1">
        <TabsList className="h-7 min-w-0 flex-1 gap-1 rounded-none bg-transparent p-0 [&>button]:h-7 [&>button]:min-w-max [&>button]:flex-1 [&>button]:basis-0 [&>button]:px-1.5 [&>button]:text-xs">
          {visible.map((item) => (
            <TabsTrigger key={item.value} value={item.value}>
              {item.label}
            </TabsTrigger>
          ))}
        </TabsList>
        {overflow.length > 0 ? (
          <DropdownMenu>
            <DropdownMenuTrigger
              aria-label={i18next.t("general:More")}
              className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none data-[state=open]:bg-background data-[state=open]:text-foreground data-[state=open]:shadow"
            >
              <MoreHorizontal className="h-4 w-4" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {overflow.map((item) => (
                <DropdownMenuItem key={item.value} onSelect={() => onChange(item.value)}>
                  {item.label}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        ) : null}
        {/* an off-screen copy of every label at its natural width, measured above */}
        <div
          ref={measureRef}
          aria-hidden
          className="pointer-events-none invisible absolute left-0 top-0 flex h-0 w-0 overflow-hidden [&>span]:shrink-0"
        >
          {methods.map((item) => (
            <span key={item.value} className="whitespace-nowrap px-1.5 text-xs font-medium">
              {item.label}
            </span>
          ))}
        </div>
      </div>
    </Tabs>
  );
}
