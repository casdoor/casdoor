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
  className?: string;
  /** hide the column entirely (used by the Forms feature) */
  hidden?: boolean;
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
