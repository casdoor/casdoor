describe("Test products", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test products", () => {
    cy.visitListPage("/products");
    cy.get("#add-button", {timeout: 20000}).click();
    cy.location("pathname").should("not.eq", "/products");
    cy.assertEditPageLoaded();
  });
});
