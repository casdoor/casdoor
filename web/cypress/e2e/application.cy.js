describe("Test application", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test application", () => {
    cy.visitListPage("/applications");
    cy.openFirstRow();
  });
});
