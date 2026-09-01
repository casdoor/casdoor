import * as React from "react";
import i18next from "i18next";
import {SlidersHorizontal} from "lucide-react";
import {Button} from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type {ColumnDef} from "@/components/crud/types";

export const columnKey = (column: ColumnDef) => column.key ?? column.dataIndex;

/**
 * How many optional columns a list shows before the rest move into the column
 * menu. The Casdoor list APIs return every field of the object and the antd
 * tables rendered all of them, which put /users at 23 columns and 3100px wide in
 * a 1130px card. The column arrays are written most-important-first, so keeping
 * the head of each one and parking the tail behind the menu fits the common case
 * on screen without a page having to name its columns twice.
 */
const DEFAULT_VISIBLE_COLUMNS = 6;

/** the row's identity and its actions are what the table is for; they never hide */
const isAlwaysVisible = (column: ColumnDef) => column.key === "op" || column.fixed === "left";

/**
 * Which columns this reader has chosen to see. `null` means they have not chosen,
 * so each column's own `defaultHidden` decides — that way tuning the defaults in
 * code still reaches everyone who never opened the menu.
 */
export function useColumnVisibility(columns: ColumnDef[], tableId?: string) {
  const storageKey = tableId ? `casdoorColumns:${tableId}` : "";

  const read = React.useCallback(() => {
    if (!storageKey) {
      return null;
    }
    try {
      const raw = window.localStorage.getItem(storageKey);
      return raw ? (JSON.parse(raw) as string[]) : null;
    } catch {
      return null;
    }
  }, [storageKey]);

  const [override, setOverride] = React.useState<string[] | null>(read);
  React.useEffect(() => setOverride(read()), [read]);

  const hidden = React.useMemo(() => {
    if (override) {
      return new Set(override.filter((key) => !columns.some((c) => columnKey(c) === key && isAlwaysVisible(c))));
    }
    const set = new Set<string>();
    let shown = 0;
    columns.forEach((column) => {
      if (isAlwaysVisible(column)) {
        return;
      }
      // `defaultHidden: false` is a column saying it earns its place past the cap
      const beyondCap = column.defaultHidden === undefined && shown >= DEFAULT_VISIBLE_COLUMNS;
      if (column.defaultHidden === true || beyondCap) {
        set.add(columnKey(column));
      } else {
        shown++;
      }
    });
    return set;
  }, [override, columns]);

  const write = (next: Set<string>) => {
    const list = Array.from(next);
    setOverride(list);
    if (storageKey) {
      try {
        window.localStorage.setItem(storageKey, JSON.stringify(list));
      } catch {
        // a full or blocked store just means the choice lasts this page view
      }
    }
  };

  return {
    hidden,
    customized: override !== null,
    toggle: (key: string) => {
      const next = new Set(hidden);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      write(next);
    },
    showAll: () => write(new Set<string>()),
    reset: () => {
      setOverride(null);
      if (storageKey) {
        try {
          window.localStorage.removeItem(storageKey);
        } catch {
          // nothing to undo
        }
      }
    },
  };
}

/** The column menu: what is on the table, and what is one click from being on it. */
export function ColumnMenu({
  columns,
  visibility,
}: {
  columns: ColumnDef[];
  visibility: ReturnType<typeof useColumnVisibility>;
}) {
  const toggleable = columns.filter((c) => !c.hidden && !isAlwaysVisible(c));
  if (toggleable.length === 0) {
    return null;
  }
  const shown = toggleable.length - toggleable.filter((c) => visibility.hidden.has(columnKey(c))).length;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">
          <SlidersHorizontal />
          {i18next.t("general:Columns")}
          <span className="tabular-nums text-muted-foreground">
            {shown}/{toggleable.length}
          </span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="max-h-[70vh] w-56 overflow-y-auto">
        <DropdownMenuLabel>{i18next.t("general:Columns")}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {toggleable.map((column) => {
          const key = columnKey(column);
          return (
            <DropdownMenuCheckboxItem
              key={key}
              checked={!visibility.hidden.has(key)}
              // the menu stays open so several columns can be picked in one go
              onSelect={(e) => e.preventDefault()}
              onCheckedChange={() => visibility.toggle(key)}
            >
              <span className="truncate">{column.title}</span>
            </DropdownMenuCheckboxItem>
          );
        })}
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={() => visibility.showAll()}>{i18next.t("general:Select all")}</DropdownMenuItem>
        <DropdownMenuItem disabled={!visibility.customized} onSelect={() => visibility.reset()}>
          {i18next.t("forget:Reset")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
