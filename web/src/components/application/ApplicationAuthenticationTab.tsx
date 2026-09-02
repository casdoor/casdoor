
import i18next from "i18next";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {FormRow} from "@/components/crud/FormRow";
import type {ApplicationTabProps} from "@/components/application/types";
import * as Setting from "@/lib/setting";

/** The "Authentication" tab: what the sign-in session allows and where it sends the user. */
export function ApplicationAuthenticationTab({application, updateField}: ApplicationTabProps) {
  return (
    <>
      <FormRow labelKey="application:Cookie expire">
        <Input
          type="number"
          value={application.cookieExpireInHours ?? 720}
          onChange={(e) => updateField("cookieExpireInHours", Setting.myParseInt(e.target.value))}
        />
      </FormRow>
      <FormRow labelKey="ldap:Default group">
        <Input value={application.defaultGroup ?? ""} onChange={(e) => updateField("defaultGroup", e.target.value)} />
      </FormRow>
      <FormRow labelKey="application:Enable signup">
        <Switch checked={!!application.enableSignUp} onCheckedChange={(v) => updateField("enableSignUp", v)} />
      </FormRow>
      <FormRow labelKey="application:Disable signin">
        <Switch checked={!!application.disableSignin} onCheckedChange={(v) => updateField("disableSignin", v)} />
      </FormRow>
      <FormRow labelKey="application:Enable guest signin">
        <Switch
          checked={!!application.enableGuestSignin}
          onCheckedChange={(v) => updateField("enableGuestSignin", v)}
        />
      </FormRow>
      <FormRow labelKey="application:Enable exclusive signin">
        <Switch
          checked={!!application.enableExclusiveSignin}
          onCheckedChange={(v) => updateField("enableExclusiveSignin", v)}
        />
      </FormRow>
      <FormRow labelKey="application:Signin session">
        <Switch
          checked={!!application.enableSigninSession}
          onCheckedChange={(v) => updateField("enableSigninSession", v)}
        />
      </FormRow>
      <FormRow labelKey="application:Auto signin">
        <Switch
          checked={!!application.enableAutoSignin}
          onCheckedChange={(v) => {
            // auto signin reuses the Casdoor session, so it needs one to exist
            if (v && !application.enableSigninSession) {
              Setting.showMessage(
                "error",
                i18next.t("application:Please enable \"Signin session\" first before enabling \"Auto signin\""),
              );
              return;
            }
            updateField("enableAutoSignin", v);
          }}
        />
      </FormRow>
      <FormRow labelKey="application:Enable Email linking">
        <Switch
          checked={!!application.enableLinkWithEmail}
          onCheckedChange={(v) => updateField("enableLinkWithEmail", v)}
        />
      </FormRow>
      <FormRow labelKey="general:Signup URL">
        <Input value={application.signupUrl ?? ""} onChange={(e) => updateField("signupUrl", e.target.value)} />
      </FormRow>
      <FormRow labelKey="general:Signin URL">
        <Input value={application.signinUrl ?? ""} onChange={(e) => updateField("signinUrl", e.target.value)} />
      </FormRow>
      <FormRow labelKey="general:Forget URL">
        <Input value={application.forgetUrl ?? ""} onChange={(e) => updateField("forgetUrl", e.target.value)} />
      </FormRow>
      <FormRow labelKey="general:Affiliation URL">
        <Input
          value={application.affiliationUrl ?? ""}
          onChange={(e) => updateField("affiliationUrl", e.target.value)}
        />
      </FormRow>
    </>
  );
}
