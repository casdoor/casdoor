import * as React from "react";
import i18next from "i18next";
import {Camera, Upload} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {FaceRecognitionModal} from "@/components/common/FaceRecognitionModal";
import * as ResourceBackend from "@/backend/ResourceBackend";
import * as Setting from "@/lib/setting";

/** A user may enrol at most five faces, as in web/src/table/FaceIdTable.js. */
const MAX_FACE_IDS = 5;

interface FaceIdTableProps {
  table: any[];
  account: any;
  onUpdateTable: (rows: any[]) => void;
}

async function dataUrlToFile(dataUrl: string, filename: string) {
  const res = await fetch(dataUrl);
  const blob = await res.blob();
  return new File([blob], filename, {type: blob.type || "image/jpeg"});
}

/**
 * The user's enrolled faces. A face can be added from the camera (descriptor),
 * from a photo (descriptor) or as an uploaded image. Ported from
 * web/src/table/FaceIdTable.js.
 */
export function FaceIdTable({table, account, onUpdateTable}: FaceIdTableProps) {
  const rows = table ?? [];
  const [descriptorModal, setDescriptorModal] = React.useState<"camera" | "image" | null>(null);
  const [cameraImageModal, setCameraImageModal] = React.useState(false);
  const [uploading, setUploading] = React.useState(false);
  const fileInputRef = React.useRef<HTMLInputElement>(null);

  const addFaceId = (faceIdData: number[]) => {
    onUpdateTable([...rows, {name: Setting.getRandomName(), faceIdData}]);
  };

  const addFaceImage = (imageUrl: string) => {
    onUpdateTable([...rows, {name: Setting.getRandomName(), imageUrl, faceIdData: []}]);
  };

  const uploadFaceImage = (file: File) => {
    setUploading(true);
    const fullFilePath = `resource/${account.owner}/${account.name}/${file.name}`;
    ResourceBackend.uploadResource(account.owner, account.name, "custom", "ResourceListPage", fullFilePath, file)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("application:File uploaded successfully"));
          addFaceImage(res.data);
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => {
        setUploading(false);
        setCameraImageModal(false);
      });
  };

  const full = rows.length >= MAX_FACE_IDS;

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-sm font-medium">{i18next.t("user:Face IDs")}</span>
        <Button size="sm" disabled={full} onClick={() => setDescriptorModal("camera")}>
          {i18next.t("application:Add Face ID")}
        </Button>
        <Button variant="outline" size="sm" disabled={full} onClick={() => setDescriptorModal("image")}>
          {i18next.t("application:Add Face ID with Image")}
        </Button>
        <Button variant="outline" size="sm" disabled={full} onClick={() => setCameraImageModal(true)}>
          <Camera />
          {i18next.t("application:Add Face ID with Camera")}
        </Button>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0];
            if (file) {
              uploadFaceImage(file);
            }
            e.target.value = "";
          }}
        />
        <Button variant="outline" size="sm" loading={uploading} disabled={full} onClick={() => fileInputRef.current?.click()}>
          <Upload />
          {i18next.t("resource:Upload a file...")}
        </Button>
      </div>

      <div className="overflow-x-auto rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="w-[200px]">{i18next.t("general:Name")}</TableHead>
              <TableHead>{i18next.t("general:Data")}</TableHead>
              <TableHead className="w-[200px]">{i18next.t("general:URL")}</TableHead>
              <TableHead className="w-[110px]">{i18next.t("general:Action")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.length === 0 ? (
              <TableRow className="hover:bg-transparent">
                <TableCell colSpan={4} className="h-20 text-center text-muted-foreground">
                  {i18next.t("general:No data")}
                </TableCell>
              </TableRow>
            ) : (
              rows.map((record: any, index: number) => {
                const data: number[] = record.faceIdData ?? [];
                return (
                  <TableRow key={`${record.name}-${index}`}>
                    <TableCell>
                      <Input
                        value={record.name ?? ""}
                        onChange={(e) => {
                          const next = [...rows];
                          next[index] = {...next[index], name: e.target.value};
                          onUpdateTable(next);
                        }}
                      />
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {data.length > 0 ? `[${data.slice(0, 3).join(", ")} ... ${data.slice(-3).join(", ")}]` : null}
                    </TableCell>
                    <TableCell className="truncate text-xs text-muted-foreground">{record.imageUrl}</TableCell>
                    <TableCell>
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => onUpdateTable([...rows.slice(0, index), ...rows.slice(index + 1)])}
                      >
                        {i18next.t("general:Delete")}
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>

      {descriptorModal !== null ? (
        <FaceRecognitionModal
          visible={true}
          withImage={descriptorModal === "image"}
          onOk={(faceIdData) => {
            addFaceId(faceIdData);
            setDescriptorModal(null);
          }}
          onCancel={() => setDescriptorModal(null)}
        />
      ) : null}

      {cameraImageModal ? (
        <FaceRecognitionModal
          visible={true}
          withImage={true}
          captureImage={true}
          onOk={async(imageDataUrl: string) => {
            uploadFaceImage(await dataUrlToFile(imageDataUrl, `face-id-${Date.now()}.jpg`));
          }}
          onCancel={() => setCameraImageModal(false)}
        />
      ) : null}
    </div>
  );
}
