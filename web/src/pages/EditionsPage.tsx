import * as React from "react";
import i18next from "i18next";
import {ArrowUpRight, Check, ChevronRight, ExternalLink, Gem, Lock, PlayCircle, Sparkles} from "lucide-react";
import {useLocation, useNavigate, useSearchParams} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardHeader} from "@/components/ui/card";
import {PageHeader} from "@/components/crud/PageHeader";
import {getEditionText, PricingUrl, useEdition} from "@/components/common/EditionNotice";
import {getEditionDemoUrl, startEditionTour} from "@/components/common/EditionTour";
import {
  type EditionFeature,
  type FeatureStop,
  getEditionFeature,
  getEditionFeatures,
  isEditionCovered,
  type PaidEdition,
} from "@/lib/edition-features";
import {cn} from "@/lib/utils";

const objectLabels: Record<string, string> = {
  organization: "general:Organizations",
  application: "general:Applications",
  cert: "general:Certs",
  permission: "general:Permissions",
  model: "general:Models",
  server: "general:MCP Servers",
};

const pathLabels: Record<string, string> = {
  "/": "general:Dashboard",
  "/compliance": "general:Compliance",
  "/marketplace": "general:Marketplace",
  "/syncers": "general:Syncers",
  "/providers": "application:Providers",
};

function getStopPath(stop: FeatureStop): string[] {
  const parts: string[] = [];
  if (stop.object) {
    parts.push(i18next.t(objectLabels[stop.object]));
  } else if (stop.path) {
    parts.push(i18next.t(pathLabels[stop.path.split("?")[0]] ?? stop.path));
  }
  if (stop.tabLabel) {
    parts.push(stop.tabLabel.includes(":") ? i18next.t(stop.tabLabel) : stop.tabLabel);
  }
  if (stop.label && stop.label !== stop.tabLabel) {
    parts.push(i18next.t(stop.label));
  }
  return parts;
}

function EditionMark({edition, className}: {edition: PaidEdition; className?: string}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-semibold",
        edition === "Enterprise" ? "bg-[#16294f] text-[#e9cf8f]" : "bg-indigo-600/10 text-indigo-700 dark:bg-indigo-400/15 dark:text-indigo-300",
        className,
      )}
    >
      {edition === "Enterprise" ? <Gem className="size-3" /> : <Sparkles className="size-3" />}
      {getEditionText(edition)}
    </span>
  );
}

function FeatureRow({feature, covered, highlighted, onTour}: {
  feature: EditionFeature;
  covered: boolean;
  highlighted: boolean;
  onTour: () => void;
}) {
  return (
    <li
      id={feature.id}
      className={cn(
        "flex scroll-mt-24 flex-col gap-2 rounded-lg border bg-background p-3 transition-colors sm:flex-row sm:items-start",
        highlighted && "border-primary ring-2 ring-primary/30",
      )}
    >
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-2">
          <span className="font-medium">{i18next.t(feature.name)}</span>
          {covered ? (
            <Check className="size-4 text-success" aria-label={i18next.t("edition:Included in your license")} />
          ) : (
            <Lock className="size-3.5 text-muted-foreground" aria-label={i18next.t("edition:Not in your license")} />
          )}
        </div>
        <p className="mt-0.5 text-sm text-muted-foreground">{i18next.t(feature.description)}</p>
        {feature.stop ? (
          <p className="mt-1.5 flex flex-wrap items-center gap-1 text-xs text-muted-foreground">
            {getStopPath(feature.stop).map((part, index) => (
              <React.Fragment key={index}>
                {index > 0 ? <ChevronRight className="size-3" /> : null}
                <span>{part}</span>
              </React.Fragment>
            ))}
          </p>
        ) : feature.config ? (
          <p className="mt-1.5 text-xs text-muted-foreground">
            {i18next.t("edition:Set in conf/app.conf")}: <code className="rounded bg-muted px-1 py-0.5">{feature.config}</code>
          </p>
        ) : null}
      </div>
      {feature.stop ? (
        <Button variant="outline" size="sm" className="shrink-0" onClick={onTour}>
          <PlayCircle className="size-4" />
          {i18next.t("edition:Show me")}
        </Button>
      ) : null}
    </li>
  );
}

function EditionSection({edition, current, highlighted}: {edition: PaidEdition; current: string | null; highlighted: string}) {
  const features = getEditionFeatures(edition);
  const tourIds = features.filter((feature) => feature.stop).map((feature) => feature.id);
  const covered = current === null || isEditionCovered(current, edition);
  const demoUrl = covered ? undefined : getEditionDemoUrl(edition, `/editions?tour=${edition.toLowerCase()}`);

  const categories: string[] = [];
  for (const feature of features) {
    if (!categories.includes(feature.category)) {
      categories.push(feature.category);
    }
  }

  return (
    <Card
      id={edition.toLowerCase()}
      className={cn(
        "scroll-mt-20 overflow-hidden",
        edition === "Enterprise" && "border-[#c9a45c]/40",
      )}
    >
      <CardHeader
        className={cn(
          "gap-3 border-b",
          edition === "Enterprise" ? "bg-[#0e1a33] text-white" : "bg-gradient-to-br from-indigo-600/[0.08] via-violet-500/[0.05] to-transparent",
        )}
      >
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center">
          <div className="min-w-0 flex-1 space-y-1.5">
            <div className="flex flex-wrap items-center gap-2">
              <EditionMark edition={edition} className={edition === "Enterprise" ? "bg-[#c9a45c]/15" : undefined} />
              {covered ? (
                <Badge variant="secondary" className="gap-1">
                  <Check className="size-3" />
                  {i18next.t("edition:Included in your license")}
                </Badge>
              ) : null}
              <span className={cn("text-xs", edition === "Enterprise" ? "text-white/60" : "text-muted-foreground")}>
                {i18next.t("edition:%s features").replace("%s", String(features.length))}
              </span>
            </div>
            <p className={cn("max-w-3xl text-sm", edition === "Enterprise" ? "text-white/75" : "text-muted-foreground")}>
              {i18next.t(edition === "Enterprise" ? "edition:Enterprise - Summary" : "edition:Commercial - Summary")}
            </p>
          </div>
          <div className="flex shrink-0 flex-wrap gap-2">
            <Button
              onClick={() => startEditionTour(tourIds)}
              className={edition === "Enterprise" ? "bg-[#c9a45c] text-[#0e1a33] hover:bg-[#e0bd6d]" : undefined}
            >
              <PlayCircle className="size-4" />
              {i18next.t("edition:Start the tour")}
            </Button>
            {demoUrl ? (
              <Button asChild variant="outline" className={edition === "Enterprise" ? "border-white/30 bg-transparent text-white hover:bg-white/10 hover:text-white" : undefined}>
                <a href={demoUrl}>
                  {i18next.t("edition:Try it on the %s demo site").replace("%s", getEditionText(edition))}
                  <ExternalLink className="size-4" />
                </a>
              </Button>
            ) : null}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-6 pt-6">
        {categories.map((category) => (
          <section key={category} className="space-y-2">
            <h3 className="text-sm font-semibold text-muted-foreground">{i18next.t(category)}</h3>
            <ul className="grid gap-2 xl:grid-cols-2">
              {features.filter((feature) => feature.category === category).map((feature) => (
                <FeatureRow
                  key={feature.id}
                  feature={feature}
                  covered={covered}
                  highlighted={highlighted === feature.id}
                  onTour={() => startEditionTour(tourIds, tourIds.indexOf(feature.id))}
                />
              ))}
            </ul>
          </section>
        ))}
      </CardContent>
    </Card>
  );
}

/** What the commercial and enterprise editions add, where each feature is in the console, and a tour through them. */
export default function EditionsPage() {
  const current = useEdition();
  const location = useLocation();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const highlighted = location.hash.replace(/^#/, "");

  // "?tour=commercial", "?tour=enterprise" or "?tour=<feature>" starts the tour, e.g. from another edition's demo site
  React.useEffect(() => {
    const tour = searchParams.get("tour");
    if (!tour) {
      return;
    }
    const edition = (["Commercial", "Enterprise"] as PaidEdition[]).find((value) => value.toLowerCase() === tour.toLowerCase());
    const feature = getEditionFeature(tour);
    const target = edition ?? feature?.edition;
    navigate("/editions", {replace: true});
    if (target) {
      const ids = getEditionFeatures(target).filter((item) => item.stop).map((item) => item.id);
      startEditionTour(ids, feature ? ids.indexOf(feature.id) : 0);
    }
  }, [searchParams, navigate]);

  React.useEffect(() => {
    if (highlighted) {
      document.getElementById(highlighted)?.scrollIntoView({block: "center"});
    }
  }, [highlighted]);

  return (
    <div className="space-y-6">
      <PageHeader
        title={i18next.t("edition:Editions and features")}
        description={i18next.t("edition:What the commercial and enterprise editions add, and where each feature is in the console")}
        actions={
          <>
            {current ? (
              <span className="text-sm text-muted-foreground">
                {i18next.t("edition:Your edition")}: <span className="font-medium text-foreground">{getEditionText(current)}</span>
              </span>
            ) : null}
            <Button asChild variant="outline">
              <a href={PricingUrl} target="_blank" rel="noreferrer">
                {i18next.t("general:View editions and pricing")}
                <ArrowUpRight className="size-4" />
              </a>
            </Button>
          </>
        }
      />
      <EditionSection edition="Commercial" current={current} highlighted={highlighted} />
      <p className="text-sm text-muted-foreground">{i18next.t("edition:The enterprise edition includes every commercial feature, and adds")}</p>
      <EditionSection edition="Enterprise" current={current} highlighted={highlighted} />
    </div>
  );
}
