import i18next from "i18next";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {FormRow} from "@/components/crud/FormRow";
import * as Conf from "@/Conf";
import * as Setting from "@/lib/setting";

export interface ThemeData {
  themeType?: string;
  colorPrimary?: string;
  borderRadius?: number;
  isCompact?: boolean;
  isEnabled?: boolean;
}

/**
 * Editor for the `themeData` an organization or application can override.
 * The shape is the one Casdoor stores (`Conf.ThemeDefault`); the antd frontend
 * rendered it with antd-token-previewer, here it is plain controls.
 */
export function ThemeEditor({
  themeData,
  onChange,
}: {
  themeData?: ThemeData | null;
  onChange: (next: ThemeData) => void;
}) {
  const current: ThemeData = themeData ?? {...Conf.ThemeDefault, isEnabled: false};
  const update = (patch: Partial<ThemeData>) => onChange({...current, ...patch});

  return (
    <div className="space-y-1 rounded-lg border p-4">
      <FormRow label={i18next.t("organization:Follow global theme")}>
        <Switch
          checked={!current.isEnabled}
          onCheckedChange={(v) => update({isEnabled: !v})}
        />
      </FormRow>
      {current.isEnabled ? (
        <>
          <FormRow label={i18next.t("theme:Primary color")}>
            <div className="flex items-center gap-2">
              <input
                type="color"
                className="h-9 w-12 cursor-pointer rounded-md border border-input bg-transparent p-1"
                value={current.colorPrimary ?? Conf.ThemeDefault.colorPrimary}
                onChange={(e) => update({colorPrimary: e.target.value})}
              />
              <Input
                className="w-32 font-mono text-xs"
                value={current.colorPrimary ?? ""}
                onChange={(e) => update({colorPrimary: e.target.value})}
              />
            </div>
          </FormRow>
          <FormRow label={i18next.t("theme:Border radius")}>
            <Input
              type="number"
              className="w-32"
              value={current.borderRadius ?? Conf.ThemeDefault.borderRadius}
              onChange={(e) => update({borderRadius: Setting.myParseInt(e.target.value)})}
            />
          </FormRow>
          <FormRow label={i18next.t("theme:Is compact")}>
            <Switch checked={!!current.isCompact} onCheckedChange={(v) => update({isCompact: v})} />
          </FormRow>
        </>
      ) : null}
    </div>
  );
}
