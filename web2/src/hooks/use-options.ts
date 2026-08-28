import * as React from "react";
import type {SearchableOption} from "@/components/common/SearchableSelect";
import * as AdapterBackend from "@/backend/AdapterBackend";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as CertBackend from "@/backend/CertBackend";
import * as GroupBackend from "@/backend/GroupBackend";
import * as ModelBackend from "@/backend/ModelBackend";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as PlanBackend from "@/backend/PlanBackend";
import * as ProductBackend from "@/backend/ProductBackend";
import * as PricingBackend from "@/backend/PricingBackend";
import * as ProviderBackend from "@/backend/ProviderBackend";
import * as RuleBackend from "@/backend/RuleBackend";
import * as RoleBackend from "@/backend/RoleBackend";
import * as UserBackend from "@/backend/UserBackend";

// The option lists feed comboboxes that filter client-side, so one large page is
// requested instead of paginating like the antd PaginateSelect did.
const PAGE_SIZE = 1000;

function useList<T = any>(loader: () => Promise<any> | null, deps: React.DependencyList): T[] {
  const [items, setItems] = React.useState<T[]>([]);
  const loaderRef = React.useRef(loader);
  loaderRef.current = loader;

  React.useEffect(() => {
    let cancelled = false;
    const promise = loaderRef.current();
    if (!promise) {
      setItems([]);
      return;
    }
    promise
      .then((res: any) => {
        if (!cancelled && res?.status === "ok") {
          setItems(res.data ?? []);
        }
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  return items;
}

/** value: "owner/name", label: "displayName (owner/name)" — same as Setting.getDisplayNameOption. */
function toIdOptions(items: any[]): SearchableOption[] {
  return items.map((item) => {
    const id = `${item.owner}/${item.name}`;
    return {
      value: id,
      label: item.displayName ? `${item.displayName} (${id})` : id,
      keywords: `${item.name} ${item.displayName ?? ""}`,
    };
  });
}

/** value: "name" — for the fields that store a bare name inside one organization. */
function toNameOptions(items: any[]): SearchableOption[] {
  return items.map((item) => ({
    value: item.name,
    label: item.displayName ? `${item.displayName} (${item.name})` : item.name,
    keywords: `${item.name} ${item.displayName ?? ""}`,
  }));
}

export function useOrganizationOptions(): SearchableOption[] {
  const items = useList(() => OrganizationBackend.getOrganizations("admin", "", 1, PAGE_SIZE), []);
  return React.useMemo(() => toNameOptions(items), [items]);
}

export function useUserOptions(organizationName: string): SearchableOption[] {
  const items = useList(() => (organizationName ? UserBackend.getUsers(organizationName, 1, PAGE_SIZE) : null), [organizationName]);
  return React.useMemo(() => toIdOptions(items), [items]);
}

export function useUserNameOptions(organizationName: string): SearchableOption[] {
  const items = useList(() => (organizationName ? UserBackend.getUsers(organizationName, 1, PAGE_SIZE) : null), [organizationName]);
  return React.useMemo(() => toNameOptions(items), [items]);
}

export function useGroupOptions(organizationName: string): SearchableOption[] {
  const items = useList(
    () => (organizationName ? GroupBackend.getGroups(organizationName, false, 1, PAGE_SIZE) : null),
    [organizationName],
  );
  return React.useMemo(() => toIdOptions(items), [items]);
}

export function useRoleOptions(organizationName: string, exclude?: string): SearchableOption[] {
  const items = useList(() => (organizationName ? RoleBackend.getRoles(organizationName, 1, PAGE_SIZE) : null), [organizationName]);
  return React.useMemo(() => toIdOptions(items).filter((option) => option.value !== exclude), [items, exclude]);
}

export function useRoleNameOptions(organizationName: string): SearchableOption[] {
  const items = useList(() => (organizationName ? RoleBackend.getRoles(organizationName, 1, PAGE_SIZE) : null), [organizationName]);
  return React.useMemo(() => toNameOptions(items), [items]);
}

export function useModelOptions(organizationName: string): SearchableOption[] {
  const items = useList(() => (organizationName ? ModelBackend.getModels(organizationName, 1, PAGE_SIZE) : null), [organizationName]);
  return React.useMemo(() => toNameOptions(items), [items]);
}

export function useAdapterOptions(organizationName: string): SearchableOption[] {
  const items = useList(
    () => (organizationName ? AdapterBackend.getAdapters(organizationName, 1, PAGE_SIZE) : null),
    [organizationName],
  );
  return React.useMemo(() => toNameOptions(items), [items]);
}

export function useApplicationOptions(organizationName: string): SearchableOption[] {
  const items = useList(
    () => (organizationName ? ApplicationBackend.getApplicationsByOrganization("admin", organizationName, 1, PAGE_SIZE) : null),
    [organizationName],
  );
  return React.useMemo(() => toNameOptions(items), [items]);
}

export function useProviderOptions(organizationName: string, category?: string): SearchableOption[] {
  const items = useList(
    () => (organizationName ? ProviderBackend.getProviders(organizationName, 1, PAGE_SIZE) : null),
    [organizationName],
  );
  return React.useMemo(
    () => toNameOptions(category ? items.filter((item: any) => item.category === category) : items),
    [items, category],
  );
}

export function useCertOptions(organizationName: string, scope?: string): SearchableOption[] {
  const items = useList(() => (organizationName ? CertBackend.getCerts(organizationName, 1, PAGE_SIZE) : null), [organizationName]);
  return React.useMemo(
    () => toNameOptions(scope ? items.filter((item: any) => item.scope === scope) : items),
    [items, scope],
  );
}

export function usePlanOptions(organizationName: string): SearchableOption[] {
  const items = useList(() => (organizationName ? PlanBackend.getPlans(organizationName, 1, PAGE_SIZE) : null), [organizationName]);
  return React.useMemo(() => toNameOptions(items), [items]);
}

export function useProductOptions(organizationName: string): SearchableOption[] {
  const items = useList(
    () => (organizationName ? ProductBackend.getProducts(organizationName, 1, PAGE_SIZE) : null),
    [organizationName],
  );
  return React.useMemo(() => toNameOptions(items), [items]);
}

export function usePricingOptions(organizationName: string): SearchableOption[] {
  const items = useList(
    () => (organizationName ? PricingBackend.getPricings(organizationName, 1, PAGE_SIZE) : null),
    [organizationName],
  );
  return React.useMemo(() => toNameOptions(items), [items]);
}

export function useRuleOptions(organizationName: string): SearchableOption[] {
  // the rules API paginates but takes no filter arguments
  const items = useList(() => (organizationName ? RuleBackend.getRules(organizationName, 1, PAGE_SIZE) : null), [organizationName]);
  return React.useMemo(() => toNameOptions(items), [items]);
}
