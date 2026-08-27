import * as React from "react";
import i18next from "i18next";
import {ArrowDown, ArrowUp, Plus, Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {cn} from "@/lib/utils";

export interface EditableColumn<T> {
  key: string;
  title: React.ReactNode;
  width?: number | string;
  render: (row: T, index: number, update: (patch: Partial<T> | Record<string, any>) => void) => React.ReactNode;
}

interface EditableTableProps<T> {
  title?: React.ReactNode;
  rows: T[] | undefined | null;
  onChange: (rows: T[]) => void;
  columns: EditableColumn<T>[];
  newRow?: () => T;
  addLabel?: React.ReactNode;
  /** hide the up/down buttons for tables where order does not matter */
  reorderable?: boolean;
  rowKey?: (row: T, index: number) => string;
  className?: string;
  emptyText?: React.ReactNode;
  disabled?: boolean;
}

/**
 * The inline, editable sub-table used all over the Casdoor edit pages
 * (account items, signup items, MFA items, provider lists, ...).
 */
export function EditableTable<T>({
  title,
  rows,
  onChange,
  columns,
  newRow,
  addLabel,
  reorderable = true,
  rowKey,
  className,
  emptyText,
  disabled,
}: EditableTableProps<T>) {
  const items = rows ?? [];

  const update = (index: number, patch: Partial<T> | Record<string, any>) => {
    const next = [...items];
    next[index] = {...(next[index] as any), ...(patch as any)};
    onChange(next);
  };

  const remove = (index: number) => onChange([...items.slice(0, index), ...items.slice(index + 1)]);

  const swap = (i: number, j: number) => {
    if (j < 0 || j >= items.length) {
      return;
    }
    const next = [...items];
    [next[i], next[j]] = [next[j], next[i]];
    onChange(next);
  };

  return (
    <div className={cn("space-y-2", className)}>
      {(title || newRow) && (
        <div className="flex items-center justify-between gap-2">
          {title ? <div className="text-sm font-medium">{title}</div> : <span />}
          {newRow && !disabled ? (
            <Button variant="outline" size="sm" onClick={() => onChange([...items, newRow()])}>
              <Plus />
              {addLabel ?? i18next.t("general:Add")}
            </Button>
          ) : null}
        </div>
      )}
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              {columns.map((column) => (
                <TableHead key={column.key} style={column.width ? {width: column.width} : undefined}>
                  {column.title}
                </TableHead>
              ))}
              {!disabled ? <TableHead className="w-[130px]">{i18next.t("general:Action")}</TableHead> : null}
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.length === 0 ? (
              <TableRow className="hover:bg-transparent">
                <TableCell colSpan={columns.length + 1} className="h-20 text-center text-muted-foreground">
                  {emptyText ?? i18next.t("general:No data")}
                </TableCell>
              </TableRow>
            ) : (
              items.map((row, index) => (
                <TableRow key={rowKey ? rowKey(row, index) : index}>
                  {columns.map((column) => (
                    <TableCell key={column.key} className="align-top">
                      {column.render(row, index, (patch) => update(index, patch))}
                    </TableCell>
                  ))}
                  {!disabled ? (
                    <TableCell className="align-top">
                      <div className="flex items-center gap-1">
                        {reorderable ? (
                          <>
                            <Button
                              variant="ghost"
                              size="iconSm"
                              disabled={index === 0}
                              aria-label="Move up"
                              onClick={() => swap(index, index - 1)}
                            >
                              <ArrowUp />
                            </Button>
                            <Button
                              variant="ghost"
                              size="iconSm"
                              disabled={index === items.length - 1}
                              aria-label="Move down"
                              onClick={() => swap(index, index + 1)}
                            >
                              <ArrowDown />
                            </Button>
                          </>
                        ) : null}
                        <Button
                          variant="ghost"
                          size="iconSm"
                          className="text-destructive"
                          aria-label={i18next.t("general:Delete")}
                          onClick={() => remove(index)}
                        >
                          <Trash2 />
                        </Button>
                      </div>
                    </TableCell>
                  ) : null}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
