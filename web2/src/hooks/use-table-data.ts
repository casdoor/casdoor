import * as React from "react";
import * as Setting from "@/lib/setting";
import type {CasdoorListResponse, TableQuery} from "@/components/crud/types";

const defaultQuery: TableQuery = {
  page: 1,
  pageSize: 10,
  sortField: "",
  sortOrder: "",
  searchText: "",
  searchedColumn: "",
};

export interface UseTableDataResult<T> {
  rows: T[] | null;
  total: number;
  loading: boolean;
  /** set when the backend answered with an error */
  errorMessage: string;
  /**
   * The backend refused the read with "Unauthorized operation". The antd
   * BaseListPage renders a 403 page instead of the table in that case.
   */
  denied: boolean;
  query: TableQuery;
  setQuery: (patch: Partial<TableQuery>) => void;
  refresh: () => void;
}

/**
 * Server-side paginated list state, mirroring the semantics of the antd
 * BaseListPage: the Casdoor list APIs take (page, pageSize, field, value,
 * sortField, sortOrder) and answer with {data, data2: total}.
 */
export function useTableData<T = any>(
  fetcher: (query: TableQuery) => Promise<CasdoorListResponse<T>>,
  deps: React.DependencyList = [],
  initial?: Partial<TableQuery>,
): UseTableDataResult<T> {
  const [query, setQueryState] = React.useState<TableQuery>({...defaultQuery, ...initial});
  const [rows, setRows] = React.useState<T[] | null>(null);
  const [total, setTotal] = React.useState(0);
  const [loading, setLoading] = React.useState(false);
  const [errorMessage, setErrorMessage] = React.useState("");
  const [denied, setDenied] = React.useState(false);
  const [nonce, setNonce] = React.useState(0);

  const fetcherRef = React.useRef(fetcher);
  fetcherRef.current = fetcher;

  React.useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setErrorMessage("");
    setDenied(false);
    fetcherRef
      .current(query)
      .then((res) => {
        if (cancelled) {
          return;
        }
        if (res.status === "ok") {
          setRows(res.data ?? []);
          setTotal(Number(res.data2 ?? (res.data ?? []).length));
        } else {
          setRows([]);
          setTotal(0);
          if ((res.data as any) === "Please login first" || (res.msg ?? "").includes("Please login first")) {
            setErrorMessage("Please login first");
          } else if (Setting.isResponseDenied(res)) {
            // the 403 page says it; a toast on top of it would be noise
            setErrorMessage(res.msg ?? "");
            setDenied(true);
          } else {
            setErrorMessage(res.msg ?? "");
            Setting.showMessage("error", res.msg ?? "");
          }
        }
      })
      .catch((e) => {
        if (cancelled) {
          return;
        }
        setRows([]);
        setErrorMessage(e?.message ?? String(e));
        Setting.showMessage("error", e?.message ?? String(e));
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query, nonce, ...deps]);

  const setQuery = React.useCallback((patch: Partial<TableQuery>) => {
    setQueryState((prev) => {
      const next = {...prev, ...patch};
      // any change other than paging resets to the first page
      if (patch.page === undefined && (patch.searchText !== undefined || patch.pageSize !== undefined)) {
        next.page = 1;
      }
      return next;
    });
  }, []);

  const refresh = React.useCallback(() => setNonce((n) => n + 1), []);

  return {rows, total, loading, errorMessage, denied, query, setQuery, refresh};
}
