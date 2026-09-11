import {Monitor, Smartphone, Tablet} from "lucide-react";

export type DeviceType = "mobile" | "tablet" | "desktop";

export function parseUserAgent(ua: string): {label: string; device: DeviceType} {
  const browser =
    /Edg(A|iOS)?\//.test(ua) ? "Edge" :
      /OPR\/|Opera/.test(ua) ? "Opera" :
        /SamsungBrowser\//.test(ua) ? "Samsung Internet" :
          /MicroMessenger\//.test(ua) ? "WeChat" :
            /Firefox\/|FxiOS\//.test(ua) ? "Firefox" :
              /Chrome\/|CriOS\//.test(ua) ? "Chrome" :
                /Safari\//.test(ua) ? "Safari" : "";
  const os =
    /iPhone/.test(ua) ? "iPhone" :
      /iPad/.test(ua) ? "iPad" :
        /Android/.test(ua) ? "Android" :
          /Windows/.test(ua) ? "Windows" :
            /CrOS/.test(ua) ? "ChromeOS" :
              /Mac OS X|Macintosh/.test(ua) ? "macOS" :
                /Linux/.test(ua) ? "Linux" : "";
  const device: DeviceType =
    /iPad|Tablet/.test(ua) || (/Android/.test(ua) && !/Mobile/.test(ua)) ? "tablet" :
      /Mobile|iPhone|Android/.test(ua) ? "mobile" : "desktop";
  return {label: [browser, os].filter(Boolean).join(" · ") || ua, device};
}

export function DeviceIcon({device}: {device: DeviceType}) {
  const Icon = device === "mobile" ? Smartphone : device === "tablet" ? Tablet : Monitor;
  return <Icon className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />;
}
