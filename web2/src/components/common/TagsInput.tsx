import * as React from "react";
import {X} from "lucide-react";
import {Badge} from "@/components/ui/badge";
import {Input} from "@/components/ui/input";
import {cn} from "@/lib/utils";

interface TagsInputProps {
  value: string[];
  onChange: (value: string[]) => void;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
}

/** Free-form list of strings, replacing antd's `<Select mode="tags" />`. */
export function TagsInput({value, onChange, placeholder, disabled, className}: TagsInputProps) {
  const [draft, setDraft] = React.useState("");
  const items = value ?? [];

  const commit = () => {
    const next = draft.trim();
    if (next !== "" && !items.includes(next)) {
      onChange([...items, next]);
    }
    setDraft("");
  };

  return (
    <div
      className={cn(
        "flex min-h-9 w-full flex-wrap items-center gap-1 rounded-md border border-input bg-transparent px-2 py-1 shadow-sm",
        disabled && "opacity-50",
        className,
      )}
    >
      {items.map((item) => (
        <Badge key={item} variant="secondary" className="gap-1 py-0.5">
          {item}
          <button
            type="button"
            disabled={disabled}
            className="opacity-60 hover:opacity-100"
            onClick={() => onChange(items.filter((i) => i !== item))}
          >
            <X className="h-3 w-3" />
          </button>
        </Badge>
      ))}
      <Input
        disabled={disabled}
        value={draft}
        placeholder={items.length === 0 ? placeholder : ""}
        className="h-6 flex-1 border-0 px-1 shadow-none focus-visible:ring-0"
        onChange={(e) => setDraft(e.target.value)}
        onBlur={commit}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === ",") {
            e.preventDefault();
            commit();
          } else if (e.key === "Backspace" && draft === "" && items.length > 0) {
            onChange(items.slice(0, -1));
          }
        }}
      />
    </div>
  );
}
