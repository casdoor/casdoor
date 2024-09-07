import React, {useCallback, useEffect, useRef} from "react";
import {Controlled as CodeMirror} from "react-codemirror2";
import "codemirror/lib/codemirror.css";
import "codemirror/mode/properties/properties";
import * as Setting from "./Setting";
import IframeEditor from "./IframeEditor";

const CasbinEditor = ({model, useIframeEditor, onModelTextChange, onSubmit}) => {
  const iframeRef = useRef(null);

  const handleModelTextChange = useCallback((newModelText) => {
    if (!Setting.builtInObject(model)) {
      onModelTextChange(newModelText);
    }
  }, [model, onModelTextChange]);

  const submitModelEdit = useCallback(() => {
    if (useIframeEditor && iframeRef.current) {
      return new Promise((resolve) => {
        const handleSubmitMessage = (event) => {
          if (event.data.type === "modelUpdate") {
            window.removeEventListener("message", handleSubmitMessage);
            handleModelTextChange(event.data.modelText);
            resolve();
          }
        };
        window.addEventListener("message", handleSubmitMessage);
        iframeRef.current.getModelText();
      });
    } else {
      return Promise.resolve();
    }
  }, [useIframeEditor, handleModelTextChange]);

  useEffect(() => {
    onSubmit(submitModelEdit);
  }, [onSubmit, submitModelEdit]);

  if (useIframeEditor) {
    return (
      <IframeEditor
        ref={iframeRef}
        modelText={model.modelText}
        onModelTextChange={handleModelTextChange}
      />
    );
  }

  return (
    <div style={{height: "100%", width: "100%"}}>
      <CodeMirror
        value={model.modelText}
        className="full-height-editor"
        options={{mode: "properties", theme: "default"}}
        onBeforeChange={(editor, data, value) => {
          handleModelTextChange(value);
        }}
      />
    </div>
  );
};

export default CasbinEditor;
