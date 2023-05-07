describe('Test payments', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test payments", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/payments");
        cy.url().should("eq", "http://localhost:7001/payments");
    });
})
