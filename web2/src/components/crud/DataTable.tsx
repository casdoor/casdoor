import * as React from "react";
import i18next from "i18next";
import {Link} from "react-router-dom";
import {ArrowDown, ArrowUp, ChevronLeft, ChevronRight, ChevronsUpDown, Filter, Search, X} from "lucide-react";
import {cn} from "@/lib/utils";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Popover, PopoverContent, PopoverTrigger} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {Skeleton} from "@/components/ui/skeleton";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import type {ColumnDef, ColumnFilterOption, TableQuery} from "@/components/crud/types";

const PAGE_SIZE_OPTIONS = [10, 20, 50, 100];

interface DataTableProps<T> {
  columns: ColumnDef<T>[];
  rows: T[] | null;
  total: number;
  loading?: boolean;
  query: TableQuery;
  onQueryChange: (patch: Partial<TableQuery>) => void;
  rowKey?: (row: T, index: number) => string;
  emptyText?: React.ReactNode;
  className?: string;
}

function Highlight({text, keyword}: {text: string; keyword: string}) {
  if (!keyword || !text) {
    return <>{text}</>;
  }
  const idx = text.toLowerCase().indexOf(keyword.toLowerCase());
  if (idx < 0) {
    return <>{text}</>;
  }
  return (
    <>
      {text.slice(0, idx)}
      <mark className="rounded-sm bg-warning/40 px-0.5 text-foreground">{text.slice(idx, idx + keyword.length)}</mark>
      {text.slice(idx + keyword.length)}
    </>
  );
}

function ColumnSearch({
  column,
  query,
  onQueryChange,
}: {
  column: ColumnDef;
  query: TableQuery;
  onQueryChange: (patch: Partial<TableQuery>) => void;
}) {
  const active = query.searchedColumn === column.dataIndex && query.searchText !== "";
  const [open, setOpen] = React.useState(false);
  const [value, setValue] = React.useState(active ? query.searchText : "");

  React.useEffect(() => {
    if (open) {
      setValue(query.searchedColumn === column.dataIndex ? query.searchText : "");
    }
  }, [open, query.searchText, query.searchedColumn, column.dataIndex]);

  const submit = () => {
    onQueryChange({searchText: value, searchedColumn: column.dataIndex, page: 1});
    setOpen(false);
  };

  /** antd's third button: apply the term but leave the popover open to refine it */
  const filter = () => {
    onQueryChange({searchText: value, searchedColumn: column.dataIndex, page: 1});
  };

  const reset = () => {
    setValue("");
    onQueryChange({searchText: "", searchedColumn: "", page: 1});
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={cn(
            "ml-1 inline-flex h-6 w-6 items-center justify-center rounded hover:bg-accent",
            active && "text-primary",
          )}
          aria-label={i18next.t("general:Search")}
        >
          <Search className="h-3.5 w-3.5" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-64 space-y-2 p-3">
        <Input
          autoFocus
          value={value}
          placeholder={i18next.t("general:Please input your search")}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              submit();
            }
          }}
        />
        <div className="flex gap-2">
          <Button size="sm" className="flex-1" onClick={submit}>
            <Search />
            {i18next.t("general:Search")}
          </Button>
          <Button size="sm" variant="outline" className="flex-1" onClick={reset}>
            <X />
            {i18next.t("forget:Reset")}
          </Button>
          <Button size="sm" variant="link" onClick={filter}>
            {i18next.t("general:Filter")}
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}

/**
 * The header filter menu, antd's `filters` with `filterMultiple: false`. Picking
 * an option filters the list on this column, which the Casdoor list APIs express
 * as the same `field`/`value` pair the per-column search uses.
 */
function ColumnFilter({
  column,
  query,
  onQueryChange,
}: {
  column: ColumnDef;
  query: TableQuery;
  onQueryChange: (patch: Partial<TableQuery>) => void;
}) {
  const [open, setOpen] = React.useState(false);
  const selected = query.searchedColumn === column.dataIndex ? query.searchText : "";
  const active = selected !== "";

  const pick = (value: string) => {
    onQueryChange({searchText: value, searchedColumn: column.dataIndex, page: 1});
    setOpen(false);
  };

  const reset = () => {
    onQueryChange({searchText: "", searchedColumn: "", page: 1});
    setOpen(false);
  };

  const renderOption = (option: ColumnFilterOption) => {
    if (option.children?.length) {
      return (
        <div key={option.value} role="group" aria-label={String(option.value)} className="py-1">
          <div className="px-2 py-1 text-xs font-semibold text-muted-foreground">{option.label}</div>
          {option.children.map((child) => renderOption(child))}
        </div>
      );
    }
    return (
      <button
        key={option.value}
        type="button"
        role="menuitemradio"
        aria-checked={selected === option.value}
        className={cn(
          "flex w-full items-center rounded px-2 py-1.5 text-left text-sm hover:bg-accent",
          selected === option.value && "bg-accent font-medium text-foreground",
        )}
        onClick={() => pick(option.value)}
      >
        {option.label}
      </button>
    );
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={cn(
            "ml-1 inline-flex h-6 w-6 items-center justify-center rounded hover:bg-accent",
            active && "text-primary",
          )}
          aria-label={i18next.t("general:Filter")}
        >
          <Filter className="h-3.5 w-3.5" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-56 p-1" data-column-filter={column.dataIndex}>
        <div role="menu" className="max-h-72 overflow-y-auto">
          {(column.filters ?? []).map((option) => renderOption(option))}
        </div>
        <div className="border-t p-1">
          <Button size="sm" variant="ghost" className="w-full" disabled={!active} onClick={reset}>
            <X />
            {i18next.t("forget:Reset")}
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}

/**
 * antd's `fixed` columns, as sticky cells. Only a leading run of `fixed: "left"`
 * columns and a trailing run of `fixed: "right"` ones can be pinned — a pinned
 * column in the middle would sit on top of its neighbours — and every one of
 * them needs a numeric width so the offsets can be summed. Anything that does
 * not qualify is simply left unpinned.
 */
interface StickyOffset {
  side: "left" | "right";
  offset: number;
  /** the innermost pinned column, which carries the divider shadow */
  edge: boolean;
}

function getStickyOffsets(columns: ColumnDef[]): Map<number, StickyOffset> {
  const result = new Map<number, StickyOffset>();
  const widthOf = (column: ColumnDef) => (typeof column.width === "number" ? column.width : null);

  let offset = 0;
  let lastLeft = -1;
  for (let i = 0; i < columns.length; i++) {
    const width = widthOf(columns[i]);
    if (columns[i].fixed !== "left" || width === null) {
      break;
    }
    result.set(i, {side: "left", offset, edge: false});
    offset += width;
    lastLeft = i;
  }
  if (lastLeft >= 0) {
    result.get(lastLeft)!.edge = true;
  }

  offset = 0;
  let lastRight = -1;
  for (let i = columns.length - 1; i > lastLeft; i--) {
    const width = widthOf(columns[i]);
    if (columns[i].fixed !== "right" || width === null) {
      break;
    }
    result.set(i, {side: "right", offset, edge: false});
    offset += width;
    lastRight = i;
  }
  if (lastRight >= 0) {
    result.get(lastRight)!.edge = true;
  }

  return result;
}

/** the class and inline offset a pinned cell needs; `null` when it is not pinned */
function stickyCell(sticky: StickyOffset | undefined, background: string) {
  if (!sticky) {
    return {className: undefined as string | undefined, style: undefined as React.CSSProperties | undefined};
  }
  return {
    className: cn(
      "sticky z-20",
      background,
      sticky.edge && (sticky.side === "left"
        ? "after:absolute after:inset-y-0 after:-right-px after:w-px after:bg-border"
        : "before:absolute before:inset-y-0 before:-left-px before:w-px before:bg-border"),
    ),
    style: sticky.side === "left" ? {left: sticky.offset} : {right: sticky.offset},
  };
}

export function DataTable<T = any>({
  columns,
  rows,
  total,
  loading,
  query,
  onQueryChange,
  rowKey,
  emptyText,
  className,
}: DataTableProps<T>) {
  const visibleColumns = columns.filter((c) => !c.hidden);
  const stickyOffsets = React.useMemo(() => getStickyOffsets(visibleColumns), [visibleColumns]);
  const pageCount = Math.max(1, Math.ceil(total / query.pageSize));
  const from = total === 0 ? 0 : (query.page - 1) * query.pageSize + 1;
  const to = Math.min(query.page * query.pageSize, total);

  const toggleSort = (column: ColumnDef) => {
    if (!column.sortable) {
      return;
    }
    let order: TableQuery["sortOrder"] = "ascend";
    if (query.sortField === column.dataIndex) {
      order = query.sortOrder === "ascend" ? "descend" : query.sortOrder === "descend" ? "" : "ascend";
    }
    onQueryChange({sortField: order === "" ? "" : column.dataIndex, sortOrder: order});
  };

  return (
    <div className={cn("space-y-3", className)}>
      <div className="rounded-lg border bg-card" data-tour="table">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              {visibleColumns.map((column, columnIndex) => {
                const sorted = query.sortField === column.dataIndex ? query.sortOrder : "";
                const sticky = stickyCell(stickyOffsets.get(columnIndex), "bg-card");
                return (
                  <TableHead
                    key={column.key ?? column.dataIndex}
                    data-column={column.dataIndex}
                    style={{
                      ...(column.width ? {width: column.width, minWidth: column.width} : null),
                      ...sticky.style,
                    }}
                    className={cn(
                      column.align === "center" && "text-center",
                      column.align === "right" && "text-right",
                      sticky.className,
                      column.className,
                    )}
                  >
                    <span className="inline-flex items-center">
                      {column.sortable ? (
                        <button
                          type="button"
                          className="inline-flex items-center gap-1 rounded px-1 py-0.5 hover:text-foreground"
                          onClick={() => toggleSort(column)}
                        >
                          {column.title}
                          {sorted === "ascend" ? (
                            <ArrowUp className="h-3 w-3" />
                          ) : sorted === "descend" ? (
                            <ArrowDown className="h-3 w-3" />
                          ) : (
                            <ChevronsUpDown className="h-3 w-3 opacity-40" />
                          )}
                        </button>
                      ) : (
                        <span className="px-1">{column.title}</span>
                      )}
                      {column.searchable && (
                        <ColumnSearch column={column} query={query} onQueryChange={onQueryChange} />
                      )}
                      {column.filters?.length ? (
                        <ColumnFilter column={column} query={query} onQueryChange={onQueryChange} />
                      ) : null}
                    </span>
                  </TableHead>
                );
              })}
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && rows === null
              ? Array.from({length: 5}).map((_, i) => (
                <TableRow key={`skeleton-${i}`} className="bg-card">
                  {visibleColumns.map((column, columnIndex) => {
                    const sticky = stickyCell(stickyOffsets.get(columnIndex), "bg-inherit");
                    return (
                      <TableCell
                        key={column.key ?? column.dataIndex}
                        style={sticky.style}
                        className={sticky.className}
                      >
                        <Skeleton className="h-4 w-full" />
                      </TableCell>
                    );
                  })}
                </TableRow>
              ))
              : (rows ?? []).length === 0
                ? (
                  <TableRow className="hover:bg-transparent">
                    <TableCell colSpan={visibleColumns.length} className="h-32 text-center text-muted-foreground">
                      {emptyText ?? i18next.t("general:No data")}
                    </TableCell>
                  </TableRow>
                )
                : (rows ?? []).map((row: any, index) => (
                  <TableRow
                    key={rowKey ? rowKey(row, index) : `${row.owner ?? ""}/${row.name ?? index}`}
                    // pinned cells inherit this, so it has to stay fully opaque
                    className="bg-card hover:bg-muted"
                  >
                    {visibleColumns.map((column, columnIndex) => {
                      const value = row[column.dataIndex];
                      // only a typed search highlights; a filter pick would paint whole cells
                      let highlighted: React.ReactNode =
                          column.searchable && query.searchedColumn === column.dataIndex && typeof value === "string" ? (
                            <Highlight text={value} keyword={query.searchText} />
                          ) : (
                            value
                          );
                      const href = column.link?.(value, row);
                      if (href) {
                        highlighted = column.linkExternal ? (
                          <a
                            href={href}
                            target="_blank"
                            rel="noreferrer"
                            className="underline-offset-4 hover:underline"
                          >
                            {highlighted}
                          </a>
                        ) : (
                          <Link to={href} className="underline-offset-4 hover:underline">
                            {highlighted}
                          </Link>
                        );
                      }
                      // the row's own hover/stripe background must show through, so
                      // the pinned cell inherits it rather than painting its own
                      const sticky = stickyCell(stickyOffsets.get(columnIndex), "bg-inherit");
                      return (
                        <TableCell
                          key={column.key ?? column.dataIndex}
                          style={sticky.style}
                          className={cn(
                            column.align === "center" && "text-center",
                            column.align === "right" && "text-right",
                            sticky.className,
                            column.className,
                          )}
                        >
                          {column.render ? column.render(value, row, index) : highlighted}
                        </TableCell>
                      );
                    })}
                  </TableRow>
                ))}
          </TableBody>
        </Table>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground">
        <div>
          {total > 0
            ? `${from}-${to} / ${total}`
            : loading
              ? ""
              : "0 / 0"}
        </div>
        <div className="flex items-center gap-2">
          <Select
            value={String(query.pageSize)}
            onValueChange={(v) => onQueryChange({pageSize: Number(v), page: 1})}
          >
            <SelectTrigger className="h-8 w-[110px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {PAGE_SIZE_OPTIONS.map((size) => (
                <SelectItem key={size} value={String(size)}>
                  {size} / page
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            size="iconSm"
            aria-label="Previous page"
            disabled={query.page <= 1 || loading}
            onClick={() => onQueryChange({page: query.page - 1})}
          >
            <ChevronLeft />
          </Button>
          <span className="min-w-[70px] text-center tabular-nums">
            {query.page} / {pageCount}
          </span>
          <Button
            variant="outline"
            size="iconSm"
            aria-label="Next page"
            disabled={query.page >= pageCount || loading}
            onClick={() => onQueryChange({page: query.page + 1})}
          >
            <ChevronRight />
          </Button>
        </div>
      </div>
    </div>
  );
}
