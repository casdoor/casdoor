import type * as React from "react";

export type SortOrder = "" | "ascend" | "descend";

export interface TableQuery {
  page: number;
  pageSize: number;
  sortField: string;
  sortOrder: SortOrder;
  /** value typed into the per-column search box */
  searchText: string;
  /** dataIndex of the column the search applies to */
  searchedColumn: string;
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
  align?: "left" | "center" | "right";
  className?: string;
  /** hide the column entirely (used by the Forms feature) */
  hidden?: boolean;
  render?: (value: any, record: T, index: number) => React.ReactNode;
}

export interface CasdoorListResponse<T = any> {
  status: "ok" | "error";
  msg?: string;
  data: T[];
  data2?: number | string;
}
