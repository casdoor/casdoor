describe("Test payments", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test payments", () => {
    cy.visitListPage("/payments");
    cy.get("#add-button", {timeout: 20000}).click();
    cy.location("pathname").should("not.eq", "/payments");
    cy.assertEditPageLoaded();
  });
});
