import i18next from "i18next";
import {Link} from "react-router-dom";
import {Button} from "@/components/ui/button";

/**
 * What a console page shows when the backend answers a read with
 * "Unauthorized operation" — the antd frontend's `<Result status="403">`. Every
 * list page and the application/user edit pages fall back to this rather than
 * rendering an empty table for an account that may not look at the object.
 */
export function UnauthorizedPage() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-24 text-center">
      <div className="text-6xl font-semibold tracking-tight text-muted-foreground">403</div>
      <h1 className="text-xl font-semibold">{i18next.t("general:Unauthorized")}</h1>
      <p className="max-w-xl text-muted-foreground">
        {i18next.t("general:Sorry, you do not have permission to access this page or logged in status invalid.")}
      </p>
      <Button asChild>
        <Link to="/">{i18next.t("general:Back Home")}</Link>
      </Button>
    </div>
  );
}
