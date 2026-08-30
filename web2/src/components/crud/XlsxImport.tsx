import * as React from "react";
import i18next from "i18next";
import * as XLSX from "xlsx";
import {Upload} from "lucide-react";
import {Button} from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import * as Setting from "@/lib/setting";

interface XlsxImportProps {
  /** the "<translated label>#<field>" list, e.g. Setting.getUserColumns() */
  columns: string[];
  /** file name of the empty template, e.g. "import-user.xlsx" */
  templateName: string;
  /** the endpoint under /api, e.g. "upload-users" */
  uploadApi: string;
  /** shown after a successful upload, e.g. "Users uploaded successfully, refreshing the page" */
  successMessage: string;
  onUploaded: () => void;
}

/**
 * "Download template" + "Upload (.xlsx)" pair used by the user, group, role and
 * permission list pages. The sheet is parsed client-side only to preview it; the
 * file itself is what the backend imports.
 */
export function XlsxImport({columns, templateName, uploadApi, successMessage, onUploaded}: XlsxImportProps) {
  const inputRef = React.useRef<HTMLInputElement>(null);
  const [file, setFile] = React.useState<File | null>(null);
  const [rows, setRows] = React.useState<any[]>([]);
  const [uploading, setUploading] = React.useState(false);

  const close = () => {
    setFile(null);
    setRows([]);
  };

  const downloadTemplate = () => {
    const empty: Record<string, null> = {};
    columns.forEach((column) => {
      empty[column] = null;
    });
    const workbook = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(workbook, XLSX.utils.json_to_sheet([empty]), "Sheet1");
    XLSX.writeFile(workbook, templateName, {compression: true});
  };

  const readFile = (selected: File) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const workbook = XLSX.read(e.target?.result, {type: "array"});
        if (!workbook.SheetNames || workbook.SheetNames.length === 0) {
          Setting.showMessage("error", i18next.t("general:No sheets found in file"));
          return;
        }
        setRows(XLSX.utils.sheet_to_json(workbook.Sheets[workbook.SheetNames[0]]));
        setFile(selected);
      } catch (err: any) {
        Setting.showMessage("error", `${i18next.t("general:Failed to upload")}: ${err?.message ?? err}`);
      }
    };
    reader.onerror = (error: any) => {
      Setting.showMessage("error", `${i18next.t("general:Failed to upload")}: ${error?.message ?? error}`);
    };
    reader.readAsArrayBuffer(selected);
  };

  const upload = () => {
    if (file === null) {
      return;
    }
    const formData = new FormData();
    formData.append("file", file);
    setUploading(true);
    fetch(`${Setting.ServerUrl}/api/${uploadApi}`, {
      method: "post",
      body: formData,
      credentials: "include",
      headers: {
        "Accept-Language": Setting.getAcceptLanguage(),
      },
    })
      .then((res) => res.json())
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", successMessage);
          onUploaded();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to upload")}: ${res.msg}`);
        }
        close();
      })
      .catch((error) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to upload")}: ${error.message}`);
      })
      .finally(() => setUploading(false));
  };

  return (
    <>
      <Button variant="outline" onClick={downloadTemplate}>
        {i18next.t("general:Download template")}
      </Button>
      <Button variant="outline" id="upload-button" onClick={() => inputRef.current?.click()}>
        <Upload />
        {i18next.t("general:Upload (.xlsx)")}
      </Button>
      <input
        ref={inputRef}
        type="file"
        accept=".xlsx"
        className="hidden"
        onChange={(e) => {
          const selected = e.target.files?.[0];
          if (selected) {
            readFile(selected);
          }
          // let the same file be picked again after a cancelled import
          e.target.value = "";
        }}
      />

      <Dialog open={file !== null} onOpenChange={(open) => (open ? undefined : close())}>
        <DialogContent className="max-w-[95vw]">
          <DialogHeader>
            <DialogTitle>{i18next.t("general:Upload (.xlsx)")}</DialogTitle>
          </DialogHeader>
          <div className="max-h-[60vh] overflow-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  {columns.map((column) => (
                    <TableHead key={column} className="whitespace-nowrap">
                      {column.split("#")[0]}
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.map((row, index) => (
                  <TableRow key={index}>
                    {columns.map((column) => (
                      <TableCell key={column} className="whitespace-nowrap">
                        {String(row[column] ?? "")}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={close}>
              {i18next.t("general:Cancel")}
            </Button>
            <Button onClick={upload} disabled={uploading}>
              {i18next.t("general:Click to Upload")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
