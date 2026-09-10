// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import * as React from "react";
import i18next from "i18next";
import {Input} from "@/components/ui/input";
import {EditableTable} from "@/components/crud/EditableTable";

interface PropertiesTableProps {
  properties: Record<string, string> | null | undefined;
  onChange: (properties: Record<string, string>) => void;
}

function toRows(properties: PropertiesTableProps["properties"]) {
  return Object.entries(properties ?? {}).map(([key, value]) => ({key, value}));
}

export function PropertiesTable({properties, onChange}: PropertiesTableProps) {
  const [source, setSource] = React.useState(properties);
  const [rows, setRows] = React.useState(() => toRows(properties));

  if (properties !== source) {
    setSource(properties);
    setRows(toRows(properties));
  }

  return (
    <EditableTable
      rows={rows}
      onChange={(nextRows) => {
        const next = Object.fromEntries(nextRows.map((row) => [row.key, row.value]));
        // The map cannot represent multiple rows with the same unfinished key.
        setRows(nextRows);
        setSource(next);
        onChange(next);
      }}
      newRow={() => ({key: "", value: ""})}
      reorderable={false}
      columns={[
        {
          key: "key",
          title: i18next.t("general:Name"),
          width: 240,
          render: (row, _i, patch) => (
            <Input value={row.key} onChange={(e) => patch({key: e.target.value})} />
          ),
        },
        {
          key: "value",
          title: i18next.t("webhook:Value"),
          render: (row, _i, patch) => (
            <Input value={row.value} onChange={(e) => patch({value: e.target.value})} />
          ),
        },
      ]}
    />
  );
}
