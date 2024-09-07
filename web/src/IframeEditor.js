import React, {forwardRef, useEffect, useImperativeHandle, useRef} from "react";

const IframeEditor = forwardRef(({modelText, onModelTextChange}, ref) => {
  const iframeRef = useRef(null);

  useEffect(() => {
    const handleMessage = (event) => {
      if (event.origin !== "http://editor.casbin.org") {
        return;
      }
      if (event.data.type === "modelUpdate") {
        onModelTextChange(event.data.modelText);
      }
    };

    window.addEventListener("message", handleMessage);
    return () => {
      window.removeEventListener("message", handleMessage);
    };
  }, [onModelTextChange]);

  useImperativeHandle(ref, () => ({
    getModelText: () => {
      iframeRef.current?.contentWindow.postMessage({type: "getModelText"}, "*");
    },
  }));

  return (
    <iframe
      ref={iframeRef}
      src={`http://editor.casbin.org/model-editor?model=${encodeURIComponent(modelText)}`}
      frameBorder="0"
      width="100%"
      height="500px"
      title="Casbin Model Editor"
    />
  );
});

export default IframeEditor;
