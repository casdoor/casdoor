import DOMPurify from "dompurify";

/**
 * Only the built-in organization's applications are edited by global admins; any
 * other organization's admin must not get script running on the Casdoor origin.
 */
export function isTrustedApplication(application?: any): boolean {
  return application?.organization === "built-in";
}

export function getSafeHtml(html: string, trusted: boolean): string {
  return trusted ? html : DOMPurify.sanitize(html, {FORCE_BODY: true});
}
