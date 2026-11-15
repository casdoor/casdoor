import DOMPurify from "dompurify";
import * as Conf from "@/Conf";

/**
 * Only global admins edit the custom HTML of the built-in organization's applications
 * and of applications with custom scripts enabled; any other organization's admin
 * must not get script running on the Casdoor origin.
 */
export function isTrustedApplication(application?: any): boolean {
  return application?.organization === "built-in" || application?.enableCustomScript === true;
}

// DOMPurify drops `target`, which would open a footer's link over the page it is on;
// a link kept with it must not hand the opened page a reference back to this one
DOMPurify.addHook("afterSanitizeAttributes", (node) => {
  if (node.tagName === "A" && node.getAttribute("target")) {
    node.setAttribute("target", "_blank");
    node.setAttribute("rel", "noopener noreferrer");
  }
});

export function getSafeHtml(html: string, trusted: boolean): string {
  if (trusted) {
    return html;
  }
  // a global admin allows these in the global settings, and the server checks where they load from before an organization admin can save them
  const tags = Conf.CustomHtmlAllowedTags.filter((tag) => tag !== "script");
  return DOMPurify.sanitize(html, {FORCE_BODY: true, ADD_ATTR: ["target"], ADD_TAGS: tags});
}
