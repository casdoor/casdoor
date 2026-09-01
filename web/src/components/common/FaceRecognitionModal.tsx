import * as React from "react";
import i18next from "i18next";
import * as faceapi from "face-api.js";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Progress} from "@/components/ui/progress";
import {Loading} from "@/components/common/Loading";
import {FaceRing, handleCameraError} from "@/components/common/FaceRecognitionCommonModal";
import * as Setting from "@/lib/setting";

interface FaceRecognitionModalProps {
  visible: boolean;
  /** true: enrol from an uploaded photo instead of the camera */
  withImage?: boolean;
  /** with `withImage`, capture a JPEG from the camera instead of a descriptor */
  captureImage?: boolean;
  onOk: (value: any) => void;
  onCancel: () => void;
}

interface PickedFile {
  name: string;
  base64: string;
}

/**
 * Local face recognition with face-api.js, ported from
 * web/src/common/modal/FaceRecognitionModal.js. The models are loaded from the
 * same static path and the 128-float descriptor handed to `onOk` is unchanged,
 * so the backend matches it exactly as before.
 *
 * Three modes, as in the antd version:
 *   withImage=false                  → camera, returns the descriptor
 *   withImage=true                   → upload a photo, returns the descriptor
 *   withImage=true, captureImage=true → camera, returns a JPEG data URL
 */
export function FaceRecognitionModal({visible, withImage, captureImage, onOk, onCancel}: FaceRecognitionModalProps) {
  const videoRef = React.useRef<HTMLVideoElement>(null);
  const detectionRef = React.useRef<number | null>(null);
  const mediaStreamRef = React.useRef<MediaStream | null>(null);
  const [modelsLoaded, setModelsLoaded] = React.useState(false);
  const [cameraReady, setCameraReady] = React.useState(false);
  const [percent, setPercent] = React.useState(0);

  const [files, setFiles] = React.useState<PickedFile[]>([]);
  const [currentFace, setCurrentFace] = React.useState<any>(null);
  const [currentIndex, setCurrentIndex] = React.useState<number | null>(null);

  const useCamera = !withImage || captureImage;
  const onOkRef = React.useRef(onOk);
  onOkRef.current = onOk;
  const onCancelRef = React.useRef(onCancel);
  onCancelRef.current = onCancel;

  React.useEffect(() => {
    if (!visible || modelsLoaded) {
      return;
    }
    const modelUrl = `${Setting.StaticBaseUrl}/casdoor/models`;
    Promise.all([
      faceapi.nets.tinyFaceDetector.loadFromUri(modelUrl),
      faceapi.nets.faceLandmark68Net.loadFromUri(modelUrl),
      faceapi.nets.faceRecognitionNet.loadFromUri(modelUrl),
    ])
      .then(() => setModelsLoaded(true))
      .catch(() => {
        Setting.showMessage("error", i18next.t("login:Model loading failure"));
        onCancelRef.current();
      });
  }, [visible, modelsLoaded]);

  React.useEffect(() => {
    if (!useCamera) {
      return;
    }
    if (!visible || !modelsLoaded) {
      if (detectionRef.current !== null) {
        window.clearInterval(detectionRef.current);
        detectionRef.current = null;
      }
      setCameraReady(false);
      return;
    }

    let cancelled = false;
    setPercent(0);
    navigator.mediaDevices
      .getUserMedia({video: {facingMode: "user"}})
      .then((stream) => {
        if (cancelled) {
          stream.getTracks().forEach((track) => track.stop());
          return;
        }
        mediaStreamRef.current = stream;
        setCameraReady(true);
      })
      .catch((error) => {
        onCancelRef.current();
        handleCameraError(error);
      });

    return () => {
      cancelled = true;
      if (detectionRef.current !== null) {
        window.clearInterval(detectionRef.current);
        detectionRef.current = null;
      }
      setCameraReady(false);
    };
  }, [visible, modelsLoaded, useCamera]);

  React.useEffect(() => {
    if (!useCamera) {
      return;
    }
    if (!cameraReady) {
      mediaStreamRef.current?.getTracks().forEach((track) => track.stop());
      mediaStreamRef.current = null;
      if (videoRef.current) {
        videoRef.current.srcObject = null;
      }
      return;
    }
    if (videoRef.current) {
      videoRef.current.srcObject = mediaStreamRef.current;
      videoRef.current.play();
    }
  }, [cameraReady, useCamera]);

  const handleStreamVideo = () => {
    if (!useCamera || detectionRef.current !== null) {
      return;
    }
    let count = 0;
    let goodCount = 0;

    detectionRef.current = window.setInterval(async() => {
      const video = videoRef.current;
      if (!modelsLoaded || !video || !visible) {
        return;
      }
      const faces = await faceapi
        .detectAllFaces(video, new faceapi.TinyFaceDetectorOptions())
        .withFaceLandmarks()
        .withFaceDescriptors();

      count++;
      if (count % 50 === 0) {
        Setting.showMessage("warning", i18next.t("login:Please ensure sufficient lighting and align your face in the center of the recognition box"));
      } else if (count > 300) {
        Setting.showMessage("error", i18next.t("login:Face recognition failed"));
        onCancelRef.current();
        return;
      }

      if (faces.length !== 1) {
        setPercent((prev) => Math.round(prev / 2));
        return;
      }

      const face = faces[0];
      setPercent(Math.round(face.detection.score * 100));
      if (face.detection.score <= 0.9) {
        return;
      }

      goodCount++;
      if (face.detection.score > 0.99 || goodCount > 10) {
        if (detectionRef.current !== null) {
          window.clearInterval(detectionRef.current);
          detectionRef.current = null;
        }
        if (captureImage) {
          const canvas = document.createElement("canvas");
          canvas.width = video.videoWidth;
          canvas.height = video.videoHeight;
          canvas.getContext("2d")?.drawImage(video, 0, 0, canvas.width, canvas.height);
          onOkRef.current(canvas.toDataURL("image/jpeg", 0.92));
        } else {
          onOkRef.current(Array.from(face.descriptor));
        }
      }
    }, 100);
  };

  const pickFiles = (fileList: FileList | null) => {
    if (!fileList) {
      return;
    }
    setCurrentFace(null);
    Promise.all(
      Array.from(fileList).map((file) =>
        new Promise<PickedFile>((resolve, reject) => {
          const reader = new FileReader();
          reader.readAsDataURL(file);
          reader.onload = () => resolve({name: file.name, base64: String(reader.result)});
          reader.onerror = reject;
        })),
    ).then((picked) => setFiles((prev) => [...prev, ...picked]));
  };

  const generateDescriptor = async() => {
    let maxScore = 0;
    for (const file of files) {
      const img = new Image();
      img.src = file.base64;
      const faceIds = await faceapi
        .detectAllFaces(img, new faceapi.TinyFaceDetectorOptions())
        .withFaceLandmarks()
        .withFaceDescriptors();
      const score = faceIds[0]?.detection.score ?? 0;
      if (score > 0.9 && score > maxScore) {
        maxScore = score;
        setCurrentFace(faceIds[0]);
        setCurrentIndex(files.indexOf(file));
      }
    }
    if (maxScore < 0.9) {
      Setting.showMessage("error", i18next.t("login:Face recognition failed"));
    }
  };

  if (useCamera) {
    return (
      <Dialog open={visible && cameraReady} onOpenChange={(next) => (next ? undefined : onCancel())}>
        <DialogContent className="sm:max-w-[350px]" onInteractOutside={(e) => e.preventDefault()}>
          <DialogHeader>
            <DialogTitle>{i18next.t("login:Face Recognition")}</DialogTitle>
          </DialogHeader>
          <Progress value={percent} />
          <div className="relative mb-12 mt-5 flex justify-center">
            {modelsLoaded ? (
              <>
                <video
                  ref={videoRef}
                  onPlay={handleStreamVideo}
                  className="h-[220px] w-[220px] rounded-full object-cover"
                />
                <FaceRing percent={percent} />
              </>
            ) : (
              <Loading />
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={onCancel}>{i18next.t("general:Cancel")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={visible} onOpenChange={(next) => (next ? undefined : onCancel())}>
      <DialogContent className="sm:max-w-[350px]" onInteractOutside={(e) => e.preventDefault()}>
        <DialogHeader>
          <DialogTitle>{i18next.t("login:Face Recognition")}</DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          <label className="flex cursor-pointer flex-col items-center justify-center rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
            <input type="file" accept="image/*" multiple className="hidden" onChange={(e) => pickFiles(e.target.files)} />
            {i18next.t("general:Click to Upload")}
          </label>
          {files.length > 0 ? (
            <ul className="space-y-1 text-xs text-muted-foreground">
              {files.map((file, index) => <li key={`${file.name}-${index}`}>{file.name}</li>)}
            </ul>
          ) : null}
          {modelsLoaded ? (
            <Button variant="outline" className="w-full" onClick={generateDescriptor}>
              {i18next.t("general:Generate")}
            </Button>
          ) : null}
          {currentFace && currentIndex !== null ? (
            <div className="space-y-2">
              <div className="text-sm">{i18next.t("application:Select")}: {files[currentIndex]?.name}</div>
              <img src={files[currentIndex]?.base64} alt="selected" className="w-full" />
            </div>
          ) : null}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>{i18next.t("general:Cancel")}</Button>
          <Button disabled={!currentFace} onClick={() => onOk(Array.from(currentFace.descriptor))}>
            {i18next.t("general:OK")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
