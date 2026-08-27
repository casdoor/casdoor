import i18next from "i18next";
import * as Setting from "@/lib/setting";

export type EditMode = "add" | "edit";

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
