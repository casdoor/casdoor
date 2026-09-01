import * as React from "react";
import i18next from "i18next";
import {Github, Network, SquareArrowOutUpRight} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Sheet, SheetContent, SheetHeader, SheetTitle} from "@/components/ui/sheet";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import * as Conf from "@/Conf";

/**
 * The AI assistant drawer, an iframe of `Conf.AiAssistantUrl`. Hidden when the
 * deployment blanks that config value, same as the antd frontend.
 */
export function AiAssistant() {
  const [open, setOpen] = React.useState(false);
  const url = Conf.AiAssistantUrl?.trim();

  if (!url) {
    return null;
  }

  return (
    <>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="ghost" size="iconSm" onClick={() => setOpen(true)} aria-label="AI Assistant">
            <Network />
          </Button>
        </TooltipTrigger>
        <TooltipContent>{i18next.t("general:Click to open AI assistant")}</TooltipContent>
      </Tooltip>

      {/* not modal: the assistant is meant to stay open while you keep working */}
      <Sheet open={open} onOpenChange={setOpen} modal={false}>
        <SheetContent side="right" className="flex w-full flex-col gap-0 p-0 sm:max-w-[500px]">
          <SheetHeader className="flex-row items-center gap-3 space-y-0 border-b p-4">
            <SheetTitle className="flex-1 text-base">AI Assistant</SheetTitle>
            <a
              href="https://github.com/casibase/casibase"
              target="_blank"
              rel="noreferrer"
              className="text-muted-foreground hover:text-foreground"
              aria-label="GitHub"
            >
              <Github className="h-5 w-5" />
            </a>
            <a
              href={url}
              target="_blank"
              rel="noreferrer"
              className="mr-6 text-muted-foreground hover:text-foreground"
              aria-label="Open in a new tab"
            >
              <SquareArrowOutUpRight className="h-5 w-5" />
            </a>
          </SheetHeader>
          <iframe
            id="iframeHelper"
            title="iframeHelper"
            src={`${url}/?isRaw=1`}
            className="h-full w-full flex-1 border-0"
            scrolling="no"
          />
        </SheetContent>
      </Sheet>
    </>
  );
}
