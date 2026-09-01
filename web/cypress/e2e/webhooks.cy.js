describe("Test webhooks", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test webhooks", () => {
    cy.visitListPage("/webhooks");
    cy.get("#add-button", {timeout: 20000}).click();
    cy.location("pathname").should("not.eq", "/webhooks");
    cy.assertEditPageLoaded();
  });
});
