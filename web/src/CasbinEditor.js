import React, {useCallback, useEffect, useRef, useState} from "react";
import {Controlled as CodeMirror} from "react-codemirror2";
import "codemirror/lib/codemirror.css";
import "codemirror/mode/properties/properties";
import * as Setting from "./Setting";
import IframeEditor from "./IframeEditor";
import {Tabs} from "antd";

const {TabPane} = Tabs;

const CasbinEditor = ({model, initialUseIframeEditor, onModelTextChange, onSubmit}) => {
  const [activeKey, setActiveKey] = useState(initialUseIframeEditor ? "advanced" : "basic");
  const iframeRef = useRef(null);
  const [localModelText, setLocalModelText] = useState(model.modelText);

  const handleModelTextChange = useCallback((newModelText) => {
    if (!Setting.builtInObject(model)) {
      setLocalModelText(newModelText);
      onModelTextChange(newModelText);
    }
  }, [model, onModelTextChange]);

  const submitModelEdit = useCallback(() => {
    if (activeKey === "advanced" && iframeRef.current) {
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
  }, [activeKey, handleModelTextChange]);

  useEffect(() => {
    onSubmit(submitModelEdit);
  }, [onSubmit, submitModelEdit]);

  const handleTabChange = (key) => {
    if (activeKey === "advanced" && key === "basic") {
      submitModelEdit().then(() => {
        setActiveKey(key);
      });
    } else {
      setActiveKey(key);
      if (key === "advanced" && iframeRef.current) {
        iframeRef.current.updateModelText(localModelText);
      }
    }
  };

  return (
    <div style={{height: "100%", width: "100%", display: "flex", flexDirection: "column"}}>
      <Tabs activeKey={activeKey} onChange={handleTabChange} style={{flex: "0 0 auto", marginTop: "-10px"}}>
        <TabPane tab="Basic Editor" key="basic" />
        <TabPane tab="Advanced Editor" key="advanced" />
      </Tabs>
      <div style={{flex: "1 1 auto", overflow: "hidden"}}>
        {activeKey === "advanced" ? (
          <IframeEditor
            ref={iframeRef}
            modelText={localModelText}
            onModelTextChange={handleModelTextChange}
            style={{width: "100%", height: "100%"}}
          />
        ) : (
          <CodeMirror
            value={localModelText}
            className="full-height-editor no-horizontal-scroll"
            options={{mode: "properties", theme: "default"}}
            onBeforeChange={(editor, data, value) => {
              handleModelTextChange(value);
            }}
          />
        )}
      </div>
    </div>
  );
};

export default CasbinEditor;
