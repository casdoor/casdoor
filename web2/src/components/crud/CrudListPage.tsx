import * as React from "react";
import i18next from "i18next";
import {Plus, RefreshCw} from "lucide-react";
import {useNavigate} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
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
  /** appended to the built-in Action column; `refresh` re-fetches the current page */
  rowActions?: (record: T, index: number, ctx: {refresh: () => void}) => React.ReactNode;
  /**
   * Name of the Form that customizes this list ("users", "applications", ...).
   * When the organization saved one, it decides which columns show and in which
   * order — see /forms.
   */
  formType?: string;
  /** set to false for read-only lists such as Sessions or Records */
  showActionColumn?: boolean;
  /**
   * The signed-in user may look but not touch: the row action becomes "View"
   * and opens the edit page in its read-only mode, and Add and Delete are
   * disabled. The antd list pages do this for anyone who is not a local admin.
   */
  readOnly?: boolean;
  /**
   * Blocks Delete for one row. A string is shown as a tooltip explaining why
   * (the group list uses it for a group that still has subgroups); `true` just
   * disables the button, the way antd does for the built-in objects.
   */
  deleteDisabled?: (record: T) => string | boolean | undefined;
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
  readOnly = false,
  deleteDisabled,
}: CrudListPageProps<T>) {
  const navigate = useNavigate();
  const {rows, total, loading, denied, query, setQuery, refresh} = useTableData<T>(fetch, deps, initialQuery);
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

  const renderDelete = (record: T) => {
    const blockedReason = deleteDisabled?.(record);
    const button = (
      <ConfirmButton
        variant="destructiveGhost"
        size="sm"
        disabled={readOnly || Boolean(blockedReason)}
        description={`${record.name ?? ""}`}
        onConfirm={() => handleDelete(record)}
      >
        {i18next.t("general:Delete")}
      </ConfirmButton>
    );
    if (typeof blockedReason !== "string" || blockedReason === "") {
      return button;
    }
    // a disabled button swallows pointer events, so the tooltip hangs off a wrapper
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <span className="inline-flex">{button}</span>
        </TooltipTrigger>
        <TooltipContent className="max-w-sm">{blockedReason}</TooltipContent>
      </Tooltip>
    );
  };

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
        align: "right",
        // antd pins it so the row's actions stay reachable on a wide table
        fixed: "right",
        render: (_: any, record: T, index: number) => (
          <div className="flex flex-wrap items-center justify-end gap-1">
            {rowActions?.(record, index, {refresh})}
            {editUrl ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  navigate(editUrl(record), readOnly ? {state: {mode: "view"}} : undefined)
                }
              >
                {i18next.t(readOnly ? "general:View" : "general:Edit")}
              </Button>
            ) : null}
            {remove ? renderDelete(record) : null}
          </div>
        ),
      },
    ];
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [columns, formItems, editUrl, remove, rowActions, showActionColumn, readOnly, deleteDisabled, rows, query.page, refresh]);

  if (denied) {
    return <UnauthorizedPage />;
  }

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
              <Button id="add-button" onClick={handleAdd} loading={adding} disabled={readOnly}>
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
