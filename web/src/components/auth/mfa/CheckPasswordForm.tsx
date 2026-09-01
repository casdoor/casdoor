import * as React from "react";
import i18next from "i18next";
import {Lock} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import * as UserBackend from "@/backend/UserBackend";

/** Step 1 of the MFA wizard: confirm the current password. */
export function CheckPasswordForm({
  user,
  onSuccess,
  onFail,
}: {
  user: any;
  onSuccess: (res: any) => void;
  onFail: (res: any) => void;
}) {
  const [password, setPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    UserBackend.checkUserPassword({...user, password})
      .then((res: any) => {
        if (res.status === "ok") {
          onSuccess(res);
        } else {
          onFail(res);
        }
      })
      .finally(() => {
        setPassword("");
        setLoading(false);
      });
  };

  return (
    <form className="mx-auto w-[300px] space-y-4" onSubmit={submit}>
      <div className="relative">
        <Lock className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input
          autoFocus
          type="password"
          className="pl-8"
          autoComplete="current-password"
          value={password}
          placeholder={i18next.t("general:Password")}
          onChange={(e) => setPassword(e.target.value)}
        />
      </div>
      <Button type="submit" className="w-full" loading={loading}>
        {i18next.t("forget:Next Step")}
      </Button>
    </form>
  );
}
