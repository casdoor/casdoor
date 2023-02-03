describe('Test sessions', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test sessions", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/sessions");
        cy.url().should("eq", "http://localhost:7001/sessions");
    });
})