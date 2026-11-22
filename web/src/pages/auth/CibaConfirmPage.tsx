// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import * as React from "react";
import i18next from "i18next";
import {Navigate, useLocation, useParams} from "react-router-dom";
import {CheckCircle2, XCircle} from "lucide-react";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {Loading} from "@/components/common/Loading";
import {useAccount} from "@/hooks/use-account";
import * as AuthBackend from "@/backend/AuthBackend";

/**
 * Opened from the email of a backchannel authentication request (CIBA): an application, e.g. a call center or
 * a counter, asked to sign the user in, and the user approves or denies it here after signing in.
 */
export default function CibaConfirmPage() {
  const {owner = "", id = ""} = useParams();
  const {account, loading} = useAccount();
  const location = useLocation();
  const [request, setRequest] = React.useState<any>(undefined);
  const [error, setError] = React.useState("");
  const [answering, setAnswering] = React.useState(false);

  React.useEffect(() => {
    if (!account) {
      return;
    }
    AuthBackend.getCibaRequest(id)
      .then((res: any) => {
        if (res.status === "ok") {
          setRequest(res.data);
        } else {
          setError(res.msg);
          setRequest(null);
        }
      })
      .catch((err: any) => {
        setError(err.message);
        setRequest(null);
      });
  }, [account, id]);

  if (loading || account === undefined) {
    return <Loading className="min-h-screen" />;
  }
  if (account === null) {
    return <Navigate to={`/login/${owner}`} replace state={{from: location.pathname}} />;
  }

  const answer = (approve: boolean) => {
    setAnswering(true);
    AuthBackend.answerCibaRequest(id, approve)
      .then((res: any) => {
        if (res.status === "ok") {
          setRequest({...request, status: res.data});
        } else {
          setError(res.msg);
        }
      })
      .catch((err: any) => setError(err.message))
      .finally(() => setAnswering(false));
  };

  const renderBody = () => {
    if (request === undefined) {
      return <p className="text-center text-sm text-muted-foreground">{i18next.t("general:Loading")}</p>;
    }
    if (request === null) {
      return (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      );
    }
    if (request.status === "Approved" || request.status === "Denied") {
      const approved = request.status === "Approved";
      return (
        <div className="flex flex-col items-center gap-3 py-4 text-center">
          {approved ? <CheckCircle2 className="h-10 w-10 text-success" /> : <XCircle className="h-10 w-10 text-destructive" />}
          <p className="text-sm">
            {approved ? i18next.t("application:The sign-in request is approved") : i18next.t("application:The sign-in request is denied")}
          </p>
        </div>
      );
    }

    const rows: [string, string][] = [
      [i18next.t("general:User"), request.userId],
      [i18next.t("general:Application"), request.appDisplayName || request.application],
      [i18next.t("payment:Message"), request.bindingMessage],
      [i18next.t("provider:Scope"), request.scope],
      [i18next.t("general:Created time"), new Date(request.createdTime).toLocaleString()],
    ];
    return (
      <div className="space-y-4">
        {request.appLogo ? <img src={request.appLogo} alt="" className="mx-auto h-10 object-contain" /> : null}
        <dl className="space-y-2 rounded-md border p-3 text-sm">
          {rows.filter(([, value]) => value && value.trim()).map(([label, value]) => (
            <div key={label} className="flex gap-3">
              <dt className="w-24 shrink-0 text-muted-foreground">{label}</dt>
              <dd className="break-all">{value}</dd>
            </div>
          ))}
        </dl>
        {request.status === "Expired" ? (
          <Alert variant="destructive">
            <AlertDescription>{i18next.t("mfa:The sign-in request has expired")}</AlertDescription>
          </Alert>
        ) : (
          <div className="grid grid-cols-2 gap-3">
            <Button variant="outline" disabled={answering} onClick={() => answer(false)}>
              {i18next.t("permission:Deny")}
            </Button>
            <Button loading={answering} onClick={() => answer(true)}>
              {i18next.t("mfa:Approve")}
            </Button>
          </div>
        )}
        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}
      </div>
    );
  };

  return (
    <AuthLayout>
      <div className="space-y-4">
        <h2 className="text-center text-xl font-semibold tracking-tight">{i18next.t("application:Confirm the sign-in request")}</h2>
        <p className="text-center text-sm text-muted-foreground">{i18next.t("application:Approve it only if you asked for it, e.g. at a counter or on a call, and the message matches what you see there")}</p>
        {renderBody()}
      </div>
    </AuthLayout>
  );
}
