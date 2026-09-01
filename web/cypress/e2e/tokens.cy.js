describe("Test tokens", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test tokens", () => {
    cy.visitListPage("/tokens");
    cy.get("#add-button", {timeout: 20000}).click();
    cy.location("pathname").should("not.eq", "/tokens");
    cy.assertEditPageLoaded();
  });
});
