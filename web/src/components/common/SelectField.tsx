import * as React from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {cn} from "@/lib/utils";

export interface Option {
  id: string | number;
  name: React.ReactNode;
  disabled?: boolean;
}

interface SelectFieldProps {
  value: string | number | undefined | null;
  onChange: (value: string) => void;
  options: Option[];
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  /** rendered when the current value is not part of the options */
  allowUnknownValue?: boolean;
}

/**
 * Drop-in replacement for the antd `<Select options={[{id, name}]} />` pattern
 * that the Casdoor edit pages use everywhere.
 */
export function SelectField({
  value,
  onChange,
  options,
  placeholder,
  disabled,
  className,
  allowUnknownValue = true,
}: SelectFieldProps) {
  const stringValue = value === undefined || value === null || value === "" ? undefined : String(value);
  const known = options.some((o) => String(o.id) === stringValue);
  const items = !known && stringValue && allowUnknownValue ? [{id: stringValue, name: stringValue}, ...options] : options;

  return (
    <Select value={stringValue} onValueChange={onChange} disabled={disabled}>
      <SelectTrigger className={cn("w-full", className)}>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {items.map((option) => (
          <SelectItem key={String(option.id)} value={String(option.id)} disabled={option.disabled}>
            {option.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
