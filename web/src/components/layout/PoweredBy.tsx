import * as Conf from "@/Conf";
import {useIsDark} from "@/hooks/use-theme";
import * as Setting from "@/lib/setting";

/**
 * The default footer: "Powered by" followed by the Casdoor wordmark, which has a
 * light and a dark variant. The antd frontend renders the same image rather than
 * the word "Casdoor", and a deployment can replace the whole line through
 * `Conf.CustomFooter` or an organization's `footerHtml`. `dark` is for a footer
 * on a dark backdrop while the page itself is light.
 */
export function PoweredBy({dark = false}: {dark?: boolean}) {
  const isDark = useIsDark() || dark;

  if (Conf.CustomFooter !== null) {
    return <>{Conf.CustomFooter}</>;
  }

  return (
    <span className="inline-flex items-center gap-1">
      Powered by
      <a href="https://casdoor.org" target="_blank" rel="noreferrer">
        <img
          src={Setting.getLogo([isDark ? "dark" : "light"])}
          alt="Casdoor"
          height={20}
          className="h-5 w-auto pb-[3px]"
        />
      </a>
    </span>
  );
}
