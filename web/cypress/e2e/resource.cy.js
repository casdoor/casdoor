describe("Test resources", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test resources", () => {
    cy.visitListPage("/resources");
  });
});
