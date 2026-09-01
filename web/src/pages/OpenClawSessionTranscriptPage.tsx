import * as React from "react";
import i18next from "i18next";
import {ArrowLeft} from "lucide-react";
import {useNavigate, useParams} from "react-router-dom";
import {Alert} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {Card, CardContent} from "@/components/ui/card";
import {CodeEditor} from "@/components/common/CodeEditor";
import {DescriptionList, type DescriptionItem} from "@/components/common/DescriptionList";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
import * as EntryBackend from "@/backend/EntryBackend";
import * as Setting from "@/lib/setting";

/** Port of `web/src/OpenClawSessionTranscriptPage.js`: the raw JSONL of a session. */
export default function OpenClawSessionTranscriptPage() {
  const {organizationName = "", entryName = ""} = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState("");
  const [transcript, setTranscript] = React.useState<any>(null);

  React.useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError("");
    setTranscript(null);

    EntryBackend.getOpenClawSessionTranscript(organizationName, entryName)
      .then((res: any) => {
        if (cancelled) {
          return;
        }
        setLoading(false);
        if (res.status === "ok" && res.data) {
          setError("");
          setTranscript(res.data);
        } else {
          setError(`${i18next.t("general:Failed to load")}: ${res.msg}`);
          setTranscript(null);
        }
      })
      .catch((err: any) => {
        if (cancelled) {
          return;
        }
        setLoading(false);
        setError(`${i18next.t("general:Failed to load")}: ${err?.message || String(err)}`);
        setTranscript(null);
      });

    return () => {
      cancelled = true;
    };
  }, [organizationName, entryName]);

  const renderContent = () => {
    if (loading) {
      return <Loading />;
    }
    if (error) {
      return <Alert variant="warning">{error}</Alert>;
    }
    if (!transcript) {
      return null;
    }

    const items: DescriptionItem[] = [
      {label: i18next.t("resource:File name"), children: transcript.fileName || "-"},
      {label: i18next.t("resource:File size"), children: Setting.getFriendlyFileSize(transcript.fileSize || 0)},
      {label: i18next.t("entry:Loaded size"), children: Setting.getFriendlyFileSize(transcript.loadedSize || 0)},
    ];

    return (
      <div className="grid gap-3">
        <DescriptionList items={items} columns={3} />
        {transcript.truncated ? (
          <Alert variant="warning">{i18next.t("entry:Transcript truncated")}</Alert>
        ) : null}
        <CodeEditor
          value={transcript.content || ""}
          readOnly
          height={Math.max(360, window.innerHeight - 360)}
          onChange={() => {}}
        />
      </div>
    );
  };

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("entry:Raw JSONL")}
        actions={
          <Button
            variant="outline"
            onClick={() => navigate(`/entries/${organizationName}/${encodeURIComponent(entryName)}`)}
          >
            <ArrowLeft className="mr-1 h-4 w-4" />
            {i18next.t("entry:Back to session")}
          </Button>
        }
      />
      <Card>
        <CardContent className="pt-6">{renderContent()}</CardContent>
      </Card>
    </div>
  );
}
