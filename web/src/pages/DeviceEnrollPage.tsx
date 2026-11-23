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
import {useSearchParams} from "react-router-dom";
import {Laptop, ShieldCheck, Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {Input} from "@/components/ui/input";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {DevicePostureCell, DeviceStatusBadge} from "@/pages/DeviceListPage";
import * as DeviceBackend from "@/backend/DeviceBackend";
import * as Setting from "@/lib/setting";

/** The user approves that the device showing the enrollment code is theirs, and sees their devices. */
export default function DeviceEnrollPage() {
  const [searchParams] = useSearchParams();
  const [code, setCode] = React.useState(searchParams.get("code") ?? "");
  const [device, setDevice] = React.useState<any>(null);
  const [approved, setApproved] = React.useState<any>(null);
  const [loading, setLoading] = React.useState(false);
  const [devices, setDevices] = React.useState<any[]>([]);
  const [version, setVersion] = React.useState(0);

  React.useEffect(() => {
    DeviceBackend.getMyDevices().then((res: any) => {
      if (res.status === "ok") {
        setDevices(res.data ?? []);
      }
    });
  }, [version]);

  const lookup = React.useCallback((value: string) => {
    if (value.replace("-", "").length < 8) {
      return;
    }
    setLoading(true);
    DeviceBackend.getDeviceEnrollment(value)
      .then((res: any) => {
        if (res.status === "ok") {
          setDevice(res.data);
        } else {
          setDevice(null);
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setLoading(false));
  }, []);

  React.useEffect(() => {
    if (searchParams.get("code")) {
      lookup(searchParams.get("code") ?? "");
    }
  }, [searchParams, lookup]);

  const approve = () => {
    setLoading(true);
    DeviceBackend.approveDeviceEnrollment(code)
      .then((res: any) => {
        if (res.status === "ok") {
          setApproved(res.data);
          setDevice(null);
          setVersion((v) => v + 1);
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setLoading(false));
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center gap-2 space-y-0">
          <Laptop className="size-5 text-primary" />
          <CardTitle className="text-base">{i18next.t("device:Enroll a device")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {approved ? (
            <div className="flex items-start gap-3 rounded-md border border-success/40 bg-success/10 p-3 text-sm">
              <ShieldCheck className="mt-0.5 size-4 text-success" />
              <div>
                <div className="font-medium">{approved.displayName}</div>
                <div>
                  {approved.status === "Trusted"
                    ? i18next.t("device:The device is enrolled and trusted, you can go back to it")
                    : i18next.t("device:The device is enrolled, an administrator approves it before it is trusted")}
                </div>
              </div>
            </div>
          ) : null}
          <p className="text-sm text-muted-foreground">{i18next.t("device:Type the code that the Casdoor device agent shows on your computer")}</p>
          <div className="flex gap-2">
            <Input
              className="font-mono uppercase tracking-widest"
              placeholder="ABCD-EFGH"
              value={code}
              onChange={(e) => {
                setCode(e.target.value);
                setDevice(null);
              }}
            />
            <Button type="button" variant="outline" loading={loading && !device} onClick={() => lookup(code)}>
              {i18next.t("edition:Next")}
            </Button>
          </div>
          {device ? (
            <div className="space-y-3 rounded-md border p-3">
              <div className="grid gap-1 text-sm sm:grid-cols-[120px_1fr]">
                <span className="text-muted-foreground">{i18next.t("entry:Device")}</span>
                <span className="font-medium">{device.displayName}</span>
                <span className="text-muted-foreground">{i18next.t("device:Platform")}</span>
                <span>{device.platform} {device.posture?.osVersion}</span>
                <span className="text-muted-foreground">{i18next.t("general:Client IP")}</span>
                <span>{device.lastIp}</span>
                <span className="text-muted-foreground">{i18next.t("device:Posture")}</span>
                <DevicePostureCell device={device} />
              </div>
              <p className="text-xs text-muted-foreground">{i18next.t("device:Only approve a device that you are using right now, it will sign in as you")}</p>
              <Button type="button" onClick={approve} loading={loading}>
                {i18next.t("device:This is my device, enroll it")}
              </Button>
            </div>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{i18next.t("device:My devices")}</CardTitle>
        </CardHeader>
        <CardContent>
          {devices.length === 0 ? (
            <p className="text-sm text-muted-foreground">{i18next.t("device:No devices yet")}</p>
          ) : (
            <div className="divide-y rounded-md border">
              {devices.map((item) => (
                <div key={item.name} className="flex flex-wrap items-center gap-3 px-3 py-2 text-sm">
                  <div className="min-w-0 flex-1">
                    <div className="font-medium">{item.displayName}</div>
                    <div className="text-xs text-muted-foreground">
                      {item.platform} {item.posture?.osVersion}, {i18next.t("device:Last seen")} {Setting.getFormattedDate(item.lastSeenTime)}
                    </div>
                    <DevicePostureCell device={item} />
                  </div>
                  <DeviceStatusBadge device={item} />
                  <ConfirmButton
                    variant="ghost"
                    size="iconSm"
                    title={i18next.t("device:Remove this device?")}
                    description={i18next.t("device:It can't sign in as you until it is enrolled again")}
                    onConfirm={() =>
                      DeviceBackend.deleteDevice(item).then((res: any) => {
                        if (res.status === "ok") {
                          setVersion((v) => v + 1);
                        } else {
                          Setting.showMessage("error", res.msg);
                        }
                      })
                    }
                  >
                    <Trash2 />
                  </ConfirmButton>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
