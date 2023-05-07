describe('Test payments', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    const selector = {	
        add: ".ant-table-title > div > .ant-btn"	
      };
    it("test payments", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/payments");
        cy.url().should("eq", "http://localhost:7001/payments");
        cy.get(selector.add,{timeout:10000}).click();	
        cy.url().should("include","http://localhost:7001/payments/")
    });
})
