describe("Test providers", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test providers", () => {
    cy.visitListPage("/providers");
    cy.openFirstRow();
  });
});
