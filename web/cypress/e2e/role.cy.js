describe("Test roles", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test roles", () => {
    cy.visitListPage("/roles");
    cy.get("#add-button", {timeout: 20000}).click();
    cy.location("pathname").should("not.eq", "/roles");
    cy.assertEditPageLoaded();
  });
});
