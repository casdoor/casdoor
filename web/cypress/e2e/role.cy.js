describe('Test roles', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test role", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/roles");
        cy.url().should("eq", "http://localhost:7001/roles");
    });
})
