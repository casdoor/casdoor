import i18next from "i18next";
import * as Setting from "@/lib/setting";

/** "view" is the read-only page a non-admin gets, as in the antd frontend. */
export type EditMode = "add" | "edit" | "view";

/**
 * The antd pages title themselves "New X" / "View X" / "Edit X" depending on the
 * mode, and the three keys always differ by that one word — so the "Edit X" key
 * a page declares is enough to reach the other two. A locale missing "New X"
 * falls back to the key text, which is what the antd frontend shows too.
 */
export function getModeTitleKey(editTitleKey: string, mode: EditMode): string {
  if (mode === "add") {
    return editTitleKey.replace(":Edit ", ":New ");
  }
  if (mode === "view") {
    return editTitleKey.replace(":Edit ", ":View ");
  }
  return editTitleKey;
}

export interface CasdoorResponse<T = any> {
  status: "ok" | "error";
  msg?: string;
  data?: T;
  data2?: any;
}

/**
 * Shared "Save" behaviour of every Casdoor edit page: POST to add-x or update-x,
 * report the outcome and let the caller decide where to navigate afterwards.
 */
export async function submitEdit<T extends Record<string, any>>(options: {
  mode: EditMode;
  record: T;
  add: (record: T) => Promise<CasdoorResponse>;
  update: (record: T) => Promise<CasdoorResponse>;
  /** called only when the backend accepted the change */
  onSaved?: (record: T, res: CasdoorResponse) => void;
  /** called when the backend rejected the change */
  onFailed?: (res: CasdoorResponse) => void;
}): Promise<boolean> {
  const {mode, record, add, update, onSaved, onFailed} = options;
  try {
    const res = mode === "add" ? await add(record) : await update(record);
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("general:Successfully saved"));
      onSaved?.(record, res);
      return true;
    }
    Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
    onFailed?.(res);
    return false;
  } catch (error: any) {
    Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    return false;
  }
}

/** Shared "Delete" behaviour of every Casdoor list page. */
export async function submitDelete<T>(options: {
  record: T;
  remove: (record: T) => Promise<CasdoorResponse>;
  onDeleted?: () => void;
}): Promise<boolean> {
  const {record, remove, onDeleted} = options;
  try {
    const res = await remove(record);
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("general:Successfully deleted"));
      onDeleted?.();
      return true;
    }
    Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
    return false;
  } catch (error: any) {
    Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    return false;
  }
}

/**
 * Deletes a selection one record at a time.
 *
 * There is no batch delete endpoint, so this is the same call the row action
 * makes, repeated in order — sequentially rather than in parallel, so a hundred
 * selected rows do not open a hundred connections at once. It reports one
 * summary instead of one toast per record.
 */
export async function submitDeleteMany<T>(options: {
  records: T[];
  remove: (record: T) => Promise<CasdoorResponse>;
  onDeleted?: (deleted: number) => void;
}): Promise<number> {
  const {records, remove, onDeleted} = options;
  let deleted = 0;
  const failures: string[] = [];

  for (const record of records) {
    try {
      const res = await remove(record);
      if (res.status === "ok") {
        deleted++;
      } else {
        failures.push(res.msg ?? "");
      }
    } catch (error: any) {
      failures.push(String(error));
    }
  }

  if (deleted > 0) {
    Setting.showMessage("success", `${i18next.t("general:Successfully deleted")}: ${deleted}`);
  }
  if (failures.length > 0) {
    Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${failures.length} - ${failures[0]}`);
  }
  onDeleted?.(deleted);
  return deleted;
}

/** Shared "Add" behaviour: create the record and jump to its edit page. */
export async function submitAdd<T>(options: {
  record: T;
  add: (record: T) => Promise<CasdoorResponse>;
  onAdded?: (res: CasdoorResponse) => void;
}): Promise<boolean> {
  const {record, add, onAdded} = options;
  try {
    const res = await add(record);
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("general:Successfully added"));
      onAdded?.(res);
      return true;
    }
    Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
    return false;
  } catch (error: any) {
    Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    return false;
  }
}

/**
 * Several Casdoor columns are `map[string]string` on the wire but are edited as
 * a two-column table (LDAP custom attributes, provider HTTP headers, user
 * properties). These convert between the two; the antd frontend did the same
 * inside `AttributesMapperTable` / `HttpHeaderTable`.
 *
 * Feeding the map straight to `EditableTable` would crash the page on `.map`,
 * and saving the rows back would replace the object with an array.
 */
export function mapToRows(
  map: Record<string, string> | null | undefined,
  keyName: string,
  valueName: string,
): Record<string, string>[] {
  if (!map || typeof map !== "object" || Array.isArray(map)) {
    return [];
  }
  return Object.entries(map).map(([key, value]) => ({[keyName]: key, [valueName]: `${value ?? ""}`}));
}

export function rowsToMap(
  rows: Record<string, any>[] | null | undefined,
  keyName: string,
  valueName: string,
): Record<string, string> {
  const map: Record<string, string> = {};
  (rows ?? []).forEach((row) => {
    map[row?.[keyName] ?? ""] = row?.[valueName] ?? "";
  });
  return map;
}
