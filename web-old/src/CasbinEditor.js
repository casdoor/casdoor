// Copyright 2024 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React, {useCallback, useEffect, useRef, useState} from "react";
import * as Setting from "./Setting";
import IframeEditor from "./IframeEditor";
import {Tabs} from "antd";
import i18next from "i18next";
import Editor from "./common/Editor";

const CasbinEditor = ({model, onModelTextChange}) => {
  const [activeKey, setActiveKey] = useState("advanced");
  const iframeRef = useRef(null);
  const [localModelText, setLocalModelText] = useState(model.modelText);

  const handleModelTextChange = useCallback((newModelText) => {
    if (!Setting.builtInObject(model)) {
      setLocalModelText(newModelText);
      onModelTextChange(newModelText);
    }
  }, [model, onModelTextChange]);

  const syncModelText = useCallback(() => {
    return new Promise((resolve) => {
      if (activeKey === "advanced" && iframeRef.current) {
        const handleSyncMessage = (event) => {
          if (event.data.type === "modelUpdate") {
            window.removeEventListener("message", handleSyncMessage);
            handleModelTextChange(event.data.modelText);
            resolve();
          }
        };
        window.addEventListener("message", handleSyncMessage);
        iframeRef.current.getModelText();
      } else {
        resolve();
      }
    });
  }, [activeKey, handleModelTextChange]);

  const handleTabChange = (key) => {
    syncModelText().then(() => {
      setActiveKey(key);
      if (key === "advanced" && iframeRef.current) {
        iframeRef.current.updateModelText(localModelText);
      }
    });
  };

  useEffect(() => {
    setLocalModelText(model.modelText);
  }, [model.modelText]);

  return (
    <div style={{height: "100%", width: "100%", display: "flex", flexDirection: "column"}}>
      <Tabs
        activeKey={activeKey}
        onChange={handleTabChange}
        style={{flex: "0 0 auto", marginTop: "-10px"}}
        items={[
          {key: "basic", label: i18next.t("model:Basic Editor")},
          {key: "advanced", label: i18next.t("model:Advanced Editor")},
        ]}
      />
      <div style={{flex: "1 1 auto", overflow: "hidden"}}>
        {activeKey === "advanced" ? (
          <IframeEditor
            ref={iframeRef}
            initialModelText={localModelText}
            onModelTextChange={handleModelTextChange}
            style={{width: "100%", height: "100%"}}
          />
        ) : (
          <Editor
            value={localModelText}
            readOnly={Setting.builtInObject(model)}
            onChange={value => {
              handleModelTextChange(value);
            }}
          />
        )}
      </div>
    </div>
  );
};

export default CasbinEditor;
