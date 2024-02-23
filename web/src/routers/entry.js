import {lazy} from "react";

// Define routes array
const entryRoutes = [
  {
    path: "/signup",
    component: lazy(() => import("../auth/SignupPage")),
    auth: true,
    unAuthRedirect: "home",
  },
  {
    path: "/signup/:applicationName",
    component: lazy(() => import("../auth/SignupPage")),
    auth: true,
    unAuthRedirect: "home",
  },
  {
    path: "/login",
    component: lazy(() => import("../auth/SelfLoginPage")),
    auth: true,
    unAuthRedirect: "home",
  },
  {
    path: "/login/:owner",
    component: lazy(() => import("../auth/SelfLoginPage")),
    auth: true,
    unAuthRedirect: "home",
  },
  {
    path: "/signup/oauth/authorize",
    component: lazy(() => import("../auth/SignupPage")),
    auth: true,
    unAuthRedirect: "home",
  },
  {
    path: "/login/oauth/authorize",
    component: lazy(() => import("../auth/LoginPage")),
    auth: true,
    unAuthRedirect: "home",
    props: {
      type: "code",
      mode: "signin",
    },
  },
  {
    path: "/login/saml/authorize/:owner/:applicationName",
    component: lazy(() => import("../auth/LoginPage")),
    auth: true,
    unAuthRedirect: "home",
    props: {
      type: "saml",
      mode: "signin",
    },
  },
  {
    path: "/forget",
    component: lazy(() => import("../auth/ForgetPage")),
    auth: true,
    unAuthRedirect: "home",
  },
  {
    path: "/forget/:applicationName",
    component: lazy(() => import("../auth/ForgetPage")),
    auth: true,
    unAuthRedirect: "home",
  },
  {
    path: "/prompt",
    component: lazy(() => import("../auth/PromptPage")),
    auth: true,
    unAuthRedirect: "login",
  },
  {
    path: "/prompt/:applicationName",
    component: lazy(() => import("../auth/PromptPage")),
    auth: true,
    unAuthRedirect: "login",
  },
  {
    path: "/result",
    component: lazy(() => import("../auth/ResultPage")),
    auth: true,
    unAuthRedirect: "home",
  },
  {
    path: "/result/:applicationName",
    component: lazy(() => import("../auth/ResultPage")),
    auth: true,
    unAuthRedirect: "home",
  },
  {
    path: "/cas/:owner/:casApplicationName/logout",
    component: lazy(() => import("../auth/CasLogout")),
    auth: true,
    unAuthRedirect: "home",
  },
  {
    path: "/cas/:owner/:casApplicationName/login",
    component: lazy(() => import("../auth/LoginPage")),
    props: {
      type: "cas",
      mode: "signin",
    },
  },
  {
    path: "/buy-plan/:owner/:pricingName",
    component: lazy(() => import("../ProductBuyPage")),
  },
  {
    path: "/buy-plan/:owner/:pricingName/result",
    component: lazy(() => import("../ProductBuyPage")),
  },
  {
    path: "/qrcode/:owner/:paymentName",
    component: lazy(() => import("../QrCodePage")),
  },
];

export default entryRoutes;
