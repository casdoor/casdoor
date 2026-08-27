import * as React from "react";
import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import type {ColumnDef} from "@/components/crud/types";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

/** A cell that links to the edit page of the row. */
export function linkColumn<T extends Record<string, any>>(options: {
  dataIndex: string;
  title?: React.ReactNode;
  to: (record: T) => string;
  width?: number | string;
  sortable?: boolean;
  searchable?: boolean;
  text?: (record: T) => string;
}): ColumnDef<T> {
  const {dataIndex, title, to, width = 140, sortable = true, searchable = true, text} = options;
  return {
    dataIndex,
    title: title ?? i18next.t("general:Name"),
    width,
    sortable,
    searchable,
    render: (value, record) => (
      <Link to={to(record)} className="font-medium text-foreground underline-offset-4 hover:underline">
        {text ? text(record) : value}
      </Link>
    ),
  };
}

export function dateColumn<T>(dataIndex = "createdTime", title?: React.ReactNode): ColumnDef<T> {
  return {
    dataIndex,
    title: title ?? i18next.t("general:Created time"),
    width: 165,
    sortable: true,
    render: (value) => (
      <span className="whitespace-nowrap tabular-nums text-muted-foreground">{Setting.getFormattedDate(value)}</span>
    ),
  };
}

export function boolColumn<T>(options: {
  dataIndex: string;
  title: React.ReactNode;
  width?: number | string;
  invertColor?: boolean;
}): ColumnDef<T> {
  const {dataIndex, title, width = 110, invertColor} = options;
  return {
    dataIndex,
    title,
    width,
    sortable: true,
    align: "center",
    render: (value) => (
      <Badge variant={value ? (invertColor ? "destructive" : "success") : "secondary"}>
        {value ? i18next.t("general:True") : i18next.t("general:False")}
      </Badge>
    ),
  };
}

export function tagsColumn<T>(options: {
  dataIndex: string;
  title: React.ReactNode;
  width?: number | string;
  to?: (value: string) => string;
}): ColumnDef<T> {
  const {dataIndex, title, width, to} = options;
  return {
    dataIndex,
    title,
    width,
    render: (value: string[]) => {
      if (!value || value.length === 0) {
        return null;
      }
      return (
        <div className="flex flex-wrap gap-1">
          {value.map((item) => (
            <Badge key={item} variant="secondary" className="font-normal">
              {to ? (
                <Link to={to(item)} className="underline-offset-2 hover:underline">
                  {Setting.getShortName(item)}
                </Link>
              ) : (
                item
              )}
            </Badge>
          ))}
        </div>
      );
    },
  };
}

/** Renders a list of "owner/name" references as links to their edit pages. */
export function refsColumn<T>(options: {
  dataIndex: string;
  title: React.ReactNode;
  urlPrefix: string;
  width?: number | string;
  max?: number;
}): ColumnDef<T> {
  const {dataIndex, title, urlPrefix, width, max = 6} = options;
  return {
    dataIndex,
    title,
    width,
    render: (value: any[]) => {
      if (!value || value.length === 0) {
        return null;
      }
      const items = value.slice(0, max);
      return (
        <div className="flex flex-wrap gap-1">
          {items.map((item: any) => {
            const id = typeof item === "string" ? item : `${item.owner}/${item.name}`;
            const label = typeof item === "string" ? Setting.getShortName(item) : item.name;
            return (
              <Badge key={id} variant="secondary" className="font-normal">
                <Link to={`${urlPrefix}/${id}`} className="underline-offset-2 hover:underline">
                  {label}
                </Link>
              </Badge>
            );
          })}
          {value.length > max ? <Badge variant="outline">+{value.length - max}</Badge> : null}
        </div>
      );
    },
  };
}

export function organizationColumn<T>(width: number | string = 140): ColumnDef<T> {
  return {
    dataIndex: "owner",
    title: i18next.t("general:Organization"),
    width,
    sortable: true,
    searchable: true,
    render: (value) => (
      <Link to={`/organizations/${value}`} className="underline-offset-4 hover:underline">
        {value}
      </Link>
    ),
  };
}

export function textColumn<T>(options: {
  dataIndex: string;
  title: React.ReactNode;
  width?: number | string;
  sortable?: boolean;
  searchable?: boolean;
  mono?: boolean;
  className?: string;
}): ColumnDef<T> {
  const {dataIndex, title, width, sortable = true, searchable = false, mono, className} = options;
  return {
    dataIndex,
    title,
    width,
    sortable,
    searchable,
    className: cn(mono && "font-mono text-xs", className),
  };
}
