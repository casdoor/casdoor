import * as React from "react";
import {ThemePreviewMessage} from "@/lib/integration";

const LoginPage = React.lazy(() => import("@/pages/auth/LoginPage"));

/**
 * The sign-in page of an application with a Marketplace theme applied, shown in a frame of the
 * install dialog. The application comes from the console that embeds this page, never from the URL.
 */
export default function ThemePreviewPage() {
  const [application, setApplication] = React.useState<any>(null);

  React.useEffect(() => {
    const onMessage = (event: MessageEvent) => {
      if (event.origin !== window.location.origin || event.source !== window.parent) {
        return;
      }
      if (event.data?.type === ThemePreviewMessage) {
        setApplication(event.data.application);
      }
    };
    window.addEventListener("message", onMessage);
    window.parent.postMessage({type: `${ThemePreviewMessage}-ready`}, window.location.origin);
    return () => window.removeEventListener("message", onMessage);
  }, []);

  if (!application) {
    return null;
  }
  return (
    <React.Suspense fallback={null}>
      <LoginPage type="login" application={application} preview="theme" />
    </React.Suspense>
  );
}
