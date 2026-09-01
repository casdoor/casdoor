import * as React from "react";
import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Switch} from "@/components/ui/switch";
import type {ColumnDef, ColumnFilterOption} from "@/components/crud/types";
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
  /** set to false where the antd table does not pin the name column */
  fixed?: false;
}): ColumnDef<T> {
  const {dataIndex, title, to, width = 140, sortable = true, searchable = true, text, fixed} = options;
  return {
    dataIndex,
    title: title ?? i18next.t("general:Name"),
    width,
    sortable,
    searchable,
    // antd pins the name column; DataTable ignores this unless the column
    // actually leads the table, so a page that puts it later is unaffected
    fixed: fixed === false ? undefined : "left",
    render: (value, record) => (
      <Link to={to(record)} className="font-medium text-foreground underline-offset-4 hover:underline">
        {text ? text(record) : value}
      </Link>
    ),
  };
}

export function dateColumn<T>(dataIndex = "createdTime", title?: React.ReactNode, options?: {sortable?: boolean; searchable?: boolean}): ColumnDef<T> {
  return {
    dataIndex,
    title: title ?? i18next.t("general:Created time"),
    width: 165,
    sortable: options?.sortable ?? true,
    searchable: options?.searchable ?? false,
    render: (value) => (
      <span className="whitespace-nowrap tabular-nums text-muted-foreground">{Setting.getFormattedDate(value)}</span>
    ),
  };
}

/**
 * A read-only boolean cell. The antd tables show a disabled `<Switch>` labelled
 * ON / OFF, so this pairs the same switch with that wording; `invertColor` paints
 * the "on" state as a warning, for flags like `isForbidden` where on is the bad one.
 */
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
      <span className="inline-flex items-center gap-1.5">
        <Switch
          checked={Boolean(value)}
          disabled
          aria-readonly
          className={cn("opacity-100", value && invertColor && "data-[state=checked]:bg-destructive")}
        />
        <span className="text-xs text-muted-foreground">
          {value ? i18next.t("general:ON") : i18next.t("general:OFF")}
        </span>
      </span>
    ),
  };
}

export function tagsColumn<T>(options: {
  dataIndex: string;
  title: React.ReactNode;
  width?: number | string;
  sortable?: boolean;
  searchable?: boolean;
  to?: (value: string) => string;
}): ColumnDef<T> {
  const {dataIndex, title, width, sortable = false, searchable = false, to} = options;
  return {
    dataIndex,
    title,
    width,
    sortable,
    searchable,
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
  sortable?: boolean;
  searchable?: boolean;
  max?: number;
}): ColumnDef<T> {
  const {dataIndex, title, urlPrefix, width, sortable = false, searchable = false, max = 6} = options;
  return {
    dataIndex,
    title,
    width,
    sortable,
    searchable,
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

export function organizationColumn<T>(
  width: number | string = 140,
  dataIndex = "owner",
  /** the rule and site lists head the same column "Owner" instead */
  title: React.ReactNode = i18next.t("general:Organization"),
  /** pinned on the lists that lead with it and pin the name behind it */
  fixed?: "left",
): ColumnDef<T> {
  return {
    dataIndex,
    title,
    width,
    sortable: true,
    searchable: true,
    fixed,
    render: (value) => (
      <Link to={`/organizations/${value}`} className="underline-offset-4 hover:underline">
        {value}
      </Link>
    ),
  };
}

/**
 * The Client IP cell, which the antd tables link to db-ip.com so an admin can
 * look up where a sign-in or a request came from.
 */
export function clientIpColumn<T>(options?: {
  dataIndex?: string;
  title?: React.ReactNode;
  width?: number | string;
  /**
   * Cleans the stored value before it is shown and looked up. The verification
   * rows keep the address as "1.2.3.4: ", which the antd column trims.
   */
  normalize?: (value: string) => string;
}): ColumnDef<T> {
  const {dataIndex = "clientIp", title, width = 140, normalize} = options ?? {};
  const column: ColumnDef<T> = {
    dataIndex,
    title: title ?? i18next.t("general:Client IP"),
    width,
    sortable: true,
    searchable: true,
  };
  if (!normalize) {
    // no rewriting, so the plain cell keeps its search highlight inside the link
    column.link = (value) => (value ? `https://db-ip.com/${value}` : undefined);
    column.linkExternal = true;
    return column;
  }
  column.render = (value) => {
    const ip = normalize(String(value ?? ""));
    return ip ? (
      <a href={`https://db-ip.com/${ip}`} target="_blank" rel="noreferrer" className="underline-offset-4 hover:underline">
        {ip}
      </a>
    ) : null;
  };
  return column;
}

export function textColumn<T>(options: {
  dataIndex: string;
  title: React.ReactNode;
  width?: number | string;
  sortable?: boolean;
  searchable?: boolean;
  /** the header filter menu the antd column declared as `filters` */
  filters?: ColumnFilterOption[];
  /** pins the column while the table scrolls sideways, antd's `fixed` */
  fixed?: "left" | "right";
  /** turns the cell into a link, keeping the search highlight inside it */
  link?: (value: any, record: T) => string | undefined;
  /** the link leaves Casdoor, so use an `<a target="_blank">` */
  linkExternal?: boolean;
  mono?: boolean;
  className?: string;
}): ColumnDef<T> {
  const {dataIndex, title, width, sortable = true, searchable = false, filters, fixed, link, linkExternal, mono, className} =
    options;
  return {
    dataIndex,
    title,
    width,
    sortable,
    searchable,
    filters,
    fixed,
    link,
    linkExternal,
    className: cn(mono && "font-mono text-xs", className),
  };
}

/**
 * A cell holding a URL: shortened, and opening in a new tab. The antd tables cut
 * these to 40-ish characters because a webhook or agent URL is far longer than
 * its column.
 */
export function urlColumn<T>(options: {
  dataIndex: string;
  title: React.ReactNode;
  width?: number | string;
  searchable?: boolean;
  /** what to open; defaults to the cell value */
  href?: (value: string, record: T) => string;
  /** characters to keep, as `Setting.getShortText` counts them */
  max?: number;
}): ColumnDef<T> {
  const {dataIndex, title, width, searchable = true, href, max} = options;
  return {
    dataIndex,
    title,
    width,
    sortable: true,
    searchable,
    render: (value: string, record: T) =>
      value ? (
        <a
          href={href ? href(value, record) : value}
          target="_blank"
          rel="noreferrer"
          className="underline-offset-4 hover:underline"
        >
          {max === undefined ? Setting.getShortText(value) : Setting.getShortText(value, max)}
        </a>
      ) : null,
  };
}

/** turns a plain list of stored values into a filter menu that shows them as-is */
export function valueFilters(values: string[]): ColumnFilterOption[] {
  return values.map((value) => ({value, label: value}));
}
