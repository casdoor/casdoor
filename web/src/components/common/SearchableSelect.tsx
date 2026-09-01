import * as React from "react";
import i18next from "i18next";
import {Check, ChevronsUpDown} from "lucide-react";
import {Button} from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {Popover, PopoverContent, PopoverTrigger} from "@/components/ui/popover";
import {cn} from "@/lib/utils";

export interface SearchableOption {
  value: string;
  label: React.ReactNode;
  keywords?: string;
  disabled?: boolean;
}

interface SearchableSelectProps {
  value?: string | null;
  onChange: (value: string) => void;
  options: SearchableOption[];
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  contentClassName?: string;
  /** keep a value that is not part of options (e.g. loaded from the backend) */
  allowUnknownValue?: boolean;
}

/**
 * Single-select combobox with a filter box — the shadcn equivalent of antd's
 * `<Select showSearch />`, used wherever the option list can get long.
 */
export function SearchableSelect({
  value,
  onChange,
  options,
  placeholder,
  disabled,
  className,
  contentClassName,
  allowUnknownValue = true,
}: SearchableSelectProps) {
  const [open, setOpen] = React.useState(false);

  const items = React.useMemo(() => {
    if (allowUnknownValue && value && !options.some((o) => o.value === value)) {
      return [{value, label: value}, ...options];
    }
    return options;
  }, [options, value, allowUnknownValue]);

  const selected = items.find((o) => o.value === value);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="outline"
          role="combobox"
          disabled={disabled}
          className={cn("w-full justify-between font-normal hover:bg-background", className)}
        >
          <span className={cn("truncate", selected === undefined && "text-muted-foreground")}>
            {selected?.label ?? placeholder ?? ""}
          </span>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className={cn("w-[--radix-popover-trigger-width] p-0", contentClassName)} align="start">
        <Command
          filter={(itemValue, search) => {
            const option = items.find((o) => o.value === itemValue);
            const haystack = `${itemValue} ${option?.keywords ?? ""} ${
              typeof option?.label === "string" ? option.label : ""
            }`.toLowerCase();
            return haystack.includes(search.toLowerCase()) ? 1 : 0;
          }}
        >
          <CommandInput placeholder={i18next.t("general:Search")} />
          <CommandList>
            <CommandEmpty>{i18next.t("general:No data")}</CommandEmpty>
            <CommandGroup>
              {items.map((option) => (
                <CommandItem
                  key={option.value}
                  value={option.value}
                  disabled={option.disabled}
                  onSelect={() => {
                    onChange(option.value);
                    setOpen(false);
                  }}
                >
                  <Check className={cn("mr-2 h-4 w-4", value === option.value ? "opacity-100" : "opacity-0")} />
                  {option.label}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
