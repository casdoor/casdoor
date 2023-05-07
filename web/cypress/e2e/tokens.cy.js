describe('Test tokens', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test records", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/tokens");
        cy.url().should("eq", "http://localhost:7001/tokens");
    });
})
