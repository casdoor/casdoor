import * as React from "react";
import i18next from "i18next";
import {cn} from "@/lib/utils";

// object/verification.go always sends six digits, and TOTP codes are six digits too
export const VerificationCodeLength = 6;

interface CodeInputProps {
  value: string;
  onChange: (value: string) => void;
  /** the number of digits, one box each */
  length?: number;
  /** called once the last digit is typed, pasted or autofilled, after the parent has re-rendered with the full code */
  onComplete?: () => void;
  id?: string;
  autoFocus?: boolean;
  disabled?: boolean;
  className?: string;
}

// the boxes only draw the digits, so password managers must not take them for fields of their own
const ignoredByPasswordManagers = {
  "data-1p-ignore": "true",
  "data-lpignore": "true",
  "data-bwignore": "true",
  "data-protonpass-ignore": "true",
  "data-form-type": "other",
};

/**
 * A verification code shown one digit per box. The boxes are drawn under one real
 * input, so a password manager or the phone's SMS autofill sees an ordinary
 * one-time-code field and fills the whole code. Typing, Backspace, the arrow keys,
 * paste and a click on a box all go to that input; the digits stay contiguous, so
 * a box after the first empty one cannot be selected.
 */
export function CodeInput({value, onChange, length = VerificationCodeLength, onComplete, id, autoFocus, disabled, className}: CodeInputProps) {
  const inputRef = React.useRef<HTMLInputElement>(null);
  const userEdit = React.useRef(false);
  // where the caret goes once the parent has re-rendered with the new code
  const pendingCaret = React.useRef<number | null>(null);
  const [focused, setFocused] = React.useState(false);
  const [caret, setCaret] = React.useState(0);
  const digits = Array.from({length}, (_, i) => value[i] ?? "");

  // a filled box is selected so the next digit replaces it; past the last digit the caret
  // waits in the first empty box, or on the last box of a full code
  const select = (index: number, code = value) => {
    const input = inputRef.current;
    if (!input) {
      return;
    }
    const start = Math.max(0, Math.min(index, code.length, length - 1));
    input.setSelectionRange(start, start < code.length ? start + 1 : start);
    setCaret(start);
  };

  React.useLayoutEffect(() => {
    // a code changed by the parent, e.g. cleared after a wrong one, is typed again from its end
    if (document.activeElement === inputRef.current) {
      select(pendingCaret.current ?? value.length);
    }
    pendingCaret.current = null;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  React.useEffect(() => {
    const byUser = userEdit.current;
    userEdit.current = false;
    if (byUser && value.length === length) {
      onComplete?.();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  const setValue = (next: string, nextCaret: number) => {
    if (next === value) {
      // nothing changed (a letter was typed): React puts the old value back, then the caret
      requestAnimationFrame(() => select(nextCaret));
      return;
    }
    userEdit.current = true;
    pendingCaret.current = nextCaret;
    onChange(next);
  };

  return (
    <div className={cn("relative grid gap-2 sm:gap-3", className)} style={{gridTemplateColumns: `repeat(${length}, minmax(0, 1fr))`}}>
      {digits.map((digit, i) => {
        const active = focused && !disabled && i === caret;
        return (
          <div key={i} className="relative">
            <input
              {...ignoredByPasswordManagers}
              className="verification-code-box pointer-events-none h-12 w-full min-w-0 rounded-md border border-input bg-transparent text-center text-lg font-medium shadow-sm transition-colors disabled:cursor-not-allowed disabled:opacity-50"
              // inline, so a theme's own input border cannot hide which box is active
              style={active ? {borderColor: "hsl(var(--ring))", boxShadow: "0 0 0 1px hsl(var(--ring))"} : undefined}
              type="text"
              readOnly
              tabIndex={-1}
              aria-hidden="true"
              autoComplete="off"
              disabled={disabled}
              value={digit}
            />
            {active && digit === "" ? (
              <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
                <div className="h-5 w-px animate-pulse bg-foreground" />
              </div>
            ) : null}
          </div>
        );
      })}
      <input
        ref={inputRef}
        id={id}
        name="otp"
        // themes style every input of the sign-in form; inline styles keep this one invisible over the boxes
        className="selection:bg-transparent disabled:cursor-not-allowed"
        style={{
          position: "absolute",
          inset: 0,
          zIndex: 1,
          width: "100%",
          height: "100%",
          margin: 0,
          padding: 0,
          border: 0,
          borderRadius: 0,
          background: "transparent",
          boxShadow: "none",
          outline: "none",
          color: "transparent",
          caretColor: "transparent",
          cursor: "text",
        }}
        type="text"
        inputMode="numeric"
        pattern="[0-9]*"
        autoComplete="one-time-code"
        maxLength={length}
        autoFocus={autoFocus}
        disabled={disabled}
        aria-label={id ? undefined : i18next.t("login:Verification code")}
        value={value}
        onFocus={() => {
          setFocused(true);
          select(value.length);
        }}
        onBlur={() => setFocused(false)}
        onSelect={(e) => setCaret(Math.min(e.currentTarget.selectionStart ?? 0, length - 1))}
        onMouseDown={(e) => {
          // a click selects the box under the pointer instead of placing a hidden caret
          e.preventDefault();
          if (disabled) {
            return;
          }
          const rect = e.currentTarget.getBoundingClientRect();
          e.currentTarget.focus();
          select(Math.floor(((e.clientX - rect.left) / rect.width) * length));
        }}
        onChange={(e) => {
          const raw = e.target.value;
          const before = raw.slice(0, e.target.selectionStart ?? raw.length).replace(/\D/g, "").length;
          setValue(raw.replace(/\D/g, "").slice(0, length), before);
        }}
        onPaste={(e) => {
          // the native paste would cut "123 456" at the length limit before the spaces are dropped
          e.preventDefault();
          const pasted = e.clipboardData.getData("text").replace(/\D/g, "");
          if (pasted === "") {
            return;
          }
          if (pasted.length >= length) {
            setValue(pasted.slice(0, length), length);
            return;
          }
          const start = Math.min(e.currentTarget.selectionStart ?? value.length, value.length);
          setValue((value.slice(0, start) + pasted + value.slice(start + pasted.length)).slice(0, length), start + pasted.length);
        }}
        onKeyDown={(e) => {
          if (e.key === "ArrowLeft" || e.key === "ArrowRight") {
            e.preventDefault();
            select(caret + (e.key === "ArrowLeft" ? -1 : 1));
          }
        }}
      />
    </div>
  );
}
