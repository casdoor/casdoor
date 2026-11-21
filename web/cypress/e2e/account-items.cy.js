// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

describe("Account items with admin-only identifiers", () => {
  const user = {id: "profile-id", owner: "test-org", name: "member", displayName: "Member"};
  const organization = {
    name: user.owner,
    accountItems: [
      {name: "ID", visible: true, viewRule: "Admin", modifyRule: "Immutable"},
      {name: "Name", visible: true, viewRule: "Admin", modifyRule: "Admin"},
      {name: "Display name", visible: true, viewRule: "Self", modifyRule: "Self"},
    ],
  };
  const displayName = '[data-label-key="general:Display name"] input';
  const id = '[data-label-key="general:ID"]';
  const name = '[data-label-key="general:Name"]';

  const openProfile = (path, account, profile) => {
    cy.intercept("GET", "**/api/**", {status: "ok", data: []});
    cy.intercept({method: "GET", pathname: "/api/get-account"}, {status: "ok", data: account, data2: organization});
    cy.intercept({method: "GET", pathname: "/api/get-user"}, {status: "ok", data: profile});
    cy.intercept({method: "GET", pathname: "/api/get-user-application"}, {status: "ok", data: {organizationObj: organization}});
    cy.visit(path, {onBeforeLoad: (win) => {
      win.localStorage.setItem("language", "en");
      win.localStorage.setItem("isTourVisible", "false");
    }});
  };

  it("saves the signed-in user's profile without showing its identifiers", () => {
    openProfile("/account", user, {...user, id: "", name: ""});
    cy.intercept({method: "POST", pathname: "/api/update-user"}, {status: "ok"}).as("saveProfile");

    cy.get(displayName).clear().type("Updated member");
    cy.get(id).should("not.exist");
    cy.get(name).should("not.exist");
    cy.contains("button", /^Save$/).click();
    cy.wait("@saveProfile").then(({request}) => {
      expect(new URL(request.url).searchParams.get("id")).to.eq("test-org/member");
      expect(request.body).to.include({owner: user.owner, name: user.name, id: "", displayName: "Updated member"});
    });
    cy.contains("Successfully saved").should("be.visible");
    cy.location("pathname").should("eq", "/account");
    cy.get(name).should("not.exist");
  });

  it("still shows the identifiers to an organization admin", () => {
    openProfile("/users/test-org/member", {...user, name: "admin", isAdmin: true}, user);

    cy.get(id).find("input").should("have.value", user.id);
    cy.get(name).find("input").should("have.value", user.name);
  });

  it("does not show self-only fields on another user's profile", () => {
    openProfile("/users/test-org/other", user, {...user, id: "", name: "", displayName: ""});

    cy.contains("h1", "Edit User").should("be.visible");
    cy.get(displayName).should("not.exist");
    cy.get(id).should("not.exist");
    cy.get(name).should("not.exist");
  });
});
