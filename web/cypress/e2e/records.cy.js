describe("Test records", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test records", () => {
    cy.visitListPage("/records");
  });
});
