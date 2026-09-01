describe("Console breadcrumbs", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("shows Home > resource on a list page", () => {
    cy.visitListPage("/organizations");
    cy.get("nav[aria-label=Breadcrumb]").within(() => {
      cy.contains("a", "Home").should("have.attr", "href", "/");
      cy.contains("Organizations").should("be.visible");
    });
  });

  it("links the resource back to its list on an edit page", () => {
    cy.visitListPage("/organizations");
    cy.openFirstRow();
    cy.get("nav[aria-label=Breadcrumb]").contains("a", "Organizations").click();
    cy.location("pathname").should("eq", "/organizations");
  });

  it("shows nothing on routes it does not know", () => {
    cy.visitPath("/");
    cy.get("nav[aria-label=Breadcrumb]").should("not.exist");
  });
});
