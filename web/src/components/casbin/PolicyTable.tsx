import * as React from "react";
import i18next from "i18next";
import {ChevronLeft, ChevronRight, Pencil, Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {SelectField} from "@/components/common/SelectField";
import * as AdapterBackend from "@/backend/AdapterBackend";
import type {EditMode} from "@/lib/crud";
import * as Setting from "@/lib/setting";

const COLUMN_KEYS = ["V0", "V1", "V2", "V3", "V4", "V5"];
const PAGE_SIZE = 100;

interface Policy {
  key: number;
  Ptype: string;
  [field: string]: any;
}

interface PolicyTableProps {
  enforcer: any;
  /** the model's sections, e.g. {p: "sub,obj,act", g: "_,_"} — the column titles come from `p` */
  modelCfg?: Record<string, string>;
  mode: EditMode;
}

/**
 * The Casbin policies of an enforcer, ported from web/src/table/PolicyTable.js.
 * Rows are edited one at a time and each save/delete goes straight to the
 * adapter through /api/update-policy, /api/add-policy and /api/remove-policy.
 */
export function PolicyTable({enforcer, modelCfg, mode}: PolicyTableProps) {
  const [policies, setPolicies] = React.useState<Policy[]>([]);
  const [loading, setLoading] = React.useState(false);
  const [editingIndex, setEditingIndex] = React.useState<number | null>(null);
  const [oldPolicy, setOldPolicy] = React.useState<Policy | null>(null);
  const [adding, setAdding] = React.useState(false);
  const [page, setPage] = React.useState(1);

  const isBuiltIn = Setting.builtInObject(enforcer);
  const canSync = enforcer.model !== "" && enforcer.adapter !== "";

  const getPolicies = React.useCallback(() => {
    setLoading(true);
    AdapterBackend.getPolicies(enforcer.owner, enforcer.name)
      .then((res: any) => {
        if (res.status === "ok") {
          setPolicies((res.data ?? []).map((policy: any, index: number) => ({...policy, key: index})));
        } else {
          Setting.showMessage("error", `${i18next.t("adapter:Failed to sync policies")}: ${res.msg}`);
        }
        setLoading(false);
      })
      .catch((error: any) => {
        setLoading(false);
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }, [enforcer.owner, enforcer.name]);

  React.useEffect(() => {
    if (mode === "edit" && enforcer.adapter !== "") {
      getPolicies();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (modelCfg === undefined) {
    return (
      <Button disabled={!canSync} onClick={getPolicies}>
        {i18next.t("general:Sync")}
      </Button>
    );
  }

  const columnTitles = modelCfg["p"] ? modelCfg["p"].split(",") : COLUMN_KEYS;
  const start = (page - 1) * PAGE_SIZE;
  const rows = policies.slice(start, start + PAGE_SIZE);
  const pageCount = Math.max(1, Math.ceil(policies.length / PAGE_SIZE));

  const updateField = (index: number, field: string, value: string) => {
    setPolicies((prev) => {
      const next = [...prev];
      next[index] = {...next[index], [field]: value};
      return next;
    });
  };

  const removeRow = (index: number) => {
    setPolicies((prev) => [...prev.slice(0, index), ...prev.slice(index + 1)]);
  };

  const addRow = () => {
    const row: Policy = {key: Date.now(), Ptype: "p"};
    setPolicies((prev) => [row, ...prev]);
    setPage(1);
    setEditingIndex(0);
    setOldPolicy(Setting.deepCopy(row));
    setAdding(true);
  };

  const cancel = (index: number) => {
    if (adding) {
      removeRow(index);
      setAdding(false);
    } else if (oldPolicy !== null) {
      setPolicies((prev) => {
        const next = [...prev];
        next[index] = {...oldPolicy};
        return next;
      });
    }
    setEditingIndex(null);
    setOldPolicy(null);
  };

  const save = (index: number) => {
    const policy = policies[index];
    if (adding) {
      AdapterBackend.AddPolicy(enforcer.owner, enforcer.name, policy).then((res: any) => {
        if (res.status === "ok") {
          setEditingIndex(null);
          setOldPolicy(null);
          setAdding(false);
          if (res.data !== "Affected") {
            Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${i18next.t("adapter:Duplicated policy rules")}`);
          } else {
            Setting.showMessage("success", i18next.t("general:Successfully added"));
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      });
      return;
    }

    AdapterBackend.UpdatePolicy(enforcer.owner, enforcer.name, [oldPolicy, policy]).then((res: any) => {
      if (res.status === "ok") {
        setEditingIndex(null);
        setOldPolicy(null);
        Setting.showMessage("success", i18next.t("general:Successfully saved"));
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
      }
    });
  };

  const deletePolicy = (index: number) => {
    AdapterBackend.RemovePolicy(enforcer.owner, enforcer.name, policies[index]).then((res: any) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully deleted"));
        removeRow(index);
      } else {
        Setting.showMessage("error", i18next.t("general:Failed to delete"));
      }
    });
  };

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center gap-2">
        <Button disabled={editingIndex !== null || !canSync} loading={loading} onClick={getPolicies}>
          {i18next.t("general:Sync")}
        </Button>
        <Button
          variant="outline"
          disabled={editingIndex !== null || !canSync || isBuiltIn}
          onClick={addRow}
        >
          {i18next.t("general:Add")}
        </Button>
      </div>
      <div className="overflow-x-auto rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="w-[110px]">{i18next.t("adapter:Rule type")}</TableHead>
              {columnTitles.map((title, i) => (
                <TableHead key={COLUMN_KEYS[i]}>{title}</TableHead>
              ))}
              <TableHead className="w-[150px]">{i18next.t("general:Action")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.length === 0 ? (
              <TableRow className="hover:bg-transparent">
                <TableCell colSpan={columnTitles.length + 2} className="h-20 text-center text-muted-foreground">
                  {i18next.t("general:No data")}
                </TableCell>
              </TableRow>
            ) : (
              rows.map((policy, rowIndex) => {
                const index = start + rowIndex;
                const editing = editingIndex === index;
                return (
                  <TableRow key={policy.key}>
                    <TableCell>
                      {editing ? (
                        <SelectField
                          value={policy.Ptype}
                          onChange={(value) => updateField(index, "Ptype", value)}
                          options={Object.keys(modelCfg).reverse().map((item) => ({id: item, name: item}))}
                        />
                      ) : policy.Ptype}
                    </TableCell>
                    {columnTitles.map((title, i) => (
                      <TableCell key={COLUMN_KEYS[i]}>
                        {editing ? (
                          <Input
                            value={policy[COLUMN_KEYS[i]] ?? ""}
                            onChange={(e) => updateField(index, COLUMN_KEYS[i], e.target.value)}
                          />
                        ) : policy[COLUMN_KEYS[i]]}
                      </TableCell>
                    ))}
                    <TableCell>
                      {editing ? (
                        <div className="flex items-center gap-1">
                          <Button size="sm" onClick={() => save(index)}>{i18next.t("general:Save")}</Button>
                          <Button variant="outline" size="sm" onClick={() => cancel(index)}>{i18next.t("general:Cancel")}</Button>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="iconSm"
                            aria-label={i18next.t("general:Edit")}
                            disabled={editingIndex !== null || isBuiltIn}
                            onClick={() => {
                              setEditingIndex(index);
                              setOldPolicy(Setting.deepCopy(policy));
                            }}
                          >
                            <Pencil />
                          </Button>
                          <Button
                            variant="ghost"
                            size="iconSm"
                            className="text-destructive"
                            aria-label={i18next.t("general:Delete")}
                            disabled={editingIndex !== null || isBuiltIn}
                            onClick={() => deletePolicy(index)}
                          >
                            <Trash2 />
                          </Button>
                        </div>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>
      {pageCount > 1 ? (
        <div className="flex items-center justify-end gap-2 text-sm text-muted-foreground">
          <span>{`${start + 1}-${Math.min(start + PAGE_SIZE, policies.length)} / ${policies.length}`}</span>
          <Button variant="outline" size="iconSm" disabled={page === 1} onClick={() => setPage(page - 1)} aria-label="Previous">
            <ChevronLeft />
          </Button>
          <Button variant="outline" size="iconSm" disabled={page === pageCount} onClick={() => setPage(page + 1)} aria-label="Next">
            <ChevronRight />
          </Button>
        </div>
      ) : null}
    </div>
  );
}
