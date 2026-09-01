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

import React from "react";
import {Breadcrumb} from "antd";
import {Link} from "react-router-dom";
import i18next from "i18next";

const RESOURCE_LABELS = {
  "apps": "general:Apps",
  "shortcuts": "general:Shortcuts",
  "account": "account:My Account",
  "organizations": "general:Organizations",
  "users": "general:Users",
  "groups": "general:Groups",
  "trees": "general:Groups",
  "invitations": "general:Invitations",
  "applications": "general:Applications",
  "providers": "application:Providers",
  "resources": "general:Resources",
  "certs": "general:Certs",
  "keys": "general:Keys",
  "agents": "general:Agents",
  "servers": "general:MCP Servers",
  "server-store": "general:MCP Store",
  "entries": "general:Entries",
  "sites": "general:Sites",
  "rules": "general:Rules",
  "roles": "general:Roles",
  "permissions": "general:Permissions",
  "models": "general:Models",
  "adapters": "general:Adapters",
  "enforcers": "general:Enforcers",
  "sessions": "general:Sessions",
  "records": "general:Records",
  "tokens": "general:Tokens",
  "verifications": "general:Verifications",
  "product-store": "general:Product Store",
  "products": "general:Products",
  "cart": "general:Cart",
  "orders": "general:Orders",
  "payments": "general:Payments",
  "plans": "general:Plans",
  "pricings": "general:Pricings",
  "subscriptions": "general:Subscriptions",
  "transactions": "general:Transactions",
  "sysinfo": "general:System Info",
  "forms": "general:Forms",
  "syncers": "general:Syncers",
  "webhooks": "general:Webhooks",
  "webhook-events": "general:Webhook Events",
  "tickets": "general:Tickets",
  "ldap": "general:LDAP",
  "mfa": "general:MFA",
};

function buildBreadcrumbItems(uri) {
  const pathSegments = (uri || "").split("/").filter(Boolean);

  const homeItem = {title: <Link to="/">{i18next.t("general:Home")}</Link>};

  if (pathSegments.length === 0) {
    return null;
  }

  const rootSegment = pathSegments[0];
  const listLabelKey = RESOURCE_LABELS[rootSegment];
  if (!listLabelKey) {
    return null;
  }

  if (pathSegments.length === 1) {
    return [
      homeItem,
      {title: i18next.t(listLabelKey)},
    ];
  }

  const lastSegment = pathSegments[pathSegments.length - 1];
  const lastLabelKey = RESOURCE_LABELS[lastSegment];
  const lastLabel = lastLabelKey ? i18next.t(lastLabelKey) : lastSegment;

  return [
    homeItem,
    {title: <Link to={`/${rootSegment}`}>{i18next.t(listLabelKey)}</Link>},
    {title: lastLabel},
  ];
}

const BreadcrumbBar = ({uri}) => {
  const items = buildBreadcrumbItems(uri);
  if (!items) {
    return null;
  }
  return <Breadcrumb items={items} style={{marginLeft: 8}} />;
};

export default BreadcrumbBar;
