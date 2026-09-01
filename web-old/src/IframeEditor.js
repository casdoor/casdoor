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

import React, {forwardRef, useEffect, useImperativeHandle, useRef, useState} from "react";

const IframeEditor = forwardRef(({initialModelText, onModelTextChange}, ref) => {
  const iframeRef = useRef(null);
  const [iframeReady, setIframeReady] = useState(false);
  const currentLang = localStorage.getItem("language") || "en";

  useEffect(() => {
    const handleMessage = (event) => {
      if (event.origin !== "https://editor.casbin.org") {return;}

      if (event.data.type === "modelUpdate") {
        onModelTextChange(event.data.modelText);
      } else if (event.data.type === "iframeReady") {
        setIframeReady(true);
        if (initialModelText && iframeRef.current?.contentWindow) {
          iframeRef.current.contentWindow.postMessage({
            type: "initializeModel",
            modelText: initialModelText,
            lang: currentLang,
          }, "*");
        }
      }
    };

    window.addEventListener("message", handleMessage);
    return () => window.removeEventListener("message", handleMessage);
  }, [onModelTextChange, initialModelText, currentLang]);

  useImperativeHandle(ref, () => ({
    getModelText: () => {
      if (iframeRef.current?.contentWindow) {
        iframeRef.current.contentWindow.postMessage({
          type: "getModelText",
        }, "*");
      }
    },
    updateModelText: (newModelText) => {
      if (iframeReady && iframeRef.current?.contentWindow) {
        iframeRef.current.contentWindow.postMessage({
          type: "updateModelText",
          modelText: newModelText,
        }, "*");
      }
    },
  }));

  return (
    <iframe
      ref={iframeRef}
      src={`https://editor.casbin.org/model-editor?lang=${currentLang}`}
      frameBorder="0"
      width="100%"
      height="500px"
      title="Casbin Model Editor"
    />
  );
});

export default IframeEditor;
