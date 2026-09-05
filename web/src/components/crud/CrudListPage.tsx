import * as React from "react";
import i18next from "i18next";
import {Plus, RefreshCw} from "lucide-react";
import {useLocation, useNavigate} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {ColumnMenu, columnKey, useColumnVisibility} from "@/components/crud/ColumnMenu";
import {DataTable} from "@/components/crud/DataTable";
import {PageHeader} from "@/components/crud/PageHeader";
import {RowActions} from "@/components/crud/RowActions";
import type {ColumnDef, CasdoorListResponse, RowAction, TableQuery} from "@/components/crud/types";
import {useFormItems} from "@/hooks/use-form-items";
import {useTableData} from "@/hooks/use-table-data";
import {submitAdd, submitDelete, submitDeleteMany} from "@/lib/crud";
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
  /**
   * Extra entries for the row's actions menu, between Edit and Delete.
   * `refresh` re-fetches the current page.
   */
  rowActions?: (record: T, index: number, ctx: {refresh: () => void}) => (RowAction | null | false | undefined)[];
  /**
   * Name of the Form that customizes this list ("users", "applications", ...).
   * When the organization saved one, it decides which columns show and in which
   * order — see /forms.
   */
  formType?: string;
  /**
   * Renders the list with these items instead of the saved Form's, which is how
   * the form editor previews the columns it is editing.
   */
  formItems?: any[];
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
  /**
   * Namespaces the reader's column choices. Defaults to the route, which is one
   * list page per path.
   */
  tableId?: string;
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
  formItems: formItemsProp,
  showActionColumn = true,
  actionColumnWidth = 180,
  readOnly = false,
  deleteDisabled,
  tableId,
}: CrudListPageProps<T>) {
  const navigate = useNavigate();
  const location = useLocation();
  const {rows, total, loading, denied, query, setQuery, refresh} = useTableData<T>(fetch, deps, initialQuery);
  const [adding, setAdding] = React.useState(false);
  const [selected, setSelected] = React.useState<Set<string>>(new Set());
  const [deletingMany, setDeletingMany] = React.useState(false);

  const keyOf = React.useCallback(
    (row: any, index: number) => (rowKey ? rowKey(row, index) : `${row.owner ?? ""}/${row.name ?? index}`),
    [rowKey],
  );

  // the selection is of rows, and a page change or a re-fetch replaces them
  React.useEffect(() => setSelected(new Set()), [query.page, query.pageSize, query.searchText, query.searchedColumn]);

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

  const savedFormItems = useFormItems(formItemsProp ? undefined : formType);
  const formItems = formItemsProp ?? savedFormItems;

  const deleteAction = (record: T): RowAction => {
    const blockedReason = deleteDisabled?.(record);
    return {
      key: "delete",
      label: i18next.t("general:Delete"),
      description: typeof blockedReason === "string" && blockedReason !== "" ? blockedReason : undefined,
      destructive: true,
      disabled: readOnly || Boolean(blockedReason),
      confirm: {description: `${record.name ?? ""}`},
      onSelect: () => handleDelete(record),
    };
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
          <RowActions
            actions={[
              editUrl
                ? {
                  key: "edit",
                  label: i18next.t(readOnly ? "general:View" : "general:Edit"),
                  onSelect: () => navigate(editUrl(record), readOnly ? {state: {mode: "view"}} : undefined),
                }
                : null,
              ...(rowActions?.(record, index, {refresh}) ?? []),
              remove ? deleteAction(record) : null,
            ]}
          />
        ),
      },
    ];
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [columns, formItems, editUrl, remove, rowActions, showActionColumn, readOnly, deleteDisabled, rows, query.page, refresh]);

  const deleteSelected = async() => {
    const records = (rows ?? []).filter((row, index) => selected.has(keyOf(row, index)) && !deleteDisabled?.(row));
    setDeletingMany(true);
    await submitDeleteMany({records, remove: remove!, onDeleted: () => {
      setSelected(new Set());
      refresh();
    }});
    setDeletingMany(false);
  };

  // a read-only viewer has nothing to do with a selection, and neither does a
  // list that cannot delete
  const selectable = Boolean(remove) && !readOnly && showActionColumn;

  const visibility = useColumnVisibility(allColumns, tableId ?? location.pathname);
  const shownColumns = React.useMemo(
    () => allColumns.filter((column) => !visibility.hidden.has(columnKey(column))),
    [allColumns, visibility.hidden],
  );

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
            <ColumnMenu columns={allColumns} visibility={visibility} />
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
        columns={shownColumns}
        rows={rows}
        total={total}
        loading={loading}
        query={query}
        onQueryChange={setQuery}
        rowKey={rowKey}
        selection={
          selectable
            ? {
              selected,
              onChange: setSelected,
              actions: (
                <ConfirmButton
                  variant="destructive"
                  size="sm"
                  loading={deletingMany}
                  description={`${selected.size}`}
                  onConfirm={deleteSelected}
                >
                  {i18next.t("general:Delete")}
                </ConfirmButton>
              ),
            }
            : undefined
        }
      />
    </div>
  );
}
