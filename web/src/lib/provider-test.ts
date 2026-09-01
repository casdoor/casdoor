import i18next from "i18next";
import * as Setting from "@/lib/setting";

/**
 * The "test this provider" calls of the provider edit page, ported from
 * web/src/common/TestEmailWidget.js, TestSmsWidget.js and
 * TestNotificationWidget.js. The request bodies are unchanged so the backend
 * keeps seeing exactly what the antd frontend sent.
 */

function testEmailProvider(provider: any, email = "") {
  const emailForm = {
    title: provider.title,
    content: provider.content,
    sender: provider.displayName,
    receivers: email === "" ? ["TestSmtpServer"] : [email],
    provider: provider.name,
    providerObject: provider,
    owner: provider.owner,
    name: provider.name,
  };

  return fetch(`${Setting.ServerUrl}/api/send-email`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(emailForm),
  }).then(res => res.json());
}

export function sendTestEmail(provider: any, email: string) {
  return testEmailProvider(provider, email)
    .then((res) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully sent"));
      } else {
        Setting.showMessage("error", res.msg);
      }
    })
    .catch(error => {
      Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    });
}

export function connectSmtpServer(provider: any) {
  return testEmailProvider(provider)
    .then((res) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("provider:SMTP connected successfully"));
      } else {
        Setting.showMessage("error", res.msg);
      }
    })
    .catch(error => {
      Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    });
}

function testSmsProvider(provider: any, phone = "") {
  const smsForm = {
    content: "123456",
    receivers: [phone],
    owner: provider.owner,
    name: provider.name,
  };

  return fetch(`${Setting.ServerUrl}/api/send-sms?provider=` + provider.name, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(smsForm),
  }).then(res => res.json());
}

export function sendTestSms(provider: any, phone: string) {
  return testSmsProvider(provider, phone)
    .then((res) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully sent"));
      } else {
        Setting.showMessage("error", res.msg);
      }
    })
    .catch(error => {
      Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    });
}

function testNotificationProvider(provider: any) {
  const notificationForm = {
    content: provider.content,
    owner: provider.owner,
    name: provider.name,
  };

  return fetch(`${Setting.ServerUrl}/api/send-notification?provider=${provider.name}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(notificationForm),
  }).then(res => res.json());
}

export function sendTestNotification(provider: any) {
  return testNotificationProvider(provider)
    .then((res) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully sent"));
      } else {
        Setting.showMessage("error", res.msg);
      }
    })
    .catch(error => {
      Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    });
}
