describe("Test models", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test models", () => {
    cy.visitListPage("/models");
    cy.openFirstRow();
  });
});
