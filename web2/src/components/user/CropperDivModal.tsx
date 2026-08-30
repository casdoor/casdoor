import * as React from "react";
import i18next from "i18next";
import Cropper from "react-cropper";
import "cropperjs/dist/cropper.css";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import * as ResourceBackend from "@/backend/ResourceBackend";
import * as Setting from "@/lib/setting";

interface CropperDivModalProps {
  /** the resource tag, also the field the backend writes the URL back to ("avatar", "idCardFront", ...) */
  tag: string;
  title: string;
  /** label of the confirm button */
  setTitle: string;
  buttonText: string;
  user: any;
  organization: any;
  disabled?: boolean;
  /** called after a successful upload so the page can pick up the new URL */
  onUploaded?: () => void;
}

function getBase64Image(src: string): Promise<string> {
  return new Promise((resolve) => {
    const image = new Image();
    image.src = src;
    image.setAttribute("crossOrigin", "anonymous");
    image.onload = () => {
      const canvas = document.createElement("canvas");
      canvas.width = image.width;
      canvas.height = image.height;
      const ctx = canvas.getContext("2d");
      ctx?.drawImage(image, 0, 0, image.width, image.height);
      resolve(canvas.toDataURL("image/png"));
    };
  });
}

/**
 * Picks a picture (from disk or from the user's existing resources), crops it and
 * uploads it as a resource named after `tag`. Ported from
 * web/src/common/modal/CropperDivModal.js — the upload call and the
 * "<tag>/<owner>/<name>.<ext>" path are unchanged, so the backend writes the URL
 * back to the same user field.
 */
export function CropperDivModal({
  tag,
  title,
  setTitle,
  buttonText,
  user,
  organization,
  disabled,
  onUploaded,
}: CropperDivModalProps) {
  const [open, setOpen] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [uploading, setUploading] = React.useState(false);
  const [options, setOptions] = React.useState<string[]>([]);
  const [image, setImage] = React.useState("");
  const cropperRef = React.useRef<any>(null);
  const fileInputRef = React.useRef<HTMLInputElement>(null);

  React.useEffect(() => {
    if (!open) {
      return;
    }
    setLoading(true);
    ResourceBackend.getResources(user.owner, user.name, "", "", "", "", "", "")
      .then((res: any) => {
        setLoading(false);
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }
        const urls = (res.data ?? [])
          .filter((item: any) => item.fileType === "image")
          .map((item: any) => item.url);
        setOptions([organization?.defaultAvatar, ...urls].filter(Boolean));
      })
      .catch(() => setLoading(false));
  }, [open, user.owner, user.name, organization?.defaultAvatar]);

  const onFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) {
      return;
    }
    const reader = new FileReader();
    reader.onload = () => setImage(String(reader.result));
    reader.readAsDataURL(file);
    e.target.value = "";
  };

  const upload = () => {
    const cropper = cropperRef.current?.cropper;
    if (!cropper) {
      Setting.showMessage("error", i18next.t("general:You must select a picture first"));
      return;
    }
    setUploading(true);
    cropper.getCroppedCanvas().toBlob((blob: Blob | null) => {
      if (blob === null) {
        setUploading(false);
        Setting.showMessage("error", i18next.t("general:You must select a picture first"));
        return;
      }
      const extension = image.substring(image.indexOf("/") + 1, image.indexOf(";base64"));
      const fullFilePath = `${tag}/${user.owner}/${user.name}.${extension}`;
      ResourceBackend.uploadResource(user.owner, user.name, tag, "CropperDivModal", fullFilePath, blob)
        .then((res: any) => {
          setUploading(false);
          if (res.status === "ok") {
            setOpen(false);
            setImage("");
            onUploaded?.();
          } else {
            Setting.showMessage("error", res.msg);
          }
        })
        .catch((error: any) => {
          setUploading(false);
          Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        });
    });
  };

  return (
    <>
      <Button variant="outline" size="sm" disabled={disabled} onClick={() => setOpen(true)}>
        {buttonText}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
          </DialogHeader>
          <div className="space-y-2">
            <input ref={fileInputRef} type="file" accept="image/*" className="hidden" onChange={onFileChange} />
            <Button variant="outline" className="w-full" onClick={() => fileInputRef.current?.click()}>
              {i18next.t("user:Select a photo...")}
            </Button>
            <SearchableSelect
              value=""
              disabled={loading}
              placeholder={i18next.t("user:Please select avatar from resources")}
              onChange={async(value) => setImage(await getBase64Image(value))}
              options={options.map((url) => ({value: url, label: url}))}
            />
            <div className="h-[320px]">
              <Cropper
                ref={cropperRef}
                style={{height: "100%", width: "100%"}}
                initialAspectRatio={1}
                src={image}
                viewMode={1}
                guides={true}
                minCropBoxHeight={10}
                minCropBoxWidth={10}
                background={false}
                responsive={true}
                autoCropArea={1}
                checkOrientation={false}
              />
            </div>
            <Button className="w-full" loading={uploading} disabled={!image} onClick={upload}>
              {setTitle}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

interface UserImageFieldProps {
  imageUrl: string;
  title: string;
  setTitle: string;
  tag: string;
  user: any;
  organization: any;
  disabled?: boolean;
  /** the upload creates a resource owned by the user, so it needs the user to exist */
  canUpload: boolean;
  onUploaded?: () => void;
}

/** One picture slot: preview plus the crop-and-upload dialog (web/src/UserEditPage.js renderImage). */
export function UserImageField({
  imageUrl,
  title,
  setTitle,
  tag,
  user,
  organization,
  disabled,
  canUpload,
  onUploaded,
}: UserImageFieldProps) {
  return (
    <div className="flex w-40 flex-col items-center gap-2 text-center">
      {imageUrl ? (
        <a target="_blank" rel="noreferrer" href={imageUrl}>
          <img src={imageUrl} alt={imageUrl} className="h-[150px] w-[150px] rounded object-contain" />
        </a>
      ) : (
        <div className="flex h-[150px] w-[150px] flex-col items-center justify-center rounded border border-dashed text-muted-foreground">
          <div className="text-3xl">+</div>
          <div className="text-xs">{`(${i18next.t("general:empty")})`}</div>
        </div>
      )}
      {canUpload ? (
        <CropperDivModal
          tag={tag}
          title={title}
          setTitle={setTitle}
          buttonText={`${title}...`}
          user={user}
          organization={organization}
          disabled={disabled}
          onUploaded={onUploaded}
        />
      ) : null}
    </div>
  );
}
