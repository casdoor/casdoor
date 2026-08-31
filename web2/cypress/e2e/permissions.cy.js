describe("Test permissions", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test permissions", () => {
    cy.visitListPage("/permissions");
    cy.openFirstRow();
  });
});
