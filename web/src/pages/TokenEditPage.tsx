import i18next from "i18next";
import copy from "copy-to-clipboard";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Textarea} from "@/components/ui/textarea";
import {CodeEditor} from "@/components/common/CodeEditor";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import * as TokenBackend from "@/backend/TokenBackend";
import * as Setting from "@/lib/setting";

/** base64url -> JSON, the part of `jwt-decode` this page actually needs */
function decodeSegment(segment: string) {
  const base64 = segment.replace(/-/g, "+").replace(/_/g, "/");
  const json = decodeURIComponent(
    atob(base64)
      .split("")
      .map((c) => `%${`00${c.charCodeAt(0).toString(16)}`.slice(-2)}`)
      .join(""),
  );
  return JSON.parse(json);
}

/**
 * Decodes the access token for the read-only panel next to it. It never
 * verifies the signature — this is the same "what is in this JWT" view the
 * antd page renders with jwt-decode.
 */
function parseAccessToken(accessToken: string | undefined) {
  try {
    const parts = (accessToken ?? "").split(".");
    if (parts.length < 2) {
      throw new Error("Invalid token specified: missing part #2");
    }
    return JSON.stringify({header: decodeSegment(parts[0]), payload: decodeSegment(parts[1])}, null, 2);
  } catch (error: any) {
    return JSON.stringify({error: error?.message ?? String(error)}, null, 2);
  }
}

function parsedResultHeight(parsedResult: string) {
  const lines = parsedResult.split("\n").length;
  return Math.min(30, Math.max(10, lines)) * 22;
}

function TokenPanel({token, update}: {token: any; update: (field: string, value: any) => void}) {
  const accessToken: string = token.accessToken ?? "";
  const parsedResult = parseAccessToken(accessToken);

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{i18next.t("token:Access token")}</span>
          <Button
            size="sm"
            disabled={accessToken === ""}
            onClick={() => {
              copy(accessToken);
              Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
            }}
          >
            {i18next.t("token:Copy access token")}
          </Button>
        </div>
        <Textarea rows={12} value={accessToken} onChange={(e) => update("accessToken", e.target.value)} />
      </div>
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{i18next.t("token:Parsed result")}</span>
          <Button
            size="sm"
            disabled={!parsedResult.includes("\"alg\":")}
            onClick={() => {
              copy(parsedResult);
              Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
            }}
          >
            {i18next.t("token:Copy parsed result")}
          </Button>
        </div>
        <CodeEditor
          value={parsedResult}
          onChange={() => undefined}
          language="json"
          readOnly
          height={parsedResultHeight(parsedResult)}
        />
      </div>
    </div>
  );
}

export default function TokenEditPage() {
  const {tokenName = ""} = useParams();

  const fields: EditField[] = [
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "application", labelKey: "general:Application"},
    {type: "text", name: "organization", labelKey: "general:Organization"},
    {type: "text", name: "user", labelKey: "general:User"},
    {type: "textarea", name: "code", labelKey: "token:Authorization code", rows: 3},
    {type: "number", name: "expiresIn", labelKey: "token:Expires in"},
    {type: "text", name: "scope", labelKey: "provider:Scope"},
    {type: "text", name: "tokenType", labelKey: "token:Token type"},
    {
      // the panel carries its own two headings, so the row label stays empty
      type: "custom",
      name: "accessToken",
      label: "",
      block: true,
      render: (ctx, update) => <TokenPanel token={ctx.record} update={update} />,
    },
    {type: "textarea", name: "refreshToken", labelKey: "token:Refresh token", rows: 6},
  ];

  return (
    <SimpleEditPage
      titleKey="token:Edit Token"
      backTo="/tokens"
      deps={[tokenName]}
      fields={fields}
      fetch={() => TokenBackend.getToken("admin", tokenName)}
      add={(record) => TokenBackend.addToken(record)}
      update={(record) => TokenBackend.updateToken("admin", tokenName, record)}
      editUrl={(record) => `/tokens/${record.name}`}
    />
  );
}
