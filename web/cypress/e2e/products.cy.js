describe('Test products', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    const selector = {
        add: ".ant-table-title > div > .ant-btn > span"
      };
    it("test products", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/products");
        cy.url().should("eq", "http://localhost:7001/products");
        cy.get(selector.add).click();
        cy.url().should("include","http://localhost:7001/products/")
    });
})
