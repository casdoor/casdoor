import * as React from "react";
import * as Setting from "@/lib/setting";

/**
 * In development the frontend and the backend run on different ports, so
 * /.well-known/openid-configuration has to be forwarded to the backend.
 * Ported from web/src/auth/OidcDiscoveryPage.js.
 */
export default function OidcDiscoveryPage() {
  React.useEffect(() => {
    if (Setting.isLocalhost()) {
      Setting.goToLink(`${Setting.ServerUrl}/.well-known/openid-configuration`);
    }
  }, []);

  return null;
}
