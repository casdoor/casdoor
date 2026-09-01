describe("Test sysinfo", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test sysinfo", () => {
    // not a list page, so it only has the heading to assert on
    cy.visitPath("/sysinfo");
    cy.get("h1", {timeout: 20000}).should("be.visible");
  });
});
