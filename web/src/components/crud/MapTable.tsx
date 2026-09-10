import * as React from "react";
import {Input} from "@/components/ui/input";
import {EditableTable} from "@/components/crud/EditableTable";

type MapValue = Record<string, string> | null | undefined;

interface MapRow {
  key: string;
  value: string;
}

function toRows(map: MapValue): MapRow[] {
  if (!map || typeof map !== "object" || Array.isArray(map)) {
    return [];
  }
  return Object.entries(map).map(([key, value]) => ({key, value: `${value ?? ""}`}));
}

interface MapTableProps {
  value: MapValue;
  onChange: (value: Record<string, string>) => void;
  keyTitle: React.ReactNode;
  valueTitle: React.ReactNode;
  keyWidth?: number | string;
}

/**
 * Edits a `map[string]string` column (user/group/product properties, LDAP
 * custom attributes, provider HTTP headers) as a two-column table.
 *
 * The rows are kept here rather than rebuilt from the map on every render: a
 * map cannot hold a new row whose key is still empty, or two rows whose keys
 * collide while being typed, so rebuilding would drop them. The antd tables
 * (`PropertyTable`, `AttributesMapperTable`, `HttpHeaderTable`) did the same.
 */
export function MapTable({value, onChange, keyTitle, valueTitle, keyWidth}: MapTableProps) {
  const [rows, setRows] = React.useState(() => toRows(value));
  // the map last seen from the parent; a different object (another record
  // loaded, a reload after save) replaces the rows
  const [source, setSource] = React.useState(value);
  if (value !== source) {
    setSource(value);
    setRows(toRows(value));
  }

  return (
    <EditableTable<MapRow>
      rows={rows}
      onChange={(next) => {
        const map: Record<string, string> = {};
        next.forEach((row) => {
          map[row.key] = row.value;
        });
        setRows(next);
        setSource(map);
        onChange(map);
      }}
      newRow={() => ({key: "", value: ""})}
      reorderable={false}
      columns={[
        {
          key: "key",
          title: keyTitle,
          width: keyWidth,
          render: (row, _i, patch) => <Input value={row.key} onChange={(e) => patch({key: e.target.value})} />,
        },
        {
          key: "value",
          title: valueTitle,
          render: (row, _i, patch) => <Input value={row.value} onChange={(e) => patch({value: e.target.value})} />,
        },
      ]}
    />
  );
}
