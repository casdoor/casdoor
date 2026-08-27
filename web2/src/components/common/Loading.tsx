import {Loader2} from "lucide-react";
import {cn} from "@/lib/utils";

export function Loading({className, size = 24}: {className?: string; size?: number}) {
  return (
    <div className={cn("flex w-full items-center justify-center py-16", className)}>
      <Loader2 className="animate-spin text-muted-foreground" style={{width: size, height: size}} />
    </div>
  );
}

export default Loading;
