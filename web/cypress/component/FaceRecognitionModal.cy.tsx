import * as React from "react";
import {mount} from "cypress/react18";
import * as faceapi from "face-api.js";
import i18next from "i18next";
import {FaceRecognitionModal} from "../../src/components/common/FaceRecognitionModal";
import "../../src/index.css";

function Harness({unmountOnClose = false, captureImage = false}: {unmountOnClose?: boolean; captureImage?: boolean}) {
  const [visible, setVisible] = React.useState(true);
  const [mounted, setMounted] = React.useState(true);
  return (
    <>
      <button onClick={() => setVisible(true)}>Reopen</button>
      <button onClick={() => setMounted(false)}>Leave page</button>
      {mounted && (!unmountOnClose || visible) ? (
        <FaceRecognitionModal
          visible={visible}
          withImage={captureImage}
          captureImage={captureImage}
          onOk={() => setVisible(false)}
          onCancel={() => setVisible(false)}
        />
      ) : null}
    </>
  );
}

describe("Face recognition camera lifecycle", () => {
  let streams: MediaStream[];

  beforeEach(() => {
    streams = [];
    // Exercise the real Dialog portal and video playback without a physical
    // camera, network model downloads, biometric data or a backend.
    cy.then(() => i18next.init({lng: "en", resources: {en: {general: {Cancel: "Cancel"}, login: {"Face Recognition": "Face Recognition"}}}}));
    cy.stub(faceapi.nets.tinyFaceDetector, "loadFromUri").resolves();
    cy.stub(faceapi.nets.faceLandmark68Net, "loadFromUri").resolves();
    cy.stub(faceapi.nets.faceRecognitionNet, "loadFromUri").resolves();
    cy.stub(faceapi.nets.tinyFaceDetector, "locateFaces").resolves([]).as("detect");
    cy.window().then((win) => {
      cy.stub(win.navigator.mediaDevices, "getUserMedia").callsFake(() => {
        const canvas = win.document.createElement("canvas");
        canvas.width = 320;
        canvas.height = 240;
        const stream = canvas.captureStream();
        canvas.getContext("2d")!.fillRect(0, 0, 320, 240);
        streams.push(stream);
        return Promise.resolve(stream);
      }).as("camera");
    });
  });

  afterEach(() => {
    streams.forEach((stream) => stream.getTracks().forEach((track) => track.stop()));
  });

  function expectPreview(index = 0) {
    cy.get("[role=dialog] video").should(($video) => {
      const video = $video[0] as HTMLVideoElement;
      expect(video.srcObject).to.equal(streams[index]);
      expect(video.videoWidth).to.equal(320);
      expect(video.paused).to.equal(false);
      expect(video.muted).to.equal(true);
      expect(video.playsInline).to.equal(true);
    });
    cy.get("@detect").should("have.been.called");
  }

  function expectStopped(index = 0) {
    cy.wrap(null).should(() => {
      expect(streams[index].getTracks().every((track) => track.readyState === "ended")).to.equal(true);
    });
  }

  it("plays after the portal mounts and releases the camera when enrollment unmounts", () => {
    mount(<Harness unmountOnClose />);
    expectPreview();
    cy.contains("button", "Cancel").click();
    expectStopped();
  });

  it("releases the acquired stream on unmount even if playback has not started", () => {
    cy.window().then((win) => {
      cy.stub(win.HTMLMediaElement.prototype, "play").returns(new Promise<void>(() => {}));
    });
    mount(<Harness unmountOnClose />);
    cy.get("[role=dialog] video").should("exist");
    cy.contains("button", "Cancel").click();
    expectStopped();
  });

  it("releases and reacquires the stream when a mounted login modal closes and reopens", () => {
    mount(<Harness />);
    expectPreview();
    cy.contains("button", "Cancel").click();
    expectStopped();
    cy.contains("button", "Reopen").click();
    expectPreview(1);
    cy.get("@camera").should("have.been.calledTwice");
    cy.contains("button", "Cancel").click();
    expectStopped(1);
  });

  it("also plays in camera image capture mode", () => {
    mount(<Harness captureImage unmountOnClose />);
    expectPreview();
    cy.contains("button", "Cancel").click();
    expectStopped();
  });

  it("stops a stream whose permission request resolves after leaving the page", () => {
    let resolveCamera: (stream: MediaStream) => void;
    cy.get("@camera").then((stub: any) => {
      stub.callsFake(() => new Promise<MediaStream>((resolve) => { resolveCamera = resolve; }));
    });
    mount(<Harness />);
    cy.get("@camera").should("have.been.calledOnce");
    cy.contains("button", "Leave page").click();
    cy.window().then((win) => {
      const stream = win.document.createElement("canvas").captureStream();
      streams.push(stream);
      resolveCamera(stream);
    });
    expectStopped();
    cy.get("[role=dialog]").should("not.exist");
  });

  it("handles playback rejection and releases the camera", () => {
    cy.window().then((win) => {
      cy.stub(win.HTMLMediaElement.prototype, "play").rejects(new DOMException("Playback denied", "NotAllowedError"));
    });
    mount(<Harness unmountOnClose />);
    expectStopped();
    cy.get("[role=dialog]").should("not.exist");
  });
});
