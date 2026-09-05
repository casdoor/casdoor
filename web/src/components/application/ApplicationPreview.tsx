import * as React from "react";
import * as Setting from "@/lib/setting";

const LoginPage = React.lazy(() => import("@/pages/auth/LoginPage"));
const SignupPage = React.lazy(() => import("@/pages/auth/SignupPage"));
const PromptPage = React.lazy(() => import("@/pages/auth/PromptPage"));

/**
 * One live page rendered from the application currently being edited, behind a
 * mask that swallows every click — the port of `renderSignupSigninPreview()` and
 * `renderPromptPreview()` in `web/src/ApplicationEditPage.js`. The pages read
 * their own theme and custom CSS off the application, so what shows here is what
 * the visitor gets.
 */
function PreviewFrame({children}: {children: React.ReactNode}) {
  return (
    <div className="relative h-[560px] overflow-auto rounded-lg border shadow-md">
      <React.Suspense fallback={null}>{children}</React.Suspense>
      {/* the preview is to look at, not to use */}
      <div className="absolute inset-0 z-10 cursor-not-allowed bg-foreground/5" />
    </div>
  );
}

/** The sign-up and sign-in pages side by side, as the antd editor shows them. */
export function ApplicationSignupSigninPreview({application}: {application: any}) {
  return (
    <div className="grid gap-4 xl:grid-cols-2">
      <PreviewFrame>
        {/* without a password form there is no sign-up form either, so antd previews
            the sign-in page in its place */}
        {Setting.isPasswordEnabled(application) ? (
          <SignupPage application={application} />
        ) : (
          <LoginPage type="login" application={application} preview="auto" />
        )}
      </PreviewFrame>
      <PreviewFrame>
        <LoginPage type="login" application={application} preview="auto" />
      </PreviewFrame>
    </div>
  );
}

export function ApplicationPromptPreview({application}: {application: any}) {
  return (
    <div className="xl:max-w-[50%]">
      <PreviewFrame>
        <PromptPage application={application} />
      </PreviewFrame>
    </div>
  );
}
