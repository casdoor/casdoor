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

import {useEffect} from "react";

let customHeadLoaded = false;

function CustomHead(props) {
  useEffect(() => {
    if (!customHeadLoaded) {
      const suffix = new Date().getTime().toString();

      if (!props.headerHtml) {return;}
      const node = document.createElement("div");
      node.innerHTML = props.headerHtml;

      node.childNodes.forEach(el => {
        if (el.nodeName === "#text") {
          return;
        }
        let innerNode = el;
        innerNode.setAttribute("app-custom-head" + suffix, "");

        if (innerNode.localName === "script") {
          const scriptNode = document.createElement("script");
          Array.from(innerNode.attributes).forEach(attr => {
            scriptNode.setAttribute(attr.name, attr.value);
          });
          scriptNode.text = innerNode.textContent;
          innerNode = scriptNode;
        }
        document.head.appendChild(innerNode);
      });
      customHeadLoaded = true;
    }
  });
}

export default CustomHead;
