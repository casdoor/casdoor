import i18next from "i18next";
import {Link} from "react-router-dom";
import {Button} from "@/components/ui/button";

export default function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-24 text-center">
      <div className="text-6xl font-semibold tracking-tight text-muted-foreground">404</div>
      <p className="text-muted-foreground">{i18next.t("general:Sorry, the page you visited does not exist.")}</p>
      <Button asChild>
        <Link to="/">{i18next.t("general:Back Home")}</Link>
      </Button>
    </div>
  );
}
