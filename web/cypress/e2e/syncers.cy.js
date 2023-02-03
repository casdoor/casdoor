describe('Test syncers', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    const selector = {
        add: ".ant-table-title > div > .ant-btn"
      };
    it("test syncers", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/syncers");
        cy.url().should("eq", "http://localhost:7001/syncers");
        cy.get(selector.add).click();
        cy.url().should("include","http://localhost:7001/syncers/")
    });
})