describe("Test syncers", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test syncers", () => {
    cy.visitListPage("/syncers");
    cy.get("#add-button", {timeout: 20000}).click();
    cy.location("pathname").should("not.eq", "/syncers");
    cy.assertEditPageLoaded();
  });
});
