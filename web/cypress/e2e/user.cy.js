describe("Test User", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test user", () => {
    cy.visitListPage("/users");
    cy.openFirstRow();
  });
});
