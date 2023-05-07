describe('Test webhooks', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test webhooks", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/webhooks");
        cy.url().should("eq", "http://localhost:7001/webhooks");
    });
})
