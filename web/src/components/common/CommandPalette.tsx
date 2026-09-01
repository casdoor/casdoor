import * as React from "react";
import i18next from "i18next";
import {AppWindow, Building2, CornerDownLeft, Search, User} from "lucide-react";
import {useNavigate} from "react-router-dom";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import {useAccount} from "@/hooks/use-account";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as UserBackend from "@/backend/UserBackend";
import {getNavGroups} from "@/lib/nav";
import * as Setting from "@/lib/setting";

interface Hit {
  key: string;
  to: string;
  label: string;
  hint?: string;
  icon: React.ReactNode;
  group: string;
}

/** long enough that a stray keystroke does not cost three list queries */
const DEBOUNCE_MS = 250;
const PER_KIND = 5;

/**
 * Searches the objects a console admin actually navigates by. The Casdoor list
 * APIs take one `field`/`value` pair, so each kind is one query on its name — the
 * same query a list page's column search sends.
 */
function useObjectSearch(term: string, open: boolean) {
  const {account} = useAccount();
  const [hits, setHits] = React.useState<Hit[]>([]);
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    const query = term.trim();
    if (!open || !account || query.length < 2) {
      setHits([]);
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    const timer = setTimeout(() => {
      const organization = Setting.getRequestOrganization(account);
      const rows = (settled: PromiseSettledResult<any>) =>
        settled.status === "fulfilled" && settled.value?.status === "ok" && Array.isArray(settled.value.data)
          ? settled.value.data
          : [];

      Promise.allSettled([
        UserBackend.getUsers(organization, 1, PER_KIND, "name", query),
        ApplicationBackend.getApplications("admin", 1, PER_KIND, "name", query),
        Setting.isAdminUser(account)
          ? OrganizationBackend.getOrganizations("admin", "", 1, PER_KIND, "name", query)
          : Promise.resolve(null),
      ]).then(([users, applications, organizations]) => {
        if (cancelled) {
          return;
        }
        setHits([
          ...rows(users).map((user: any) => ({
            key: `user/${user.owner}/${user.name}`,
            to: `/users/${user.owner}/${user.name}`,
            label: user.displayName || user.name,
            hint: `${user.owner}/${user.name}`,
            icon: <User />,
            group: i18next.t("general:Users"),
          })),
          ...rows(applications).map((application: any) => ({
            key: `application/${application.name}`,
            to: `/applications/${application.organization}/${application.name}`,
            label: application.displayName || application.name,
            hint: application.name,
            icon: <AppWindow />,
            group: i18next.t("general:Applications"),
          })),
          ...rows(organizations).map((item: any) => ({
            key: `organization/${item.name}`,
            to: `/organizations/${item.name}`,
            label: item.displayName || item.name,
            hint: item.name,
            icon: <Building2 />,
            group: i18next.t("general:Organizations"),
          })),
        ]);
        setLoading(false);
      });
    }, DEBOUNCE_MS);

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [term, open, account]);

  return {hits, loading};
}

/** Owns the ⌘K / Ctrl+K shortcut, so the layout can hand `open` to both the palette and its trigger. */
export function useCommandPalette() {
  const [open, setOpen] = React.useState(false);

  React.useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key.toLowerCase() === "k" && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        setOpen((previous) => !previous);
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, []);

  return {open, setOpen};
}

function groupBy<T>(items: T[], key: (item: T) => string) {
  return items.reduce<Record<string, T[]>>((acc, item) => {
    (acc[key(item)] ??= []).push(item);
    return acc;
  }, {});
}

/**
 * Every console page by name, and the users, applications and organizations
 * behind them by name — which is the only way to reach one object out of five
 * thousand without paging through a list.
 */
export function CommandPalette({open, onOpenChange}: {open: boolean; onOpenChange: (open: boolean) => void}) {
  const navigate = useNavigate();
  const {account} = useAccount();
  const [term, setTerm] = React.useState("");
  const {hits, loading} = useObjectSearch(term, open);

  // recomputed per render: the labels come from i18next, which changes language
  // without changing this component's inputs
  const pages = getNavGroups(account).flatMap((group) =>
    group.items.filter((item) => !item.href).map((item) => ({...item, group: group.label})),
  );

  const go = (to: string) => {
    onOpenChange(false);
    setTerm("");
    navigate(to);
  };

  return (
    <CommandDialog
      open={open}
      onOpenChange={(next) => {
        onOpenChange(next);
        if (!next) {
          setTerm("");
        }
      }}
    >
      <CommandInput
        value={term}
        onValueChange={setTerm}
        placeholder={i18next.t("general:Please input your search")}
      />
      <CommandList>
        <CommandEmpty>{loading ? "..." : i18next.t("general:No data")}</CommandEmpty>

        {/* the object hits are already filtered by the server, so each one carries
            a `value` cmdk cannot narrow further — the search term itself */}
        {Object.entries(groupBy(hits, (hit) => hit.group)).map(([group, groupHits]) => (
          <CommandGroup key={group} heading={group}>
            {groupHits.map((hit) => (
              <CommandItem key={hit.key} value={`${term} ${hit.key}`} onSelect={() => go(hit.to)}>
                {hit.icon}
                <span className="flex-1 truncate">{hit.label}</span>
                {hit.hint && hit.hint !== hit.label ? (
                  <span className="truncate text-xs text-muted-foreground">{hit.hint}</span>
                ) : null}
              </CommandItem>
            ))}
          </CommandGroup>
        ))}
        {hits.length > 0 ? <CommandSeparator /> : null}

        {Object.entries(groupBy(pages, (page) => page.group)).map(([group, groupPages]) => (
          <CommandGroup key={group} heading={group}>
            {groupPages.map((page) => (
              <CommandItem key={page.key} value={`${group} ${page.label}`} onSelect={() => go(page.key)}>
                <CornerDownLeft className="opacity-50" />
                {page.label}
              </CommandItem>
            ))}
          </CommandGroup>
        ))}
      </CommandList>
    </CommandDialog>
  );
}

/** The header's search box, which only ever opens the palette. */
export function CommandPaletteTrigger({onOpen}: {onOpen: () => void}) {
  const isMac = typeof navigator !== "undefined" && /Mac|iPhone|iPad/.test(navigator.platform);
  return (
    <button
      type="button"
      onClick={onOpen}
      className="hidden h-8 items-center gap-2 rounded-md border border-input bg-background px-2.5 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground md:flex"
    >
      <Search className="h-3.5 w-3.5" />
      <span>{i18next.t("general:Search")}</span>
      <kbd className="ml-1 rounded border bg-muted px-1.5 py-0.5 text-[10px] font-medium leading-none">
        {isMac ? "⌘" : "Ctrl+"}K
      </kbd>
    </button>
  );
}
