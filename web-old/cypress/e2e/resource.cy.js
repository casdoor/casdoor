describe('Test resource', () => {
    beforeEach(()=>{
        cy.login();
    })
    it("test resource", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/resources");
        cy.url().should("eq", "http://localhost:7001/resources");
    });
})
