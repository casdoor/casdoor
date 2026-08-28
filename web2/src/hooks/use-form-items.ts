import * as React from "react";
import {useAccount} from "@/hooks/use-account";
import * as FormBackend from "@/backend/FormBackend";

/**
 * Loads the Form that customizes a list page's columns, the same lookup the antd
 * BaseListPage did: an account with a tag first looks for "<type>-tag-<tag>" and
 * falls back to the plain "<type>" form.
 */
export function useFormItems(formType: string | undefined): any[] {
  const {account} = useAccount();
  const owner = account?.owner;
  const tag = account?.tag ?? "";
  const [formItems, setFormItems] = React.useState<any[]>([]);

  React.useEffect(() => {
    if (!formType || !owner) {
      setFormItems([]);
      return;
    }

    let cancelled = false;
    const load = (formName: string) => FormBackend.getForm(owner, formName);

    const apply = (res: any) => {
      if (!cancelled) {
        setFormItems(res?.status === "ok" && res.data ? res.data.formItems ?? [] : []);
      }
      return res?.status === "ok" && res.data;
    };

    if (tag !== "") {
      load(`${formType}-tag-${tag}`)
        .then((res: any) => {
          if (!apply(res)) {
            return load(formType).then(apply);
          }
        })
        .catch(() => undefined);
    } else {
      load(formType).then(apply).catch(() => undefined);
    }

    return () => {
      cancelled = true;
    };
  }, [formType, owner, tag]);

  return formItems;
}
