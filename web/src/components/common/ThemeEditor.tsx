import i18next from "i18next";
import {Check} from "lucide-react";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {FormRow} from "@/components/crud/FormRow";
import * as Conf from "@/Conf";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

export interface ThemeData {
  themeType?: string;
  colorPrimary?: string;
  borderRadius?: number;
  isCompact?: boolean;
  isEnabled?: boolean;
}

export const BLUE_COLOR = "#1677FF";
export const PINK_COLOR = "#ED4192";
export const GREEN_COLOR = "#00B96B";

/** the swatches of web/src/common/theme/ColorPicker.js */
const PRESET_COLORS = [
  BLUE_COLOR, "#5734d3", "#9E339F", PINK_COLOR, "#E0282E", "#F4801A", "#F2BD27", GREEN_COLOR,
];

/** picking a theme resets the tokens it defines, exactly as the antd editor does */
const THEMES_INFO: Record<string, Partial<ThemeData>> = {
  default: {},
  dark: {borderRadius: 2},
  lark: {colorPrimary: GREEN_COLOR, borderRadius: 4},
  comic: {colorPrimary: PINK_COLOR, borderRadius: 16},
};

const THEME_TYPES = [
  {value: "default", labelKey: "general:Default"},
  {value: "dark", labelKey: "theme:Dark"},
  {value: "lark", labelKey: "theme:Document"},
  {value: "comic", labelKey: "theme:Blossom"},
];

/**
 * Editor for the `themeData` an organization or application can override.
 * The stored shape is the one Casdoor expects (`Conf.ThemeDefault` plus
 * `themeType` / `isCompact`); the antd frontend rendered it with
 * antd-token-previewer, here it is shadcn controls over the same values.
 */
export function ThemeEditor({
  themeData,
  onChange,
  followLabelKey = "organization:Follow global theme",
}: {
  themeData?: ThemeData | null;
  onChange: (next: ThemeData) => void;
  /** what the object inherits from when it does not override: the global theme, or its organization's */
  followLabelKey?: string;
}) {
  const current: ThemeData = themeData ?? {...Conf.ThemeDefault, isEnabled: false};
  const update = (patch: Partial<ThemeData>) => onChange({...current, ...patch});

  const themeType = current.themeType ?? Conf.ThemeDefault.themeType;
  const colorPrimary = current.colorPrimary ?? Conf.ThemeDefault.colorPrimary;
  const borderRadius = current.borderRadius ?? Conf.ThemeDefault.borderRadius;

  const selectTheme = (value: string) =>
    // a theme carries its own primary colour and radius, so re-apply the preset
    update({...Conf.ThemeDefault, ...current, themeType: value, ...THEMES_INFO[value]});

  return (
    <div className="space-y-1 rounded-lg border p-4">
      <FormRow label={i18next.t(followLabelKey)}>
        <Switch checked={!current.isEnabled} onCheckedChange={(v) => update({isEnabled: !v})} />
      </FormRow>
      {current.isEnabled ? (
        <>
          <FormRow label={i18next.t("theme:Theme")}>
            <div className="flex flex-wrap gap-3">
              {THEME_TYPES.map((item) => (
                <button
                  key={item.value}
                  type="button"
                  onClick={() => selectTheme(item.value)}
                  className={cn(
                    "w-28 overflow-hidden rounded-lg border-2 bg-white p-1 text-center transition-colors",
                    themeType === item.value ? "border-primary" : "border-transparent hover:border-border",
                  )}
                >
                  <img
                    className="h-16 w-full object-contain"
                    src={`${Setting.StaticBaseUrl}/img/theme_${item.value}.svg`}
                    alt={item.value}
                  />
                  <span className="block pt-1 text-xs text-neutral-700">{i18next.t(item.labelKey)}</span>
                </button>
              ))}
            </div>
          </FormRow>
          <FormRow label={i18next.t("theme:Primary color")}>
            <div className="flex flex-wrap items-center gap-2">
              {PRESET_COLORS.map((color) => (
                <button
                  key={color}
                  type="button"
                  aria-label={color}
                  onClick={() => update({colorPrimary: color})}
                  className="flex h-6 w-6 items-center justify-center rounded-md ring-offset-background transition-transform hover:scale-110"
                  style={{backgroundColor: color}}
                >
                  {colorPrimary?.toUpperCase() === color.toUpperCase() ? (
                    <Check className="h-4 w-4 text-white" />
                  ) : null}
                </button>
              ))}
              <input
                type="color"
                aria-label={i18next.t("theme:Primary color")}
                className="h-8 w-10 cursor-pointer rounded-md border border-input bg-transparent p-1"
                value={colorPrimary}
                onChange={(e) => update({colorPrimary: e.target.value})}
              />
              <Input
                className="w-28 font-mono text-xs"
                value={current.colorPrimary ?? ""}
                onChange={(e) => update({colorPrimary: e.target.value})}
              />
            </div>
          </FormRow>
          <FormRow label={i18next.t("theme:Border radius")}>
            <div className="flex items-center gap-3">
              <Input
                type="number"
                min={0}
                className="w-24"
                value={borderRadius}
                onChange={(e) => update({borderRadius: Setting.myParseInt(e.target.value)})}
              />
              <input
                type="range"
                min={0}
                max={20}
                className="h-1.5 w-40 cursor-pointer appearance-none rounded-full bg-secondary accent-primary"
                value={Math.min(borderRadius ?? 0, 20)}
                onChange={(e) => update({borderRadius: Setting.myParseInt(e.target.value)})}
              />
              <span className="text-xs text-muted-foreground">{borderRadius}px</span>
            </div>
          </FormRow>
          <FormRow label={i18next.t("theme:Is compact")}>
            <Switch checked={!!current.isCompact} onCheckedChange={(v) => update({isCompact: v})} />
          </FormRow>
        </>
      ) : null}
    </div>
  );
}
