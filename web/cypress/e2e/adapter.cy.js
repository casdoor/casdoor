describe('Test adapter', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test adapter", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/adapters");
        cy.url().should("eq", "http://localhost:7001/adapters");
    });
})
