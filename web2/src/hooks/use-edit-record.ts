import * as React from "react";
import {useLocation} from "react-router-dom";
import * as Setting from "@/lib/setting";
import type {CasdoorResponse, EditMode} from "@/lib/crud";

interface UseEditRecordOptions<T> {
  /** loads the record in "edit" mode */
  fetch: () => Promise<CasdoorResponse<T>>;
  /** applied to the record right after it is loaded (edit) or handed over (add) */
  transform?: (record: T) => T;
  deps?: React.DependencyList;
}

export interface UseEditRecordResult<T> {
  record: T | null;
  setRecord: React.Dispatch<React.SetStateAction<T | null>>;
  updateField: (field: string, value: any) => void;
  /** patches several fields in one render, for fields that must move together */
  updateFields: (patch: Record<string, any>) => void;
  loading: boolean;
  notFound: boolean;
  mode: EditMode;
  setMode: (mode: EditMode) => void;
  reload: () => void;
}

/**
 * Loads the object an edit page works on. In "add" mode the list page hands the
 * freshly built object over through the router state (as the antd frontend did
 * with `history.push({mode: "add", organization})`), so no request is made.
 */
export function useEditRecord<T extends Record<string, any>>({
  fetch,
  transform,
  deps = [],
}: UseEditRecordOptions<T>): UseEditRecordResult<T> {
  const location = useLocation();
  const state = (location.state ?? {}) as {mode?: EditMode; record?: T};

  const [mode, setMode] = React.useState<EditMode>(state.mode === "add" ? "add" : "edit");
  const [record, setRecord] = React.useState<T | null>(() =>
    state.mode === "add" && state.record ? (transform ? transform(state.record) : state.record) : null,
  );
  const [loading, setLoading] = React.useState(mode !== "add");
  const [notFound, setNotFound] = React.useState(false);
  const [nonce, setNonce] = React.useState(0);

  const fetchRef = React.useRef(fetch);
  fetchRef.current = fetch;
  const transformRef = React.useRef(transform);
  transformRef.current = transform;

  React.useEffect(() => {
    if (mode === "add" && record !== null) {
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    fetchRef
      .current()
      .then((res) => {
        if (cancelled) {
          return;
        }
        if (res.status === "ok") {
          if (res.data === null || res.data === undefined) {
            setNotFound(true);
            setRecord(null);
          } else {
            const next = transformRef.current ? transformRef.current(res.data as T) : (res.data as T);
            setRecord(next);
          }
        } else {
          Setting.showMessage("error", res.msg ?? "");
          setNotFound(true);
        }
      })
      .catch((e) => {
        if (!cancelled) {
          Setting.showMessage("error", e?.message ?? String(e));
          setNotFound(true);
        }
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
  }, [nonce, ...deps]);

  const updateField = React.useCallback((field: string, value: any) => {
    setRecord((prev) => (prev === null ? prev : ({...prev, [field]: value} as T)));
  }, []);

  const updateFields = React.useCallback((patch: Record<string, any>) => {
    setRecord((prev) => (prev === null ? prev : ({...prev, ...patch} as T)));
  }, []);

  return {
    record,
    setRecord,
    updateField,
    updateFields,
    loading,
    notFound,
    mode,
    setMode,
    reload: () => setNonce((n) => n + 1),
  };
}
