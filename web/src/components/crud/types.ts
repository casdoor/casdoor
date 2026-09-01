import type * as React from "react";

export type SortOrder = "" | "ascend" | "descend";

export interface TableQuery {
  page: number;
  pageSize: number;
  sortField: string;
  sortOrder: SortOrder;
  /** value typed into the per-column search box, or picked from its filter menu */
  searchText: string;
  /** dataIndex of the column the search applies to */
  searchedColumn: string;
}

/**
 * One entry of a column's filter menu. The antd tables declare these as
 * `filters: [{text, value}]` and every list page turns the picked value into the
 * `field`/`value` pair of the same list API, which is exactly what the per-column
 * search sends — so a filter here is just a canned search.
 */
export interface ColumnFilterOption {
  label: React.ReactNode;
  value: string;
  /** antd's two-level menu, used by the provider type filter */
  children?: ColumnFilterOption[];
}

export interface ColumnDef<T = any> {
  /** unique key, defaults to dataIndex */
  key?: string;
  /** field of the record, also the `field` sent to the Casdoor list API */
  dataIndex: string;
  title: React.ReactNode;
  width?: number | string;
  sortable?: boolean;
  searchable?: boolean;
  /** renders a filter menu in the header cell; picking one filters on this column */
  filters?: ColumnFilterOption[];
  align?: "left" | "center" | "right";
  /**
   * Pins the column while the table scrolls sideways, antd's `fixed`. Only a
   * leading run of "left" columns and a trailing run of "right" ones can be
   * pinned, and each needs a numeric `width` so the offsets can be added up —
   * `DataTable` quietly ignores the flag otherwise.
   */
  fixed?: "left" | "right";
  className?: string;
  /** hide the column entirely (used by the Forms feature) */
  hidden?: boolean;
  /**
   * Overrides where this column sits relative to `DataTable`'s default cap on how
   * many optional columns a list opens with. `true` parks it in the column menu
   * however early it is declared; `false` keeps it on the table however late.
   * Leave it unset to let the cap decide.
   */
  defaultHidden?: boolean;
  /**
   * Wraps the cell content in a link, keeping the search highlight the plain
   * cell has. Return undefined to leave the cell unlinked. Ignored when `render`
   * is set, which owns the whole cell.
   */
  link?: (value: any, record: T) => string | undefined;
  /** the link leaves Casdoor, so render an `<a target="_blank">` instead of a router Link */
  linkExternal?: boolean;
  render?: (value: any, record: T, index: number) => React.ReactNode;
}

export interface CasdoorListResponse<T = any> {
  status: "ok" | "error";
  msg?: string;
  data: T[];
  data2?: number | string;
}

/**
 * One button in a list row's action column.
 *
 * Every list page used to spell these out as JSX, which meant each one repeated
 * the same variant, size, confirmation dialog and disabled-with-a-reason tooltip
 * by hand. They are descriptors now, so `RowActions` renders the lot uniformly.
 */
export interface RowAction {
  key?: string;
  label: React.ReactNode;
  /** why a disabled action is unavailable, shown as a tooltip */
  description?: React.ReactNode;
  /** an internal route; the action becomes a link instead of a button */
  href?: string;
  onSelect?: () => void | Promise<void>;
  disabled?: boolean;
  loading?: boolean;
  destructive?: boolean;
  /** ask before running. `true` uses the default "Sure to delete?" wording. */
  confirm?: boolean | {title?: React.ReactNode; description?: React.ReactNode; confirmText?: React.ReactNode};
}
