import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Progress} from "@/components/ui/progress";
import * as Setting from "@/lib/setting";

export function handleCameraError(error: any) {
  if (!(error instanceof DOMException)) {
    return;
  }
  if (error.name === "NotFoundError" || error.name === "DevicesNotFoundError") {
    Setting.showMessage("error", i18next.t("login:Please ensure that you have a camera device for facial recognition"));
  } else if (error.name === "NotAllowedError" || error.name === "PermissionDeniedError") {
    Setting.showMessage("error", i18next.t("login:Please provide permission to access the camera"));
  } else if (error.name === "NotReadableError" || error.name === "TrackStartError") {
    Setting.showMessage("error", i18next.t("login:The camera is currently in use by another webpage"));
  } else if (error.name === "TypeError") {
    Setting.showMessage("error", i18next.t("login:Please load the webpage using HTTPS, otherwise the camera cannot be accessed"));
  } else {
    Setting.showMessage("error", error.message);
  }
}

/** The circular progress ring drawn around the camera preview. */
export function FaceRing({percent}: {percent: number}) {
  return (
    <div className="pointer-events-none absolute left-1/2 top-1/2 h-[240px] w-[240px] -translate-x-1/2 -translate-y-1/2">
      <svg width="240" height="240" fill="none">
        <circle
          strokeDasharray="700"
          strokeDashoffset={700 - 6.9115 * percent}
          strokeWidth="4"
          cx="120"
          cy="120"
          r="110"
          stroke="#5734d3"
          transform="rotate(-90, 120, 120)"
          strokeLinecap="round"
          style={{transition: "all .2s linear"}}
        />
      </svg>
    </div>
  );
}

interface FaceRecognitionCommonModalProps {
  visible: boolean;
  onOk: (capturedImages: string[]) => void;
  onCancel: () => void;
}

/**
 * Camera capture for applications that have a Face ID provider: it grabs a few
 * frames and hands the base64 images to the backend, which does the recognition.
 * Ported from web/src/common/modal/FaceRecognitionCommonModal.js — the same four
 * frames are captured over the same timing.
 */
export function FaceRecognitionCommonModal({visible, onOk, onCancel}: FaceRecognitionCommonModalProps) {
  const videoRef = React.useRef<HTMLVideoElement>(null);
  const mediaStreamRef = React.useRef<MediaStream | null>(null);
  const capturedRef = React.useRef<string[]>([]);
  const [percent, setPercent] = React.useState(0);
  const [captured, setCaptured] = React.useState(false);
  const [hasImages, setHasImages] = React.useState(false);

  const onOkRef = React.useRef(onOk);
  onOkRef.current = onOk;

  React.useEffect(() => {
    if (!visible) {
      setCaptured(false);
      capturedRef.current = [];
      setHasImages(false);
      setPercent(0);
      return;
    }

    let cancelled = false;
    navigator.mediaDevices
      .getUserMedia({video: {facingMode: "user"}})
      .then((stream) => {
        if (cancelled) {
          stream.getTracks().forEach((track) => track.stop());
          return;
        }
        mediaStreamRef.current = stream;
        setCaptured(true);
      })
      .catch(handleCameraError);

    return () => {
      cancelled = true;
    };
  }, [visible]);

  React.useEffect(() => {
    if (!captured) {
      mediaStreamRef.current?.getTracks().forEach((track) => track.stop());
      mediaStreamRef.current = null;
      if (videoRef.current) {
        videoRef.current.srcObject = null;
      }
      return;
    }

    let tick = 0;
    const video = videoRef.current;
    if (video) {
      video.srcObject = mediaStreamRef.current;
      video.play();
    }

    // the first three seconds let the user settle, then four frames are taken
    const timer = window.setInterval(() => {
      tick++;
      if (tick >= 8) {
        window.clearInterval(timer);
        setPercent(0);
        onOkRef.current(capturedRef.current);
        return;
      }
      if (tick > 3 && video) {
        setPercent((tick - 4) * 20);
        const canvas = document.createElement("canvas");
        canvas.width = video.videoWidth;
        canvas.height = video.videoHeight;
        canvas.getContext("2d")?.drawImage(video, 0, 0, canvas.width, canvas.height);
        capturedRef.current = [...capturedRef.current, canvas.toDataURL("image/png")];
        setHasImages(true);
      }
    }, 1000);

    return () => window.clearInterval(timer);
  }, [captured]);

  return (
    <Dialog open={visible} onOpenChange={(next) => (next ? undefined : onCancel())}>
      <DialogContent className="sm:max-w-[350px]" onInteractOutside={(e) => e.preventDefault()}>
        <DialogHeader>
          <DialogTitle>{i18next.t("login:Face Recognition")}</DialogTitle>
        </DialogHeader>
        <Progress value={percent} />
        <div className="relative mb-12 mt-5 flex justify-center">
          <video ref={videoRef} className="h-[220px] w-[220px] rounded-full object-cover" />
          <FaceRing percent={percent} />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>{i18next.t("general:Cancel")}</Button>
          <Button disabled={!hasImages} onClick={() => onOk(capturedRef.current)}>{i18next.t("general:OK")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
