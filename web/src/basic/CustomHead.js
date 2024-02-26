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

function CustomHead(props) {
  useEffect(() => {
    const suffix = new Date().getTime().toString();

    if (!props.headerHtml) {return;}
    const node = document.createElement("div");
    node.innerHTML = props.headerHtml;

    node.childNodes.forEach(el => {
      try {
        el.setAttribute("app-custom-head" + suffix, "");
      } catch {
        document.head.appendChild(el);
        return;
      }
      if (el.localName === "script") {
        const node = document.createElement("script");
        Array.from(el.attributes).forEach(attr => {
          node.setAttribute(attr.name, attr.value);
        });
        node.text = el.textContent;
        document.head.appendChild(node);
        return;
      }
      document.head.appendChild(el);
    });

    return () => {
      for (const el of document.head.children) {
        if (el.getAttribute("app-custom-head" + suffix) !== null) {
          document.head.removeChild(el);
        }
      }
    };
  });
}

export default CustomHead;
