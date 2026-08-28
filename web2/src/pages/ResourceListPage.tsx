import * as React from "react";
import i18next from "i18next";
import {Upload} from "lucide-react";
import {Link} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as ResourceBackend from "@/backend/ResourceBackend";
import * as Setting from "@/lib/setting";

export default function ResourceListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  const fileInputRef = React.useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = React.useState(false);

  // same call as the antd page: tag "custom", parent "ResourceListPage",
  // path "resource/<owner>/<user>/<filename>"
  const handleUpload = (file: File, refresh: () => void) => {
    if (!account) {
      return;
    }
    setUploading(true);
    const fullFilePath = `resource/${account.owner}/${account.name}/${file.name}`;
    ResourceBackend.uploadResource(account.owner, account.name, "custom", "ResourceListPage", fullFilePath, file)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("application:File uploaded successfully"));
          refresh();
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      })
      .finally(() => {
        setUploading(false);
      });
  };

  const columns: ColumnDef<any>[] = [
    textColumn({dataIndex: "provider", title: i18next.t("general:Provider"), width: 150, searchable: true}),
    organizationColumn(),
    textColumn({dataIndex: "application", title: i18next.t("general:Application"), width: 150, searchable: true}),
    {
      dataIndex: "user",
      title: i18next.t("general:User"),
      width: 130,
      sortable: true,
      searchable: true,
      render: (value, record) =>
        value ? (
          <Link to={`/users/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    textColumn({dataIndex: "parent", title: i18next.t("general:Parent"), width: 130}),
    textColumn({dataIndex: "name", title: i18next.t("general:Name"), width: 200, searchable: true}),
    dateColumn(),
    textColumn({dataIndex: "tag", title: i18next.t("general:Tag"), width: 110}),
    textColumn({dataIndex: "fileType", title: i18next.t("general:Type"), width: 110}),
    textColumn({dataIndex: "fileFormat", title: i18next.t("resource:Format"), width: 110}),
    {
      dataIndex: "fileSize",
      title: i18next.t("resource:File size"),
      width: 120,
      sortable: true,
      render: (value) => Setting.getFriendlyFileSize(value),
    },
    {
      dataIndex: "url",
      title: i18next.t("general:URL"),
      width: 200,
      render: (value) =>
        value ? (
          <a href={value} target="_blank" rel="noreferrer" className="underline-offset-4 hover:underline">
            {Setting.getShortText(value)}
          </a>
        ) : null,
    },
    {
      key: "preview",
      dataIndex: "url",
      title: i18next.t("general:Preview"),
      width: 130,
      render: (value, record) =>
        value && record.fileType === "image" ? (
          <a href={value} target="_blank" rel="noreferrer">
            <img src={value} alt={record.name} className="h-10 object-contain" />
          </a>
        ) : value ? (
          <a href={value} target="_blank" rel="noreferrer" className="underline-offset-4 hover:underline">
            {i18next.t("general:Preview")}
          </a>
        ) : null,
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Resources")}
      columns={columns}
      deps={[organizationName, account?.name]}
      showActionColumn
      toolbar={({refresh}) => (
        <>
          <input
            ref={fileInputRef}
            type="file"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) {
                handleUpload(file, refresh);
              }
              // let the same file be picked again
              e.target.value = "";
            }}
          />
          <Button loading={uploading} onClick={() => fileInputRef.current?.click()}>
            <Upload />
            {i18next.t("resource:Upload a file...")}
          </Button>
        </>
      )}
      fetch={(q) =>
        ResourceBackend.getResources(
          organizationName,
          account?.name ?? "",
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      remove={(r) => ResourceBackend.deleteResource(r)}
      actionColumnWidth={120}
      rowKey={(row, index) => `${row.owner}/${row.name}/${index}`}
    />
  );
}
