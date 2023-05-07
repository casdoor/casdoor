describe('Test adapter', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    const selector = {	
        add: ".ant-table-title > div > .ant-btn"	
      };
    it("test adapter", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/adapters");
        cy.url().should("eq", "http://localhost:7001/adapters");
        cy.get(selector.add,{timeout:10000}).click();	
        cy.url().should("include","http://localhost:7001/adapters/built-in/")
    });
})
