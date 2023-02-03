describe('Test certs', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test certs", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/certs");
        cy.url().should("eq", "http://localhost:7001/certs");
        cy.visit("http://localhost:7001/certs/cert-built-in");
        cy.url().should("eq", "http://localhost:7001/certs/cert-built-in");
    });
})