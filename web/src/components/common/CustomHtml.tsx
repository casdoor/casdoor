import * as React from "react";
import {getSafeHtml} from "@/lib/custom-html";
import * as Setting from "@/lib/setting";

interface CustomHtmlProps {
  html?: string | null;
  trusted?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

/**
 * Renders application-supplied HTML. <script> tags inserted through innerHTML are
 * never executed by the browser, so for trusted applications they are re-created
 * here, as the antd Setting.RenderCustomHtml() did.
 */
export function CustomHtml({html, trusted = false, className, style}: CustomHtmlProps) {
  const containerRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const container = containerRef.current;
    if (!container || !trusted) {
      return;
    }
    container.querySelectorAll("script").forEach((oldScript) => {
      const script = document.createElement("script");
      Array.from(oldScript.attributes).forEach((attr) => script.setAttribute(attr.name, attr.value));
      script.textContent = oldScript.textContent;
      oldScript.parentNode?.replaceChild(script, oldScript);
    });
  }, [html, trusted]);

  if (!html) {
    return null;
  }

  return <div ref={containerRef} className={className} style={style} dangerouslySetInnerHTML={{__html: getSafeHtml(html, trusted)}} />;
}

/** A `customCss` / `formCss` value, injected as a stylesheet. */
export function CustomStyle({css}: {css?: string | null}) {
  const inner = Setting.getStyleInnerCss(css);
  if (!inner || inner.trim() === "") {
    return null;
  }
  return <style dangerouslySetInnerHTML={{__html: inner}} />;
}
