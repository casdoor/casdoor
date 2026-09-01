import * as React from "react";
import {Link} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import type {RowAction} from "@/components/crud/types";

function RowActionButton({action}: {action: RowAction}) {
  // a wall of solid red down the page is noise, so a destructive row action only
  // commits to the colour under the pointer
  const variant = action.destructive ? "destructiveGhost" : "outline";
  const confirm = typeof action.confirm === "object" ? action.confirm : undefined;

  let button: React.ReactNode;
  if (action.confirm) {
    button = (
      <ConfirmButton
        variant={variant}
        size="sm"
        disabled={action.disabled}
        loading={action.loading}
        title={confirm?.title}
        description={confirm?.description}
        confirmText={confirm?.confirmText}
        destructive={action.destructive}
        onConfirm={() => action.onSelect?.()}
      >
        {action.label}
      </ConfirmButton>
    );
  } else if (action.href && !action.disabled) {
    button = (
      <Button variant={variant} size="sm" asChild>
        <Link to={action.href}>{action.label}</Link>
      </Button>
    );
  } else {
    button = (
      <Button
        variant={variant}
        size="sm"
        disabled={action.disabled}
        loading={action.loading}
        onClick={() => action.onSelect?.()}
      >
        {action.label}
      </Button>
    );
  }

  if (!action.description) {
    return button;
  }
  // a disabled button swallows pointer events, so the tooltip hangs off a wrapper
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="inline-flex">{button}</span>
      </TooltipTrigger>
      <TooltipContent className="max-w-sm">{action.description}</TooltipContent>
    </Tooltip>
  );
}

/** The buttons of a list row's action column, laid out the way every list shares. */
export function RowActions({actions}: {actions: (RowAction | null | false | undefined)[]}) {
  const items = actions.filter(Boolean) as RowAction[];
  if (items.length === 0) {
    return null;
  }
  return (
    <div className="flex flex-wrap items-center justify-end gap-1">
      {items.map((action, index) => (
        <RowActionButton key={action.key ?? index} action={action} />
      ))}
    </div>
  );
}

