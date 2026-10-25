import * as React from "react";
import {cn} from "@/lib/utils";

// object/verification.go always sends six digits, and TOTP codes are six digits too
export const VerificationCodeLength = 6;

interface CodeInputProps {
  value: string;
  onChange: (value: string) => void;
  /** the number of digits, one box each */
  length?: number;
  /** called once the last box is typed, after the parent has re-rendered with the full code */
  onComplete?: () => void;
  id?: string;
  autoFocus?: boolean;
  disabled?: boolean;
  className?: string;
}

/**
 * A verification code typed one digit per box: focus moves to the next box as a digit
 * is typed, Backspace goes back, and a pasted or autofilled code is spread over the boxes.
 * The digits stay contiguous, so a box after the first empty one cannot be focused.
 */
export function CodeInput({value, onChange, length = VerificationCodeLength, onComplete, id, autoFocus, disabled, className}: CodeInputProps) {
  const refs = React.useRef<(HTMLInputElement | null)[]>([]);
  // the box focused right after a change sees the new value before the parent re-renders
  const latest = React.useRef(value);
  latest.current = value;
  const userEdit = React.useRef(false);
  const digits = Array.from({length}, (_, i) => value[i] ?? "");

  const focusBox = (index: number) => {
    const box = refs.current[Math.max(0, Math.min(index, length - 1))];
    box?.focus();
    box?.select();
  };

  React.useEffect(() => {
    const byUser = userEdit.current;
    userEdit.current = false;
    if (byUser && value.length === length) {
      onComplete?.();
    }
    // a code cleared by the parent, e.g. after a wrong one, is typed again from its first empty box
    if (refs.current.indexOf(document.activeElement as HTMLInputElement) > value.length) {
      focusBox(value.length);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  const setValue = (next: string) => {
    userEdit.current = true;
    latest.current = next.replace(/\D/g, "").slice(0, length);
    onChange(latest.current);
  };

  const fill = (index: number, text: string) => {
    const typed = text.replace(/\D/g, "");
    if (typed === "") {
      return;
    }
    const next = (value.slice(0, index) + typed + value.slice(index + typed.length)).slice(0, length);
    setValue(next);
    focusBox(Math.min(index + typed.length, length - 1));
  };

  const remove = (index: number) => {
    setValue(value.slice(0, index) + value.slice(index + 1));
    focusBox(index);
  };

  return (
    <div className={cn("grid gap-2 sm:gap-3", className)} style={{gridTemplateColumns: `repeat(${length}, minmax(0, 1fr))`}}>
      {digits.map((digit, i) => (
        <input
          key={i}
          ref={(el) => {
            refs.current[i] = el;
          }}
          id={i === 0 ? id : undefined}
          className="verification-code-box h-12 w-full min-w-0 rounded-md border border-input bg-transparent text-center text-lg font-medium shadow-sm transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          type="text"
          inputMode="numeric"
          pattern="[0-9]*"
          autoComplete={i === 0 ? "one-time-code" : "off"}
          autoFocus={autoFocus && i === 0}
          disabled={disabled}
          aria-label={`${i + 1} / ${length}`}
          value={digit}
          onFocus={(e) => {
            if (i > latest.current.length) {
              focusBox(latest.current.length);
            } else {
              e.target.select();
            }
          }}
          onChange={(e) => {
            let typed = e.target.value.replace(/\D/g, "");
            if (e.target.value === "") {
              remove(i);
              return;
            }
            // a digit typed next to the one already in the box, rather than over it
            if (digit !== "" && typed.length === 2) {
              typed = typed.startsWith(digit) ? typed.slice(1) : typed.slice(0, 1);
            }
            fill(i, typed);
          }}
          onPaste={(e) => {
            e.preventDefault();
            fill(i, e.clipboardData.getData("text"));
          }}
          onKeyDown={(e) => {
            if (e.key === "Backspace") {
              e.preventDefault();
              if (digit !== "") {
                remove(i);
              } else if (i > 0) {
                remove(i - 1);
              }
            } else if (e.key === "ArrowLeft") {
              e.preventDefault();
              focusBox(i - 1);
            } else if (e.key === "ArrowRight") {
              e.preventDefault();
              focusBox(Math.min(i + 1, value.length));
            }
          }}
        />
      ))}
    </div>
  );
}
