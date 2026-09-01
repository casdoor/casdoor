import * as React from "react";
import i18next from "i18next";
import {ShieldCheck} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {CaptchaWidget} from "@/components/common/CaptchaWidget";
import * as UserBackend from "@/backend/UserBackend";

export interface CaptchaHandle {
  loadCaptcha: () => void;
}

interface CaptchaModalProps {
  owner: string;
  name: string;
  /** open the dialog (ignored when `noModal` renders the widget inline) */
  visible?: boolean;
  isCurrentProvider?: boolean;
  /** render the captcha inline instead of inside a dialog */
  noModal?: boolean;
  onOk?: (captchaType: string, captchaToken: string, clientSecret: string) => void;
  onCancel?: () => void;
  /** inline mode: report every keystroke/token so the parent can submit it */
  onUpdateToken?: (captchaType: string, captchaToken: string, clientSecret: string) => void;
  innerRef?: React.MutableRefObject<CaptchaHandle | null>;
}

/**
 * The captcha challenge, either as a dialog (before login / send-code) or inline
 * in the sign-in form. Ported from web/src/common/modal/CaptchaModal.js: the
 * provider is fetched from /api/get-captcha and the (captchaType, captchaToken,
 * clientSecret) triple is handed back to the caller unchanged.
 */
export function CaptchaModal({
  owner,
  name,
  visible,
  isCurrentProvider = false,
  noModal = false,
  onOk,
  onCancel,
  onUpdateToken,
  innerRef,
}: CaptchaModalProps) {
  const [captchaType, setCaptchaType] = React.useState("none");
  const [clientId, setClientId] = React.useState("");
  const [clientSecret, setClientSecret] = React.useState("");
  const [subType, setSubType] = React.useState("");
  const [clientId2, setClientId2] = React.useState("");
  const [clientSecret2, setClientSecret2] = React.useState("");

  const [open, setOpen] = React.useState(false);
  const [captchaImg, setCaptchaImg] = React.useState("");
  const [captchaToken, setCaptchaToken] = React.useState("");
  const defaultInputRef = React.useRef<HTMLInputElement>(null);

  const handleOk = React.useCallback(
    (type = captchaType, token = captchaToken, secret = clientSecret) => {
      onOk?.(type, token, secret);
    },
    [captchaType, captchaToken, clientSecret, onOk],
  );

  const handleCancel = () => {
    setCaptchaToken("");
    onCancel?.();
  };

  const loadCaptcha = React.useCallback(
    (shouldFocus = false) => {
      UserBackend.getCaptcha(owner, name, isCurrentProvider).then((res: any) => {
        if (res.type === "none") {
          onOk?.("none", "", "");
        } else if (res.type === "Default") {
          setOpen(true);
          setClientSecret(res.captchaId);
          setCaptchaImg(res.captchaImage);
          setCaptchaType("Default");
          // the previous code was consumed by the last verification, so the stale input must go
          setCaptchaToken("");
          if (noModal) {
            onUpdateToken?.("Default", "", res.captchaId);
          }
          if (shouldFocus) {
            defaultInputRef.current?.focus();
          }
        } else {
          setOpen(true);
          setCaptchaType(res.type);
          setClientId(res.clientId);
          setClientSecret(res.clientSecret);
          setSubType(res.subType);
          setClientId2(res.clientId2);
          setClientSecret2(res.clientSecret2);
        }
      });
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [owner, name, isCurrentProvider, noModal],
  );

  React.useEffect(() => {
    if (visible || noModal) {
      loadCaptcha();
    } else {
      setCaptchaToken("");
      setOpen(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visible, noModal]);

  // A widget-based captcha resolves on its own, so submit as soon as it does.
  React.useEffect(() => {
    if (captchaToken !== "" && captchaType !== "Default" && !noModal) {
      handleOk();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [captchaToken]);

  React.useEffect(() => {
    if (innerRef) {
      innerRef.current = {loadCaptcha: () => loadCaptcha(true)};
    }
  }, [innerRef, loadCaptcha]);

  const onTokenChange = (token: string) => {
    setCaptchaToken(token);
    if (noModal) {
      onUpdateToken?.(captchaType, token, clientSecret);
    }
  };

  const renderDefaultCaptcha = () =>
    noModal ? (
      <div className="flex gap-2">
        <div className="relative flex-1">
          <ShieldCheck className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            ref={defaultInputRef}
            className="pl-8"
            value={captchaToken}
            placeholder={i18next.t("general:Captcha")}
            onChange={(e) => onTokenChange(e.target.value)}
          />
        </div>
        <img
          src={`data:image/png;base64,${captchaImg}`}
          onClick={() => loadCaptcha(true)}
          className="h-9 w-[110px] cursor-pointer rounded-md border object-cover"
          alt="captcha"
        />
      </div>
    ) : (
      <div className="flex flex-col items-center gap-3">
        <img
          src={`data:image/png;base64,${captchaImg}`}
          onClick={() => loadCaptcha(true)}
          className="h-20 w-[200px] cursor-pointer rounded-md border object-cover"
          alt="captcha"
        />
        <div className="relative w-[200px]">
          <ShieldCheck className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            ref={defaultInputRef}
            className="pl-8"
            value={captchaToken}
            placeholder={i18next.t("general:Captcha")}
            onChange={(e) => setCaptchaToken(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                handleOk();
              }
            }}
          />
        </div>
      </div>
    );

  const renderCaptcha = () =>
    captchaType === "Default" ? (
      renderDefaultCaptcha()
    ) : (
      <div className="flex justify-center">
        <CaptchaWidget
          captchaType={captchaType}
          subType={subType}
          siteKey={clientId}
          clientSecret={clientSecret}
          clientId2={clientId2}
          clientSecret2={clientSecret2}
          onChange={onTokenChange}
        />
      </div>
    );

  if (noModal) {
    return captchaType === "none" ? null : renderCaptcha();
  }

  const okDisabled = captchaType === "Default" && !/^\d{5}$/.test(captchaToken);

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) {
          handleCancel();
        }
        setOpen(next);
      }}
    >
      <DialogContent className="sm:max-w-[350px]">
        <DialogHeader>
          <DialogTitle>{i18next.t("general:Captcha")}</DialogTitle>
        </DialogHeader>
        <div className="py-2">{renderCaptcha()}</div>
        {captchaType === "Default" ? (
          <Button disabled={okDisabled} onClick={() => handleOk()}>
            {i18next.t("general:OK")}
          </Button>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
