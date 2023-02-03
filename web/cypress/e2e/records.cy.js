describe('Test records', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test records", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/records");
        cy.url().should("eq", "http://localhost:7001/records");
    });
})