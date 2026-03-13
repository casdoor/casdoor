// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
import {Alert, Button, Input, Modal} from "antd";
import {ExportOutlined, ImportOutlined} from "@ant-design/icons";
import copy from "copy-to-clipboard";
import * as Setting from "./Setting";

export const ApplicationImportExport = {
  renderImportExportButtons(_application, onExportJson, onShowImportModal) {
    return (
      <>
        <Button style={{marginLeft: "20px"}} icon={<ExportOutlined />} onClick={onExportJson} data-cy="export-json-button">
          Export JSON
        </Button>
        <Button style={{marginLeft: "10px"}} icon={<ImportOutlined />} onClick={onShowImportModal} data-cy="import-json-button">
          Import JSON
        </Button>
      </>
    );
  },

  exportApplicationJson(application) {
    if (!application) {
      Setting.showMessage("error", "No application to export");
      return;
    }

    const copyDefinedFields = (source, target, fields) => {
      fields.forEach(field => {
        if (source[field] !== undefined) {
          target[field] = source[field];
        }
      });
    };

    const minimalApp = {};
    const appFields = [
      "name",
      "displayName",
      "organization",
      "logo",
      "homepageUrl",
      "description",
      "enablePassword",
      "enableSignUp",
      "disableSignin",
      "enableCodeSignin",
      "enableWebAuthn",
      "themeData",
      "formCss",
      "formSideHtml",
      "headerHtml",
      "footerHtml",
      "signupHtml",
      "signinHtml",
    ];
    copyDefinedFields(application, minimalApp, appFields);

    if (application.providers && application.providers.length > 0) {
      const providerOptionalFields = [
        "canSignUp",
        "canSignIn",
        "canUnlink",
        "rule",
        "prompted",
        "signupGroup",
      ];
      minimalApp.providers = application.providers.map(p => {
        const item = {name: p.name};
        copyDefinedFields(p, item, providerOptionalFields);
        return item;
      });
    }

    if (application.signinItems && application.signinItems.length > 0) {
      const signinItemFields = [
        "visible",
        "label",
        "customCss",
        "placeholder",
        "rule",
        "isCustom",
      ];
      minimalApp.signinItems = application.signinItems.map(item => {
        const newItem = {name: item.name};
        copyDefinedFields(item, newItem, signinItemFields);
        return newItem;
      });
    }

    if (application.signupItems && application.signupItems.length > 0) {
      const signupItemFields = [
        "visible",
        "required",
        "prompted",
        "label",
        "customCss",
        "placeholder",
        "rule",
      ];
      minimalApp.signupItems = application.signupItems.map(item => {
        const newItem = {name: item.name};
        copyDefinedFields(item, newItem, signupItemFields);
        return newItem;
      });
    }

    if (application.signinMethods && application.signinMethods.length > 0) {
      minimalApp.signinMethods = application.signinMethods;
    }

    const jsonStr = JSON.stringify(minimalApp, null, 2);
    copy(jsonStr);
    Setting.showMessage("success", "Copied to clipboard successfully");
  },

  importApplicationJson(importJson, existingApplication, onImportSuccess) {
    const jsonStr = importJson.trim();
    if (!jsonStr) {
      Setting.showMessage("error", "Please paste JSON content");
      return;
    }

    let appData;
    try {
      appData = JSON.parse(jsonStr);
      if (typeof appData !== "object" || appData === null || Array.isArray(appData)) {
        Setting.showMessage("error", "Invalid JSON format");
        return;
      }
    } catch (e) {
      Setting.showMessage("error", "Invalid JSON format: " + e.message);
      return;
    }

    if (!existingApplication) {
      Setting.showMessage("error", "No existing application to import into");
      return;
    }

    // Verify name and organization match current application (prevent hijacking other apps)
    if (appData.name && appData.name !== existingApplication.name) {
      Setting.showMessage("error", "Failed to import: Name does not match");
      return;
    }
    if (appData.organization && appData.organization !== existingApplication.organization) {
      Setting.showMessage("error", "Failed to import: Organization does not match");
      return;
    }

    const mergedApp = Setting.deepCopy(existingApplication);

    // Filter out dangerous keys to prevent prototype pollution
    const dangerousKeys = ["__proto__", "constructor", "prototype"];
    Object.keys(appData).forEach(key => {
      if (dangerousKeys.includes(key)) {
        return;
      }
      // Do not allow import to change application identity fields
      if (key === "name" || key === "organization") {
        return;
      }
      mergedApp[key] = appData[key];
    });

    onImportSuccess(mergedApp);
  },

  renderImportModal(visible, importJson, onImportJsonChange, onImport, onCancel) {
    return (
      <Modal
        title="Import JSON"
        open={visible}
        onOk={onImport}
        onCancel={onCancel}
        width={700}
        okText="Import"
        cancelText="Cancel"
      >
        <Alert
          message="Warning"
          description="This will import all application settings including theme, authentication methods, providers and other configurations. Ensure the JSON is from a trusted source."
          type="warning"
          style={{marginBottom: "15px"}}
          showIcon
        />
        <p style={{marginBottom: "10px"}}>Paste JSON content here</p>
        <Input.TextArea
          rows={15}
          value={importJson}
          onChange={(e) => onImportJsonChange(e.target.value)}
          placeholder={JSON.stringify({displayName: "My Custom App", enablePassword: true, enableSignUp: false, themeData: {colorPrimary: "#1890ff"}}, null, 2)}
        />
      </Modal>
    );
  },
};
