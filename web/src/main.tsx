// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React from "react";
import {createRoot} from "react-dom/client";
import {BrowserRouter} from "react-router-dom";
import "./index.css";
import "./i18n";
import "./lib/fetch-filter";
import App from "./App";
import ErrorBoundary from "./components/common/ErrorBoundary";
import {ThemeProvider} from "./hooks/use-theme";
import {AccountProvider} from "./hooks/use-account";
import {useLanguageVersion} from "./hooks/use-language";
import {TooltipProvider} from "./components/ui/tooltip";
import {Toaster} from "./components/ui/sonner";

// re-rendered from the root on a language change, so every screen picks up the new
// bundle right away instead of waiting for a reload
function Root() {
  useLanguageVersion();

  return (
    <ErrorBoundary>
      <ThemeProvider>
        <TooltipProvider delayDuration={200}>
          <BrowserRouter>
            <AccountProvider>
              <App />
              <Toaster />
            </AccountProvider>
          </BrowserRouter>
        </TooltipProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

const container = document.getElementById("root")!;
createRoot(container).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>,
);
