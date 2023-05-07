describe('Test syncers', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test syncers", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/syncers");
        cy.url().should("eq", "http://localhost:7001/syncers");
    });
})
