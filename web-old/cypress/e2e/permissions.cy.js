describe('Test permissions', () => {
    beforeEach(()=>{
        cy.login();
    })
    it("test permissions", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/permissions");
        cy.url().should("eq", "http://localhost:7001/permissions");
        cy.visit("http://localhost:7001/permissions/built-in/permission-built-in");
        cy.url().should("eq", "http://localhost:7001/permissions/built-in/permission-built-in");
    });
})
