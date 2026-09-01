describe("Test roles", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test roles", () => {
    cy.visitListPage("/roles");
    cy.openFirstRow();
  });
});
