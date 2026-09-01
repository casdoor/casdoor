describe("Test sessions", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test sessions", () => {
    cy.visitListPage("/sessions");
  });
});
