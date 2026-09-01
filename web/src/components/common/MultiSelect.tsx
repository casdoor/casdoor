import * as React from "react";
import i18next from "i18next";
import {Check, ChevronsUpDown, X} from "lucide-react";
import {Badge} from "@/components/ui/badge";
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

export interface MultiSelectOption {
  value: string;
  label: React.ReactNode;
  /** plain-text used for filtering */
  keywords?: string;
  disabled?: boolean;
}

interface MultiSelectProps {
  value: string[];
  onChange: (value: string[]) => void;
  options: MultiSelectOption[];
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  /** allow values that are not part of `options` (free tagging) */
  creatable?: boolean;
  maxDisplay?: number;
}

/** Replacement for antd's `<Select mode="multiple" />`. */
export function MultiSelect({
  value,
  onChange,
  options,
  placeholder,
  disabled,
  className,
  creatable = false,
  maxDisplay = 20,
}: MultiSelectProps) {
  const [open, setOpen] = React.useState(false);
  const [search, setSearch] = React.useState("");
  const selected = value ?? [];

  const labelOf = React.useCallback(
    (v: string) => options.find((o) => o.value === v)?.label ?? v,
    [options],
  );

  const toggle = (v: string) => {
    if (selected.includes(v)) {
      onChange(selected.filter((item) => item !== v));
    } else {
      onChange([...selected, v]);
    }
  };

  const canCreate =
    creatable && search.trim() !== "" && !options.some((o) => o.value === search.trim()) && !selected.includes(search.trim());

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="outline"
          role="combobox"
          disabled={disabled}
          className={cn(
            "h-auto min-h-9 w-full justify-between px-3 py-1.5 font-normal hover:bg-background",
            className,
          )}
        >
          <span className="flex flex-1 flex-wrap gap-1">
            {selected.length === 0 ? (
              <span className="text-muted-foreground">{placeholder ?? ""}</span>
            ) : (
              <>
                {selected.slice(0, maxDisplay).map((v) => (
                  <Badge key={v} variant="secondary" className="gap-1 py-0.5">
                    {labelOf(v)}
                    <span
                      role="button"
                      tabIndex={-1}
                      className="rounded-sm opacity-60 hover:opacity-100"
                      onClick={(e) => {
                        e.stopPropagation();
                        onChange(selected.filter((item) => item !== v));
                      }}
                    >
                      <X className="h-3 w-3" />
                    </span>
                  </Badge>
                ))}
                {selected.length > maxDisplay && (
                  <Badge variant="outline">+{selected.length - maxDisplay}</Badge>
                )}
              </>
            )}
          </span>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[--radix-popover-trigger-width] p-0" align="start">
        <Command
          filter={(itemValue, searchValue) => {
            const option = options.find((o) => o.value === itemValue);
            const haystack = `${itemValue} ${option?.keywords ?? ""}`.toLowerCase();
            return haystack.includes(searchValue.toLowerCase()) ? 1 : 0;
          }}
        >
          <CommandInput placeholder={i18next.t("general:Search")} value={search} onValueChange={setSearch} />
          <CommandList>
            <CommandEmpty>{i18next.t("general:No data")}</CommandEmpty>
            <CommandGroup>
              {canCreate && (
                <CommandItem
                  value={search.trim()}
                  onSelect={() => {
                    toggle(search.trim());
                    setSearch("");
                  }}
                >
                  <Check className="mr-2 h-4 w-4 opacity-0" />
                  {search.trim()}
                </CommandItem>
              )}
              {options.map((option) => (
                <CommandItem
                  key={option.value}
                  value={option.value}
                  disabled={option.disabled}
                  onSelect={() => toggle(option.value)}
                >
                  <Check
                    className={cn("mr-2 h-4 w-4", selected.includes(option.value) ? "opacity-100" : "opacity-0")}
                  />
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
