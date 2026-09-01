import * as React from "react";
import i18next from "i18next";
import {Check, X} from "lucide-react";
import {cn} from "@/lib/utils";

/**
 * The organization `passwordOptions` checkers, ported verbatim from
 * `web/src/common/PasswordChecker.js`. The backend re-validates on every write, so
 * these only exist to tell the user why a password will be rejected.
 */
const checkers: Record<string, (password: string) => string> = {
  AtLeast6: (password) =>
    password.length < 6 ? i18next.t("user:The password must have at least 6 characters") : "",
  AtLeast8: (password) =>
    password.length < 8 ? i18next.t("user:The password must have at least 8 characters") : "",
  Aa123: (password) =>
    /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9]).+$/.test(password)
      ? ""
      : i18next.t("user:The password must contain at least one uppercase letter, one lowercase letter and one digit"),
  SpecialChar: (password) =>
    /^(?=.*[!-/:-@[-`{-~]).+$/.test(password)
      ? ""
      : i18next.t("user:The password must contain at least one special character"),
  NoRepeat: (password) =>
    /(.)\1+/.test(password) ? i18next.t("user:The password must not contain any repeated characters") : "",
};

function getOptionDescription(option: string): string {
  switch (option) {
  case "AtLeast6": return i18next.t("user:The password must have at least 6 characters");
  case "AtLeast8": return i18next.t("user:The password must have at least 8 characters");
  case "Aa123": return i18next.t("user:The password must contain at least one uppercase letter, one lowercase letter and one digit");
  case "SpecialChar": return i18next.t("user:The password must contain at least one special character");
  case "NoRepeat": return i18next.t("user:The password must not contain any repeated characters");
  default: return option;
  }
}

/** "" when the password satisfies every option, otherwise the first failure message. */
export function checkPasswordComplexity(password: string, options?: string[] | null): string {
  if (!password?.length) {
    return i18next.t("login:Please input your password!");
  }

  if (!options || options.length === 0) {
    return "";
  }

  for (const option of options) {
    const checkerFunc = checkers[option];
    if (checkerFunc) {
      const errorMsg = checkerFunc(password);
      if (errorMsg !== "") {
        return errorMsg;
      }
    }
  }
  return "";
}

/** Whether the requirement list is worth showing: typed something, not yet valid. */
export function getPasswordPopoverOpen(password: string, options?: string[] | null): boolean {
  return password.length > 0 && checkPasswordComplexity(password, options) !== "";
}

/**
 * The shadcn stand-in for antd's `renderPasswordPopover`: one line per option with a
 * tick or a cross. Rendered inline rather than in a popover, which behaves better on
 * mobile and inside a dialog.
 */
export function PasswordRequirements({
  options,
  password,
  className,
}: {
  options?: string[] | null;
  password: string;
  className?: string;
}) {
  if (!options || options.length === 0) {
    return null;
  }

  return (
    <div className={cn("space-y-1 text-xs", className)}>
      {options.map((option) => {
        const ok = checkers[option] ? checkers[option](password) === "" : true;
        return (
          <div key={option} className={cn("flex items-start gap-1.5", ok ? "text-success" : "text-muted-foreground")}>
            {ok ? (
              <Check className="mt-0.5 h-3.5 w-3.5 shrink-0" />
            ) : (
              <X className="mt-0.5 h-3.5 w-3.5 shrink-0 text-destructive" />
            )}
            <span>{getOptionDescription(option)}</span>
          </div>
        );
      })}
    </div>
  );
}
