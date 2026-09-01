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

import React from "react";
import {Alert, Button, QRCode} from "antd";
import copy from "copy-to-clipboard";
import * as Setting from "../Setting";
import i18next from "i18next";

export const generateCasdoorAppUrl = (accessToken, forQrCode = true) => {
  let qrUrl = "";
  let error = null;

  if (!accessToken) {
    error = i18next.t("general:Access token is empty");
    return {qrUrl, error};
  }

  qrUrl = `casdoor-authenticator://login?serverUrl=${window.location.origin}&accessToken=${accessToken}`;

  if (forQrCode && qrUrl.length >= 2000) {
    qrUrl = "";
    error = i18next.t("general:QR code is too large");
  }

  return {qrUrl, error};
};

export const CasdoorAppQrCode = ({accessToken, icon}) => {
  const {qrUrl, error} = generateCasdoorAppUrl(accessToken, true);

  if (error) {
    return <Alert message={error} type="error" showIcon />;
  }

  return (
    <QRCode
      value={qrUrl}
      icon={icon}
      errorLevel="M"
      size={230}
      bordered={false}
    />
  );
};

export const CasdoorAppUrl = ({accessToken}) => {
  const {qrUrl, error} = generateCasdoorAppUrl(accessToken, false);

  const handleCopyUrl = async() => {
    if (!window.isSecureContext) {
      return;
    }

    copy(qrUrl);
    Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
  };

  if (error) {
    return <Alert message={error} type="error" showIcon />;
  }

  return (
    <div>
      <div style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        marginBottom: "10px",
      }}>
        {window.isSecureContext && (
          <Button size="small" type="primary" onClick={handleCopyUrl} style={{marginLeft: "10px"}}>
            {i18next.t("resource:Copy Link")}
          </Button>
        )}
      </div>
      <div
        style={{
          padding: "10px",
          maxWidth: "400px",
          maxHeight: "100px",
          overflow: "auto",
          wordBreak: "break-all",
          whiteSpace: "pre-wrap",
          cursor: "pointer",
          userSelect: "all",
          backgroundColor: "#f5f5f5",
          borderRadius: "4px",
        }}
        onClick={(e) => {
          const selection = window.getSelection();
          const range = document.createRange();
          range.selectNodeContents(e.target);
          selection.removeAllRanges();
          selection.addRange(range);
        }}
      >
        {qrUrl}
      </div>
    </div>
  );
};
