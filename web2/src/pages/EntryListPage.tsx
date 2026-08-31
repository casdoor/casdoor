import * as React from "react";
import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {clientIpColumn, dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {Popover, PopoverContent, PopoverTrigger} from "@/components/ui/popover";
import {EntryMessageViewer} from "@/components/entry/EntryMessageViewer";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as EntryBackend from "@/backend/EntryBackend";
import * as ProviderBackend from "@/backend/ProviderBackend";
import {newEntry} from "@/pages/defaults";
import * as Setting from "@/lib/setting";

/**
 * The Log providers of the current organization, keyed by name. The message
 * viewer needs the provider to know whether an entry is SELinux or OpenClaw, and
 * fetching one per row would hammer `get-provider`.
 */
function useLogProviderMap(owner: string) {
  const [providerMap, setProviderMap] = React.useState<Record<string, any>>({});

  React.useEffect(() => {
    if (!owner) {
      setProviderMap({});
      return;
    }

    let cancelled = false;
    ProviderBackend.getProviders(owner)
      .then((res: any) => {
        if (cancelled) {
          return;
        }
        if (res.status !== "ok") {
          setProviderMap({});
          return;
        }
        const map: Record<string, any> = {};
        (res.data || []).forEach((provider: any) => {
          if (provider?.category === "Log" && provider?.name) {
            map[provider.name] = provider;
          }
        });
        setProviderMap(map);
      })
      .catch(() => {
        if (!cancelled) {
          setProviderMap({});
        }
      });

    return () => {
      cancelled = true;
    };
  }, [owner]);

  return providerMap;
}

/**
 * The message cell: a short preview that opens the full viewer in a popover.
 * The antd frontend opened it on hover, but the viewer is interactive here (the
 * session graph can be panned and its nodes clicked), so it opens on click.
 */
function MessageCell({text, record, provider}: {text: string; record: any; provider: any}) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button type="button" className="text-left underline-offset-4 hover:underline">
          {Setting.getShortText(text, 60)}
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="max-h-[70vh] w-[min(90vw,720px)] overflow-y-auto">
        <EntryMessageViewer entry={record} provider={provider} block />
      </PopoverContent>
    </Popover>
  );
}

export default function EntryListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  const providerMap = useLogProviderMap(organizationName);

  const columns: ColumnDef<any>[] = [
    organizationColumn(),
    linkColumn({dataIndex: "name", to: (r) => `/entries/${r.owner}/${r.name}`, width: 180}),
    dateColumn(),
    {
      ...textColumn({dataIndex: "provider", title: i18next.t("general:Provider"), width: 150, searchable: true}),
      link: (value, record: any) => (value ? `/providers/${record.owner}/${value}` : undefined),
    },
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 120, searchable: true}),
    clientIpColumn(),
    textColumn({dataIndex: "userAgent", title: i18next.t("general:User agent"), width: 200, searchable: true}),
    {
      dataIndex: "message",
      title: i18next.t("payment:Message"),
      width: 220,
      sortable: true,
      searchable: true,
      render: (text: string, record: any) =>
        text ? <MessageCell text={text} record={record} provider={providerMap[record.provider] ?? null} /> : null,
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Entries")}
      columns={columns}
      formType="entries"
      deps={[organizationName]}
      fetch={(q) =>
        EntryBackend.getEntries(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newEntry(account) : undefined}
      editUrl={(r) => `/entries/${r.owner}/${r.name}`}
      remove={(r) => EntryBackend.deleteEntry(r)}
    />
  );
}
