// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

describe("Syncer organization filtering", () => {
  function openSyncers(owner, selection) {
    cy.intercept("GET", "**/api/get-account*", {
      status: "ok",
      data: {owner, name: "admin", displayName: "Admin", isAdmin: true},
      data2: {name: owner, displayName: owner, enableTour: false},
    });
    cy.intercept("GET", "**/api/get-organization-names*", {
      status: "ok",
      data: [
        {name: "built-in", displayName: "Built-in"},
        {name: "tenant", displayName: "Tenant"},
      ],
    });
    cy.intercept("GET", "**/api/get-syncers?*", (req) => {
      const organization = new URL(req.url).searchParams.get("organization");
      const syncers = [
        {owner: "admin", name: "tenant-syncer", organization: "tenant", type: "Database"},
      ].filter((syncer) => !organization || syncer.organization === organization);
      req.reply({status: "ok", data: syncers, data2: syncers.length});
    }).as("syncers");
    cy.visit("/syncers", {
      onBeforeLoad(win) {
        win.localStorage.setItem("organization", selection);
        win.localStorage.setItem("language", "en");
        win.localStorage.setItem("isTourVisible", "false");
      },
    });
  }

  function expectOrganization(organization) {
    cy.get("@syncers.all").should((requests) => {
      expect(requests).not.to.be.empty;
      const {request} = requests[requests.length - 1];
      expect(new URL(request.url).searchParams.get("organization")).to.eq(organization);
    });
  }

  function selectOrganization(label) {
    cy.get("header [role=combobox]").first().click();
    cy.contains("[role=option]", new RegExp(`^${label}$`)).click();
  }

  it("loads all syncers and refreshes when switching organizations", () => {
    openSyncers("built-in", "All");
    expectOrganization("");
    cy.contains("tbody", "tenant-syncer").should("be.visible");

    selectOrganization("Built-in");
    expectOrganization("built-in");
    cy.contains("tbody", "tenant-syncer").should("not.exist");

    selectOrganization("All");
    expectOrganization("");
    cy.contains("tbody", "tenant-syncer").should("be.visible");

    selectOrganization("Tenant");
    expectOrganization("tenant");
    cy.contains("tbody", "tenant-syncer").should("be.visible");
  });

  it("keeps tenant administrators scoped to their own organization", () => {
    openSyncers("tenant", "All");
    expectOrganization("tenant");
    cy.contains("tbody", "tenant-syncer").should("be.visible");
  });
});
