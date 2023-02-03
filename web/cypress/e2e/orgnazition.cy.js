describe('Test Orgnazition', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test org", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/organizations");
        cy.url().should("eq", "http://localhost:7001/organizations");
        cy.visit("http://localhost:7001/organizations/built-in");
        cy.url().should("eq", "http://localhost:7001/organizations/built-in");
        cy.visit("http://localhost:7001/organizations/built-in/users");
        cy.url().should("eq", "http://localhost:7001/organizations/built-in/users");
    });
})
