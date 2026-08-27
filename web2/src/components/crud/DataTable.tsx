import * as React from "react";
import i18next from "i18next";
import {ArrowDown, ArrowUp, ChevronLeft, ChevronRight, ChevronsUpDown, Search, X} from "lucide-react";
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
import type {ColumnDef, TableQuery} from "@/components/crud/types";

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
        </div>
      </PopoverContent>
    </Popover>
  );
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
      <div className="rounded-lg border bg-card">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              {visibleColumns.map((column) => {
                const sorted = query.sortField === column.dataIndex ? query.sortOrder : "";
                return (
                  <TableHead
                    key={column.key ?? column.dataIndex}
                    style={column.width ? {width: column.width, minWidth: column.width} : undefined}
                    className={cn(
                      column.align === "center" && "text-center",
                      column.align === "right" && "text-right",
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
                    </span>
                  </TableHead>
                );
              })}
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && rows === null
              ? Array.from({length: 5}).map((_, i) => (
                <TableRow key={`skeleton-${i}`}>
                  {visibleColumns.map((column) => (
                    <TableCell key={column.key ?? column.dataIndex}>
                      <Skeleton className="h-4 w-full" />
                    </TableCell>
                  ))}
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
                  <TableRow key={rowKey ? rowKey(row, index) : `${row.owner ?? ""}/${row.name ?? index}`}>
                    {visibleColumns.map((column) => {
                      const value = row[column.dataIndex];
                      const highlighted =
                          query.searchedColumn === column.dataIndex && typeof value === "string" ? (
                            <Highlight text={value} keyword={query.searchText} />
                          ) : (
                            value
                          );
                      return (
                        <TableCell
                          key={column.key ?? column.dataIndex}
                          className={cn(
                            column.align === "center" && "text-center",
                            column.align === "right" && "text-right",
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
