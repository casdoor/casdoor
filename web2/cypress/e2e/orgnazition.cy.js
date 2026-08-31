describe("Test organization", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test organization", () => {
    cy.visitListPage("/organizations");
    cy.openFirstRow();
    cy.visitPath("/organizations/built-in");
    cy.assertEditPageLoaded();
    // the org-scoped user list, a different route from /users
    cy.visitListPage("/organizations/built-in/users");
  });
});
