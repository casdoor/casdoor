describe("Test adapters", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test adapters", () => {
    cy.visitListPage("/adapters");
    cy.get("#add-button", {timeout: 20000}).click();
    cy.location("pathname").should("not.eq", "/adapters");
    cy.assertEditPageLoaded();
  });
});
