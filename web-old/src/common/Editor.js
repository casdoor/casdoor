// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

import React from "react";
import CodeMirror from "@uiw/react-codemirror";
import {materialDark} from "@uiw/codemirror-theme-material";
import {langs} from "@uiw/codemirror-extensions-langs";

const defaultFrameStyle = {
  borderRadius: 10,
  overflow: "hidden",
  border: "1px solid #e2e8f0",
  boxShadow: "0 1px 2px rgba(15, 23, 42, 0.05), 0 4px 12px rgba(15, 23, 42, 0.06)",
};

export const Editor = (props) => {
  const {
    frameless = false,
    wrapperStyle: userWrapperStyle,
    fillWidth,
    fillHeight,
    dark,
    lang,
    value,
    onChange,
    readOnly,
    style: userStyle,
    maxWidth,
    minWidth,
    maxHeight,
    minHeight,
    ...rest
  } = props;

  let height = props.height;
  let width = props.width;
  let style = {};
  const copy2StyleProps = [
    "width", "maxWidth", "minWidth",
    "height", "maxHeight", "minHeight",
  ];
  if (fillHeight) {
    height = "100%";
    style = {...style, height: "100%"};
  }
  if (fillWidth) {
    width = "100%";
    style = {...style, width: "100%"};
  }
  /**
   * @uiw/react-codemirror style props sucha as "height" "width"
   * may need to be configured with "style" in some scenarios to take effect
   */
  copy2StyleProps.forEach(el => {
    if (["number", "string"].includes(typeof props[el])) {
      style = {...style, [el]: props[el]};
    }
  });
  if (!frameless) {
    const {maxWidth: _omitMw, minWidth: _omitMi, ...innerRest} = style;
    style = {...innerRest, borderRadius: 0};
  }
  if (userStyle) {
    style = {...style, ...userStyle};
  }

  let extensions = [];
  switch (lang) {
  case "javascript":
  case "js":
    extensions = [langs.javascript()];
    break;
  case "html":
    extensions = [langs.html()];
    break;
  case "css":
    extensions = [langs.css()];
    break;
  case "xml":
    extensions = [langs.xml()];
    break;
  case "json":
    extensions = [langs.json()];
    break;
  }

  const codeMirror = (
    <CodeMirror
      {...rest}
      value={value}
      width={width}
      height={height}
      style={style}
      readOnly={readOnly}
      theme={dark ? materialDark : "light"}
      extensions={extensions}
      onChange={onChange}
    />
  );

  if (frameless) {
    return codeMirror;
  }

  const outerStyle = {
    boxSizing: "border-box",
    ...defaultFrameStyle,
    ...userWrapperStyle,
  };
  if (fillWidth) {
    outerStyle.width = "100%";
  }
  if (["number", "string"].includes(typeof maxWidth)) {
    outerStyle.maxWidth = maxWidth;
  }
  if (["number", "string"].includes(typeof minWidth)) {
    outerStyle.minWidth = minWidth;
  }
  if (fillHeight) {
    outerStyle.height = "100%";
    outerStyle.display = "flex";
    outerStyle.flexDirection = "column";
    if (["number", "string"].includes(typeof minHeight)) {
      outerStyle.minHeight = minHeight;
    } else {
      outerStyle.minHeight = 0;
    }
  } else if (["number", "string"].includes(typeof minHeight)) {
    outerStyle.minHeight = minHeight;
  }
  if (["number", "string"].includes(typeof maxHeight)) {
    outerStyle.maxHeight = maxHeight;
  }

  return (
    <div style={outerStyle}>
      {codeMirror}
    </div>
  );
};

export default Editor;
