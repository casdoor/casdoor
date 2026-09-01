import * as Setting from "@/lib/setting";

/**
 * Picks the captcha provider that matches the application's active captcha rule.
 * Ported from LoginPage.renderCaptchaModal / SignupPage.renderCaptchaModal: the
 * provider is chosen by rule, not by a fixed priority.
 */
export function getCaptchaProvider(application: any): any | null {
  const rule = Setting.getCaptchaRule(application);
  if (rule === Setting.CaptchaRule.Never) {
    return null;
  }
  const items = Setting.getCaptchaProviderItems(application) ?? [];
  const matched = items.filter((item: any) => item.rule === rule);
  return matched.length > 0 ? matched[0].provider : null;
}
