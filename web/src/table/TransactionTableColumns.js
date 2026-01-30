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
import {Link} from "react-router-dom";
import * as Setting from "../Setting";
import i18next from "i18next";
import {Button} from "antd";
import PopconfirmModal from "../common/modal/PopconfirmModal";

export function getTransactionTableColumns(options = {}) {
  const {
    includeOrganization = false,
    includeUser = false,
    includeTag = true,
    includeActions = false,
    getColumnSearchProps = null,
    account = null,
    onEdit = null,
    onDelete = null,
  } = options;

  const columns = [];

  // Use function-based sorter for client-side, boolean for server-side
  const getSorter = (dataIndex) => {
    if (includeActions) {
      return true; // Server-side sorting
    } else if (getColumnSearchProps) {
      // Client-side sorting
      return (a, b) => {
        const aVal = a[dataIndex] || "";
        const bVal = b[dataIndex] || "";
        return aVal.toString().localeCompare(bVal.toString());
      };
    }
    return false;
  };

  if (includeOrganization) {
    columns.push({
      title: i18next.t("general:Organization"),
      dataIndex: "owner",
      key: "owner",
      width: "120px",
      fixed: "left",
      sorter: getSorter("owner"),
      ...(getColumnSearchProps ? getColumnSearchProps("owner") : {}),
      render: (text, record, index) => {
        return (
          <Link to={`/organizations/${text}`}>
            {text}
          </Link>
        );
      },
    });
  }

  columns.push({
    title: i18next.t("general:Name"),
    dataIndex: "name",
    key: "name",
    width: includeOrganization ? "180px" : "280px",
    fixed: includeOrganization ? "left" : false,
    sorter: getSorter("name"),
    ...(getColumnSearchProps ? getColumnSearchProps("name") : {}),
    render: (text, record, index) => {
      return (
        <Link to={`/transactions/${record.owner}/${record.name}`}>
          {text}
        </Link>
      );
    },
  });

  columns.push({
    title: i18next.t("general:Created time"),
    dataIndex: "createdTime",
    key: "createdTime",
    width: "160px",
    sorter: getSorter("createdTime"),
    render: (text, record, index) => {
      return Setting.getFormattedDate(text);
    },
  });

  if (includeTag) {
    columns.push({
      title: i18next.t("user:Tag"),
      dataIndex: "tag",
      key: "tag",
      width: "120px",
      sorter: getSorter("tag"),
      ...(getColumnSearchProps ? getColumnSearchProps("tag") : {}),
    });
  }

  if (includeUser) {
    columns.push({
      title: i18next.t("general:User"),
      dataIndex: "user",
      key: "user",
      width: "120px",
      sorter: getSorter("user"),
      ...(getColumnSearchProps ? getColumnSearchProps("user") : {}),
      render: (text, record, index) => {
        if (!text || Setting.isAnonymousUserName(text)) {
          return text;
        }

        return (
          <Link to={`/users/${record.owner}/${text}`}>
            {text}
          </Link>
        );
      },
    });
  }

  columns.push({
    title: i18next.t("general:Application"),
    dataIndex: "application",
    key: "application",
    width: "150px",
    sorter: getSorter("application"),
    ...(getColumnSearchProps ? getColumnSearchProps("application") : {}),
    render: (text, record, index) => {
      if (!text) {
        return text;
      }
      return (
        <Link to={`/applications/${record.owner}/${record.application}`}>
          {text}
        </Link>
      );
    },
  });

  columns.push({
    title: i18next.t("provider:Domain"),
    dataIndex: "domain",
    key: "domain",
    width: includeOrganization ? "200px" : "270px",
    sorter: getSorter("domain"),
    ...(getColumnSearchProps ? getColumnSearchProps("domain") : {}),
    render: (text, record, index) => {
      if (!text) {
        return null;
      }

      return (
        <a href={text} target="_blank" rel="noopener noreferrer">
          {text}
        </a>
      );
    },
  });

  columns.push({
    title: i18next.t("provider:Category"),
    dataIndex: "category",
    key: "category",
    width: "120px",
    sorter: getSorter("category"),
    ...(getColumnSearchProps ? getColumnSearchProps("category") : {}),
  });

  columns.push({
    title: i18next.t("provider:Type"),
    dataIndex: "type",
    key: "type",
    width: "140px",
    sorter: getSorter("type"),
    ...(getColumnSearchProps ? getColumnSearchProps("type") : {}),
    render: (text, record, index) => {
      if (text && record.domain) {
        const chatUrl = `${record.domain}/chats/${text}`;
        return (
          <a href={chatUrl} target="_blank" rel="noopener noreferrer">
            {text}
          </a>
        );
      }
      return text;
    },
  });

  columns.push({
    title: i18next.t("provider:Subtype"),
    dataIndex: "subtype",
    key: "subtype",
    width: "140px",
    sorter: getSorter("subtype"),
    ...(getColumnSearchProps ? getColumnSearchProps("subtype") : {}),
    render: (text, record, index) => {
      if (text && record.domain) {
        const messageUrl = `${record.domain}/messages/${text}`;
        return (
          <a href={messageUrl} target="_blank" rel="noopener noreferrer">
            {text}
          </a>
        );
      }
      return text;
    },
  });

  columns.push({
    title: i18next.t("general:Provider"),
    dataIndex: "provider",
    key: "provider",
    width: "150px",
    sorter: getSorter("provider"),
    ...(getColumnSearchProps ? getColumnSearchProps("provider") : {}),
    render: (text, record, index) => {
      if (!text) {
        return text;
      }
      if (record.domain) {
        const casibaseUrl = `${record.domain}/providers/${text}`;
        return (
          <a href={casibaseUrl} target="_blank" rel="noopener noreferrer">
            {text}
          </a>
        );
      }
      return (
        <Link to={`/providers/${record.owner}/${text}`}>
          {text}
        </Link>
      );
    },
  });

  columns.push({
    title: i18next.t("general:Payment"),
    dataIndex: "payment",
    key: "payment",
    width: "120px",
    sorter: getSorter("payment"),
    ...(getColumnSearchProps ? getColumnSearchProps("payment") : {}),
    render: (text, record, index) => {
      if (!text) {
        return text;
      }
      return (
        <Link to={`/payments/${record.owner}/${text}`}>
          {text}
        </Link>
      );
    },
  });

  columns.push({
    title: i18next.t("general:State"),
    dataIndex: "state",
    key: "state",
    width: "120px",
    sorter: getSorter("state"),
    ...(getColumnSearchProps ? getColumnSearchProps("state") : {}),
  });

  columns.push({
    title: i18next.t("transaction:Amount"),
    dataIndex: "amount",
    key: "amount",
    width: "180px",
    sorter: getSorter("amount"),
    ...(getColumnSearchProps ? getColumnSearchProps("amount") : {}),
    fixed: (Setting.isMobile()) ? "false" : "right",
    render: (text, record, index) => {
      return Setting.getPriceDisplay(record.amount, record.currency);
    },
  });

  if (includeActions && account && onEdit && onDelete) {
    columns.push({
      title: i18next.t("general:Action"),
      dataIndex: "",
      key: "op",
      width: "200px",
      fixed: (Setting.isMobile()) ? "false" : "right",
      render: (text, record, index) => {
        const isAdmin = Setting.isLocalAdminUser(account);
        return (
          <div>
            <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => onEdit(record, isAdmin)}>{isAdmin ? i18next.t("general:Edit") : i18next.t("general:View")}</Button>
            <PopconfirmModal
              title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
              onConfirm={() => onDelete(index)}
              disabled={!isAdmin}
            >
            </PopconfirmModal>
          </div>
        );
      },
    });
  }

  return columns;
}
