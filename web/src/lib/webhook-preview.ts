import * as Setting from "@/lib/setting";

/**
 * The sample payload the webhook preview renders, lifted from
 * `web/src/WebhookEditPage.js` so the two frontends show the same thing.
 */
const applicationTemplate = {
  owner: "admin", // this.props.account.applicationName,
  name: "application_123",
  organization: "built-in",
  createdTime: "2022-01-01T01:03:42+08:00",
  displayName: "New Application - 123",
  logo: `${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256.png`,
  enablePassword: true,
  enableSignUp: true,
  disableSignin: false,
  enableSigninSession: false,
  enableCodeSignin: false,
  enableSamlCompress: false,
};

const previewTemplate = {
  "id": 9078,
  "owner": "built-in",
  "name": "68f55b28-7380-46b1-9bde-64fe1576e3b3",
  "createdTime": "2022-01-01T01:03:42+08:00",
  "organization": "built-in",
  "clientIp": "159.89.126.192",
  "user": "admin",
  "method": "POST",
  "requestUri": "/api/add-application",
  "action": "login",
  "isTriggered": false,
  "object": JSON.stringify(applicationTemplate),
};

const userTemplate = {
  "owner": "built-in",
  "name": "admin",
  "createdTime": "2020-07-16T21:46:52+08:00",
  "updatedTime": "",
  "deletedTime": "",
  "id": "9eb20f79-3bb5-4e74-99ac-39e3b9a171e8",
  "type": "normal-user",
  "password": "***",
  "passwordSalt": "",
  "displayName": "Admin",
  "avatar": "https://cdn.casbin.com/usercontent/admin/avatar/1596241359.png",
  "permanentAvatar": "https://cdn.casbin.com/casdoor/avatar/casbin/admin.png",
  "email": "admin@example.com",
  "phone": "",
  "location": "",
  "address": null,
  "affiliation": "",
  "title": "",
  "score": 10000,
  "ranking": 10,
  "isOnline": false,
  "isAdmin": true,
  "isForbidden": false,
  "isDeleted": false,
  "signupApplication": "app-casnode",
  "properties": {
    "bio": "",
    "checkinDate": "20200801",
    "editorType": "",
    "emailVerifiedTime": "2020-07-16T21:46:52+08:00",
    "fileQuota": "50",
    "location": "",
    "no": "22",
    "oauth_QQ_displayName": "",
    "oauth_QQ_verifiedTime": "",
    "oauth_WeChat_displayName": "",
    "oauth_WeChat_verifiedTime": "",
    "onlineStatus": "false",
    "phoneVerifiedTime": "",
    "renameQuota": "3",
    "tagline": "",
    "website": "",
  },
};

/** The record a webhook would post, with the extended user the settings ask for. */
export function buildWebhookPreview(webhook: any): string {
  const preview: Record<string, any> = {...previewTemplate};
  if (webhook?.isUserExtended) {
    if (webhook.tokenFields && webhook.tokenFields.length !== 0) {
      const extendedUser: Record<string, any> = {};
      webhook.tokenFields.forEach((field: string) => {
        const key = field.replace(field[0], field[0].toLowerCase());
        extendedUser[key] = (userTemplate as Record<string, any>)[key];
      });
      preview["extendedUser"] = extendedUser;
    } else {
      preview["extendedUser"] = userTemplate;
    }
  }
  return JSON.stringify(preview, null, 2);
}
