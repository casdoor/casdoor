import * as React from "react";
import i18next from "i18next";
import dayjs from "dayjs";
import {Send, User} from "lucide-react";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Textarea} from "@/components/ui/textarea";
import * as TicketBackend from "@/backend/TicketBackend";
import * as Setting from "@/lib/setting";

/**
 * The conversation on a support ticket, ported from `renderMessages` in
 * `web/src/TicketEditPage.js`. Messages are appended through
 * `/api/add-ticket-message` against the saved ticket, so the thread is only
 * offered once the ticket exists.
 */
export function TicketMessages({
  organizationName,
  ticketName,
  messages,
  account,
  onSent,
}: {
  organizationName: string;
  ticketName: string;
  messages: any[];
  account: any;
  onSent: () => void;
}) {
  const [text, setText] = React.useState("");
  const [sending, setSending] = React.useState(false);

  const send = () => {
    if (!text.trim()) {
      Setting.showMessage("error", i18next.t("ticket:Please enter a message"));
      return;
    }

    setSending(true);
    TicketBackend.addTicketMessage(organizationName, ticketName, {
      author: account?.name,
      text,
      timestamp: dayjs().format(),
      isAdmin: Setting.isAdminUser(account),
    })
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully sent"));
          setText("");
          onSent();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to send")}: ${res.msg}`);
        }
      })
      .catch((error) => Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`))
      .finally(() => setSending(false));
  };

  return (
    <div className="space-y-3">
      <div className="divide-y rounded-lg border">
        {(messages ?? []).length === 0 ? (
          <div className="p-4 text-center text-sm text-muted-foreground">{i18next.t("general:No data")}</div>
        ) : (
          (messages ?? []).map((message: any, index: number) => (
            <div key={index} className="flex gap-3 p-3">
              <div
                className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-white ${
                  message.isAdmin ? "bg-primary" : "bg-success"
                }`}
              >
                <User className="h-4 w-4" />
              </div>
              <div className="min-w-0 space-y-1">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-sm font-medium">{message.author}</span>
                  {message.isAdmin ? <Badge variant="secondary">{i18next.t("general:Admin")}</Badge> : null}
                  <span className="text-xs text-muted-foreground">{Setting.getFormattedDate(message.timestamp)}</span>
                </div>
                <p className="whitespace-pre-wrap break-words text-sm">{message.text}</p>
              </div>
            </div>
          ))
        )}
      </div>

      <div className="flex gap-2">
        <Textarea
          rows={3}
          value={text}
          placeholder={i18next.t("ticket:Type your message here...")}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
              e.preventDefault();
              send();
            }
          }}
        />
        <Button className="h-auto shrink-0" loading={sending} onClick={send}>
          <Send />
          {i18next.t("general:Send")}
        </Button>
      </div>
      <p className="text-xs text-muted-foreground">{i18next.t("ticket:Press Ctrl+Enter to send")}</p>
    </div>
  );
}
