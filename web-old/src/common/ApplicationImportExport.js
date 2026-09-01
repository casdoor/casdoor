// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

import React, {useState} from "react";
import {Button, Input, Modal} from "antd";
import {UploadOutlined} from "@ant-design/icons";
import copy from "copy-to-clipboard";
import i18next from "i18next";
import * as Setting from "../Setting";

// Fields that represent UI/theme customization and are safe to transfer between applications.
const UI_FIELDS = [
  "logo",
  "favicon",
  "formBackgroundUrl",
  "formBackgroundUrlMobile",
  "formCss",
  "formCssMobile",
  "formOffset",
  "formSideHtml",
  "themeData",
  "headerHtml",
  "pageHtml",
  "footerHtml",
  "signupHtml",
  "signinHtml",
];

export function exportApplicationJson(application) {
  const payload = {
    name: application.name,
    organization: application.organization,
  };
  for (const key of UI_FIELDS) {
    if (application[key] !== undefined) {
      payload[key] = application[key];
    }
  }
  const json = JSON.stringify(payload, null, 2);
  copy(json);
  Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
}

export function ApplicationImportModal({application, onImport}) {
  const [visible, setVisible] = useState(false);
  const [jsonText, setJsonText] = useState("");

  function handleOk() {
    let parsed;
    try {
      parsed = JSON.parse(jsonText);
    } catch (e) {
      Setting.showMessage("error", e.message);
      return;
    }

    if (parsed.name !== application.name || parsed.organization !== application.organization) {
      Setting.showMessage("error", i18next.t("general:Invalid application"));
      return;
    }

    const updates = {};
    for (const key of UI_FIELDS) {
      if (Object.prototype.hasOwnProperty.call(parsed, key)) {
        updates[key] = parsed[key];
      }
    }

    onImport(updates);
    Setting.showMessage("success", i18next.t("general:Successfully modified"));
    setVisible(false);
    setJsonText("");
  }

  function handleCancel() {
    setVisible(false);
    setJsonText("");
  }

  return (
    <>
      <Button style={{marginLeft: "20px"}} icon={<UploadOutlined />} onClick={() => setVisible(true)}>
        {i18next.t("application:Import JSON")}
      </Button>
      <Modal
        title={i18next.t("application:Import JSON")}
        open={visible}
        onOk={handleOk}
        onCancel={handleCancel}
        okText={i18next.t("general:OK")}
        cancelText={i18next.t("general:Cancel")}
        width={600}
      >
        <p>{i18next.t("application:Import JSON - description")}</p>
        <Input.TextArea
          rows={12}
          value={jsonText}
          onChange={e => setJsonText(e.target.value)}
          placeholder="{ ... }"
        />
      </Modal>
    </>
  );
}
