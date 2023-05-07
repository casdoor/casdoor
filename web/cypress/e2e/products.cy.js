describe('Test products', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test products", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/products");
        cy.url().should("eq", "http://localhost:7001/products");
    });
})
