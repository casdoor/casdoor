import i18next from "i18next";
import {Monitor, Moon, Sun} from "lucide-react";
import {Button} from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {useTheme} from "@/hooks/use-theme";
import {cn} from "@/lib/utils";

export function ThemeToggle({className}: {className?: string}) {
  const {theme, resolvedTheme, setTheme} = useTheme();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="iconSm" className={className} aria-label={i18next.t("theme:Theme")}>
          {resolvedTheme === "dark" ? <Moon /> : <Sun />}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem className={cn(theme === "light" && "font-semibold")} onSelect={() => setTheme("light")}>
          <Sun />
          {i18next.t("theme:Light")}
        </DropdownMenuItem>
        <DropdownMenuItem className={cn(theme === "dark" && "font-semibold")} onSelect={() => setTheme("dark")}>
          <Moon />
          {i18next.t("theme:Dark")}
        </DropdownMenuItem>
        <DropdownMenuItem className={cn(theme === "system" && "font-semibold")} onSelect={() => setTheme("system")}>
          <Monitor />
          {i18next.t("theme:System")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
