import {Languages} from "lucide-react";
import {Button} from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {useLanguage} from "@/hooks/use-language";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

interface LanguageSelectProps {
  /** restrict the list, as the organization setting does */
  languages?: string[];
  className?: string;
}

export function LanguageSelect({languages, className}: LanguageSelectProps) {
  const current = useLanguage();

  const items = (Setting.Countries as any[]).filter(
    (country) => !languages || languages.length === 0 || languages.includes(country.key),
  );

  if (items.length === 0) {
    return null;
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="iconSm" className={className} aria-label="Language">
          <Languages />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="max-h-80 overflow-y-auto">
        {items.map((country) => (
          <DropdownMenuItem
            key={country.key}
            className={cn(current === country.key && "font-semibold")}
            onSelect={() => Setting.setLanguage(country.key)}
          >
            <img
              src={`${Setting.StaticBaseUrl}/flag-icons/${country.country}.svg`}
              alt={country.alt}
              className="h-4 w-5 rounded-[2px] object-cover"
            />
            {country.label}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default LanguageSelect;
