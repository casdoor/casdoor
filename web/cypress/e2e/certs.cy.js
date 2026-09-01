describe("Test certs", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test certs", () => {
    cy.visitListPage("/certs");
    cy.openFirstRow();
  });
});
