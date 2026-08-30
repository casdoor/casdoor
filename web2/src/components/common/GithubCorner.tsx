import * as Conf from "@/Conf";

/**
 * The "Fork me on GitHub" ribbon, shown when the backend sets showGithubCorner.
 * Ported from web/src/common/CustomGithubCorner.js, which wraps
 * `react-github-corner`; the same SVG is inlined here instead.
 */
export function GithubCorner({href = "https://github.com/casdoor/casdoor", size = 60}: {href?: string; size?: number}) {
  if (!Conf.ShowGithubCorner) {
    return null;
  }

  return (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      aria-label="View source on GitHub"
      className="github-corner fixed right-0 top-0 z-40"
    >
      <svg
        width={size}
        height={size}
        viewBox="0 0 250 250"
        aria-hidden="true"
        className="fill-foreground text-background"
        style={{border: 0, position: "absolute", right: 0, top: 0}}
      >
        <path d="M0,0 L115,115 L130,115 L142,142 L250,250 L250,0 Z" />
        <path
          d="M128.3,109.0 C113.8,99.7 119.0,89.6 119.0,89.6 C122.0,82.7 120.5,78.6 120.5,78.6 C119.2,72.0 123.4,76.3 123.4,76.3 C127.3,80.9 125.5,87.3 125.5,87.3 C122.9,97.6 130.6,101.9 134.4,103.2"
          fill="currentColor"
          className="octo-arm origin-[130px_106px]"
        />
        <path
          d="M115.0,115.0 C114.9,115.1 118.7,116.5 119.8,115.4 L133.7,101.6 C136.9,99.2 139.9,98.4 142.2,98.6 C133.8,88.0 127.5,74.4 143.8,58.0 C148.5,53.4 154.0,51.2 159.7,51.0 C160.3,49.4 163.2,43.6 171.4,40.1 C171.4,40.1 176.1,42.5 178.8,56.2 C183.1,58.4 187.2,61.2 190.9,64.9 C194.5,68.5 197.3,72.6 199.5,77.0 C213.2,79.6 215.7,84.4 215.7,84.4 C212.2,92.6 206.4,95.5 204.8,96.1 C204.6,101.8 202.4,107.3 197.8,112.0 C181.4,128.3 167.8,122.0 157.2,113.6 C157.4,116.0 156.6,118.9 154.2,122.1 L140.4,136.0 C139.3,137.1 140.7,140.9 140.8,140.8 Z"
          fill="currentColor"
          className="octo-body"
        />
      </svg>
    </a>
  );
}
