describe('Test models', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test org", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/models");
        cy.url().should("eq", "http://localhost:7001/models");
        cy.visit("http://localhost:7001/models/built-in/model-built-in");
        cy.url().should("eq", "http://localhost:7001/models/built-in/model-built-in");
    });
})