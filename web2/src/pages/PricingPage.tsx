import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams, useSearchParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {Loading} from "@/components/common/Loading";
import {useAccount} from "@/hooks/use-account";
import * as PlanBackend from "@/backend/PlanBackend";
import * as PricingBackend from "@/backend/PricingBackend";
import * as Setting from "@/lib/setting";

function PlanCard({plan, link}: {plan: any; link: string}) {
  const navigate = useNavigate();

  return (
    <Card className="flex w-full min-w-[280px] max-w-[340px] flex-col">
      <CardHeader>
        <CardTitle className="text-xl">{plan.displayName}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-4">
        <div>
          <span className="text-4xl font-bold">
            {Setting.getCurrencySymbol(plan.currency)} {plan.price}
          </span>
          <span className="ml-1 text-base font-semibold text-muted-foreground">
            {plan.period === "Yearly" ? i18next.t("plan:per year") : i18next.t("plan:per month")}
          </span>
        </div>
        <p className="min-h-[60px] text-muted-foreground">{plan.description}</p>
        <Button className="mt-auto h-12 w-full" onClick={() => navigate(link)}>
          {i18next.t("pricing:Getting started")}
        </Button>
      </CardContent>
    </Card>
  );
}

/**
 * The public plan picker at /select-plan/:owner/:pricingName. Ported from
 * web/src/pricing/PricingPage.js — a signed-in visitor goes to /buy-plan, an
 * anonymous one to the application's sign-up page carrying plan and pricing.
 */
export default function PricingPage() {
  const params = useParams();
  const [search] = useSearchParams();
  const navigate = useNavigate();
  const {account} = useAccount();

  const owner = params.owner ?? "";
  const pricingName = params.pricingName ?? "";
  const userName = search.get("user");

  const [pricing, setPricing] = React.useState<any>(null);
  const [plans, setPlans] = React.useState<any[] | null>(null);
  const [periods, setPeriods] = React.useState<string[]>([]);
  const [selectedPeriod, setSelectedPeriod] = React.useState<string | undefined>(undefined);

  React.useEffect(() => {
    if (userName) {
      Setting.showMessage("info", i18next.t("pricing:paid-user do not have active subscription or pending subscription, please select a plan to buy"));
    }
    // only on mount, as in the antd page
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  React.useEffect(() => {
    if (!owner || !pricingName) {
      return;
    }
    PricingBackend.getPricing(owner, pricingName).then((res: any) => {
      if (res.status === "error") {
        Setting.showMessage("error", res.msg);
        return;
      }
      setPricing(res.data);
    });
  }, [owner, pricingName]);

  React.useEffect(() => {
    if (pricing === null) {
      return;
    }
    Promise.all((pricing.plans ?? []).map((plan: string) => PlanBackend.getPlan(owner, plan, true)))
      .then((results: any[]) => {
        if (results.some((result) => result.status === "error")) {
          Setting.showMessage("error", i18next.t("general:Failed to get"));
          return;
        }
        const loaded = results.map((result) => result.data);
        const loadedPeriods = [...new Set(loaded.map((plan: any) => plan.period).filter((period: string) => period !== ""))] as string[];
        setPlans(loaded);
        setPeriods(loadedPeriods);
        setSelectedPeriod(loadedPeriods[0]);
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${error}`);
      });
  }, [pricing, owner]);

  if (pricing === null || plans === null) {
    return <Loading className="min-h-screen" />;
  }

  const getUrlByPlan = (planName: string) => {
    if (account) {
      const buyer = userName || account.name;
      return `/buy-plan/${pricing.owner}/${pricing.name}?plan=${planName}${buyer ? `&user=${buyer}` : ""}`;
    }
    return `/signup/${pricing.application}?plan=${planName}&pricing=${pricing.name}`;
  };

  return (
    <div className="mx-auto max-w-6xl px-4 py-12 text-center">
      <h1 className="mb-4 text-5xl font-bold">{pricing.displayName}</h1>
      <p className="text-xl text-muted-foreground">{pricing.description}</p>

      {periods.length > 1 ? (
        <div className="mt-10 flex justify-center gap-2">
          {periods.map((period) => (
            <Button
              key={period}
              size="lg"
              variant={selectedPeriod === period ? "default" : "outline"}
              onClick={() => setSelectedPeriod(period)}
            >
              {period}
            </Button>
          ))}
        </div>
      ) : null}

      <div className="mt-10 flex flex-wrap justify-center gap-6">
        {plans
          .filter((plan: any) => plan.period === selectedPeriod)
          .map((plan: any) => (
            <PlanCard key={plan.name} plan={plan} link={getUrlByPlan(plan.name)} />
          ))}
      </div>

      {pricing.trialDuration > 0 ? (
        <p className="mt-8 italic text-muted-foreground">
          {i18next.t("pricing:Free")} {pricing.trialDuration}-{i18next.t("pricing:days trial available!")}
        </p>
      ) : null}

      <div className="mt-8">
        <Button variant="ghost" onClick={() => navigate("/")}>{i18next.t("general:Back Home")}</Button>
      </div>
    </div>
  );
}
