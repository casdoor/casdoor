import * as React from "react";
import i18next from "i18next";
import {Plus, RefreshCw} from "lucide-react";
import {useNavigate} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {DataTable} from "@/components/crud/DataTable";
import {PageHeader} from "@/components/crud/PageHeader";
import type {ColumnDef, CasdoorListResponse, TableQuery} from "@/components/crud/types";
import {useFormItems} from "@/hooks/use-form-items";
import {useTableData} from "@/hooks/use-table-data";
import {submitAdd, submitDelete} from "@/lib/crud";
import * as Setting from "@/lib/setting";

export interface CrudListPageProps<T extends Record<string, any>> {
  title: React.ReactNode;
  description?: React.ReactNode;
  columns: ColumnDef<T>[];
  fetch: (query: TableQuery) => Promise<CasdoorListResponse<T>>;
  /** re-fetch when any of these change */
  deps?: React.DependencyList;
  /** builds the default object for a new row; omit to hide the Add button */
  newRecord?: () => T;
  /**
   * When set, the new object is POSTed right away and the user is then sent to
   * its edit page. When omitted, the object is handed to the edit page through
   * the router state and only saved when the user presses Save.
   */
  add?: (record: T) => Promise<any>;
  remove?: (record: T) => Promise<any>;
  /** where the "Add" flow and the name links point at */
  editUrl?: (record: T) => string;
  addButtonLabel?: React.ReactNode;
  /** extra buttons next to Refresh/Add; a function form gets `refresh` to re-fetch the list */
  toolbar?: React.ReactNode | ((ctx: {refresh: () => void}) => React.ReactNode);
  rowKey?: (row: T, index: number) => string;
  initialQuery?: Partial<TableQuery>;
  /** appended to the built-in Action column */
  rowActions?: (record: T, index: number) => React.ReactNode;
  /**
   * Name of the Form that customizes this list ("users", "applications", ...).
   * When the organization saved one, it decides which columns show and in which
   * order — see /forms.
   */
  formType?: string;
  /** set to false for read-only lists such as Sessions or Records */
  showActionColumn?: boolean;
  actionColumnWidth?: number | string;
}

export function CrudListPage<T extends Record<string, any>>({
  title,
  description,
  columns,
  fetch,
  deps = [],
  newRecord,
  add,
  remove,
  editUrl,
  addButtonLabel,
  toolbar,
  rowKey,
  initialQuery,
  rowActions,
  formType,
  showActionColumn = true,
  actionColumnWidth = 180,
}: CrudListPageProps<T>) {
  const navigate = useNavigate();
  const {rows, total, loading, query, setQuery, refresh} = useTableData<T>(fetch, deps, initialQuery);
  const [adding, setAdding] = React.useState(false);

  const handleAdd = async() => {
    if (!newRecord) {
      return;
    }
    const record = newRecord();
    if (add) {
      setAdding(true);
      const ok = await submitAdd({record, add});
      setAdding(false);
      if (ok) {
        if (editUrl) {
          navigate(editUrl(record));
        } else {
          refresh();
        }
      }
      return;
    }
    if (editUrl) {
      navigate(editUrl(record), {state: {mode: "add", record}});
    }
  };

  const handleDelete = async(record: T) => {
    if (!remove) {
      return;
    }
    await submitDelete({
      record,
      remove,
      onDeleted: () => {
        // step back a page when the last row of the page was removed
        if (query.page > 1 && (rows ?? []).length === 1) {
          setQuery({page: query.page - 1});
        } else {
          refresh();
        }
      },
    });
  };

  const formItems = useFormItems(formType);

  const allColumns = React.useMemo<ColumnDef<T>[]>(() => {
    const visibleColumns = Setting.filterTableColumns(columns, formItems) as ColumnDef<T>[];
    if (!showActionColumn) {
      return visibleColumns;
    }
    return [
      ...visibleColumns,
      {
        key: "op",
        dataIndex: "op",
        title: i18next.t("general:Action"),
        width: actionColumnWidth,
        render: (_: any, record: T, index: number) => (
          <div className="flex flex-wrap items-center gap-1">
            {rowActions?.(record, index)}
            {editUrl ? (
              <Button variant="outline" size="sm" onClick={() => navigate(editUrl(record))}>
                {i18next.t("general:Edit")}
              </Button>
            ) : null}
            {remove ? (
              <ConfirmButton
                variant="destructive"
                size="sm"
                description={`${record.name ?? ""}`}
                onConfirm={() => handleDelete(record)}
              >
                {i18next.t("general:Delete")}
              </ConfirmButton>
            ) : null}
          </div>
        ),
      },
    ];
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [columns, formItems, editUrl, remove, rowActions, showActionColumn, rows, query.page]);

  return (
    <div className="space-y-4">
      <PageHeader
        title={title}
        description={description}
        actions={
          <>
            {typeof toolbar === "function" ? toolbar({refresh}) : toolbar}
            <Button variant="outline" size="iconSm" onClick={refresh} aria-label="Refresh" disabled={loading}>
              <RefreshCw className={loading ? "animate-spin" : undefined} />
            </Button>
            {newRecord ? (
              <Button onClick={handleAdd} loading={adding}>
                <Plus />
                {addButtonLabel ?? i18next.t("general:Add")}
              </Button>
            ) : null}
          </>
        }
      />
      <DataTable
        columns={allColumns}
        rows={rows}
        total={total}
        loading={loading}
        query={query}
        onQueryChange={setQuery}
        rowKey={rowKey}
      />
    </div>
  );
}
