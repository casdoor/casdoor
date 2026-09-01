import React from "react";
import * as Setting from "./Setting";

export const TourObj = {
  home: [
    {
      title: "Welcome to casdoor",
      description: "You can learn more about the use of CasDoor at https://casdoor.org/.",
      cover: (
        <img
          alt="casdoor.png"
          src={`${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256.png`}
        />
      ),
    },
    {
      title: "Statistic cards",
      description: "Here are four statistic cards for user information.",
      id: "statistic",
    },
    {
      title: "Import users",
      description: "You can add new users or update existing Casdoor users by uploading a XLSX file of user information.",
      id: "echarts-chart",
    },
  ],
  webhooks: [
    {
      title: "Webhook List",
      description: "Event systems allow you to build integrations, which subscribe to certain events on Casdoor. When one of those event is triggered, we'll send a POST json payload to the configured URL. The application parsed the json payload and carry out the hooked function. Events consist of signup, login, logout, update users, which are stored in the action field of the record. Event systems can be used to update an external issue from users.",
    },
  ],
  syncers: [
    {
      title: "Syncer List",
      description: "Casdoor stores users in user table. Don't worry about migrating your application user data into Casdoor, when you plan to use Casdoor as an authentication platform. Casdoor provides syncer to quickly help you sync user data to Casdoor.",
    },
  ],
  sysinfo: [
    {
      title: "CPU Usage",
      description: "You can see the CPU usage in real time.",
      id: "cpu-card",
    },
    {
      title: "Memory Usage",
      description: "You can see the Memory usage in real time.",
      id: "memory-card",
    },
    {
      title: "API Latency",
      description: "You can see the usage statistics of each API latency in real time.",
      id: "latency-card",
    },
    {
      title: "API Throughput",
      description: "You can see the usage statistics of each API throughput in real time.",
      id: "throughput-card",
    },
    {
      title: "About Casdoor",
      description: "You can get more Casdoor information in this card.",
      id: "about-card",
    },
  ],
  subscriptions: [
    {
      title: "Subscription List",
      description: "Subscription helps to manage user's selected plan that make easy to control application's features access.",
    },
  ],
  pricings: [
    {
      title: "Price List",
      description: "Casdoor can be used as subscription management system via plan, pricing and subscription.",
    },
  ],
  plans: [
    {
      title: "Plan List",
      description: "Plan  describe list of application's features with own name and price. Plan features depends on Casdoor role with set of permissions.That allow to describe plan's features independ on naming and price. For example: plan may has diffrent prices depends on county or date.",
    },
  ],
  payments: [
    {
      title: "Payment List",
      description: "After the payment is successful, you can see the transaction information of the products in Payment, such as organization, user, purchase time, product name, etc.",
    },
  ],
  products: [
    {
      title: "Session List",
      description: "You can add the product (or service) you want to sell. The following will tell you how to add a product.",
    },
  ],
  sessions: [
    {
      title: "Session List",
      description: "You can get Session ID in this list.",
    },
  ],
  tokens: [
    {
      title: "Token List",
      description: "Casdoor is based on OAuth. Tokens are users' OAuth token.You can get access token in this list.",
    },
  ],
  enforcers: [
    {
      title: "Enforcer List",
      description: "In addition to the API interface for requesting enforcement of permission control, Casdoor also provides other interfaces that help external applications obtain permission policy information, which is also listed here.",
    },
  ],
  adapters: [
    {
      title: "Adapter List",
      description: "Casdoor supports using the UI to connect the adapter and manage the policy rules. In Casbin, the policy storage is implemented as an adapter (aka middleware for Casbin). A Casbin user can use an adapter to load policy rules from a storage, or save policy rules to it.",
    },
  ],
  models: [
    {
      title: "Model List",
      description: "Model defines your permission policy structure, and how requests should match these permission policies and their effects. Then you can user model in Permission.",
    },
  ],
  permissions: [
    {
      title: "Permission List",
      description: "All users associated with a single Casdoor organization are shared between the organization's applications and therefore have access to the applications. Sometimes you may want to restrict users' access to certain applications, or certain resources in a certain application. In this case, you can use Permission implemented by Casbin.",
    },
    {
      title: "Permission Add",
      description: "In the Casdoor Web UI, you can add a Model for your organization in the Model configuration item, and a Policy for your organization in the Permission configuration item. ",
      id: "add-button",
    },
    {
      title: "Permission Upload",
      description: "With Casbin Online Editor, you can get Model and Policy files suitable for your usage scenarios. You can easily import the Model file into Casdoor through the Casdoor Web UI for use by the built-in Casbin. ",
      id: "upload-button",
    },
  ],
  roles: [
    {
      title: "Role List",
      description: "Each user may have multiple roles. You can see the user's roles on the user's profile.",
    },
  ],
  resources: [
    {
      title: "Resource List",
      description: "You can upload resources in casdoor. Before upload resources, you need to configure a storage provider. Please see Storage Provider.",
    },
    {
      title: "Upload Resource",
      description: "Users can upload resources such as files and images to the previously configured cloud storage.",
      id: "upload-button",
    },
  ],
  providers: [
    {
      title: "Provider List",
      description: "We have 6 kinds of providers:OAuth providers、SMS Providers、Email Providers、Storage Providers、Payment Provider、Captcha Provider.",
    },
    {
      title: "Provider Add",
      description: "You must add the provider to application, then you can use the provider in your application",
      id: "add-button",
    },
  ],
  organizations: [
    {
      title: "Organization List",
      description: "Organization is the basic unit of Casdoor, which manages users and applications. If a user signed in to an organization, then he can access all applications belonging to the organization without signing in again.",
    },
  ],
  groups: [
    {
      title: "Group List",
      description: "In the groups list pages, you can see all the groups in organizations.",
    },
  ],
  users: [
    {
      title: "User List",
      description: "As an authentication platform, Casdoor is able to manage users.",
    },
    {
      title: "Import users",
      description: "You can add new users or update existing Casdoor users by uploading a XLSX file of user information.",
      id: "upload-button",
    },
  ],
  applications: [
    {
      title: "Application List",
      description: "If you want to use Casdoor to provide login service for your web Web APPs, you can add them as Casdoor applications. Users can access all applications in their organizations without login twice.",
    },
  ],
};

export const TourUrlList = ["home", "organizations", "groups", "users", "applications", "providers", "resources", "roles", "permissions", "models", "adapters", "enforcers", "tokens", "sessions", "products", "payments", "plans", "pricings", "subscriptions", "sysinfo", "syncers", "webhooks"];

export function getNextUrl(pathName = window.location.pathname) {
  return TourUrlList[TourUrlList.indexOf(pathName.replace("/", "")) + 1] || "";
}

let orgIsTourVisible = true;

export function setOrgIsTourVisible(visible) {
  orgIsTourVisible = visible;
  if (orgIsTourVisible === false) {
    setIsTourVisible(false);
  }
}

export function setIsTourVisible(visible) {
  localStorage.setItem("isTourVisible", visible);
  window.dispatchEvent(new Event("storageTourChanged"));
}

export function setTourLogo(tourLogoSrc) {
  if (tourLogoSrc !== "") {
    TourObj["home"][0]["cover"] = (<img alt="casdoor.png" src={tourLogoSrc} />);
  }
}

export function getTourVisible() {
  return localStorage.getItem("isTourVisible") !== "false";
}

export function getNextButtonChild(nextPathName) {
  return nextPathName !== "" ?
    `Go to "${nextPathName.charAt(0).toUpperCase()}${nextPathName.slice(1)} List"`
    : "Finish";
}

export function getSteps() {
  const path = window.location.pathname.replace("/", "");
  const res = TourObj[path];
  if (res === undefined) {
    return [];
  } else {
    return res;
  }
}
