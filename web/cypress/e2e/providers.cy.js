describe('Test providers', () => {
    beforeEach(()=>{
        cy.visit("http://localhost:7001");
        cy.login();
    })
    it("test providers", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/providers");
        cy.url().should("eq", "http://localhost:7001/providers");
        cy.visit("http://localhost:7001/providers/admin/provider_captcha_default");
        cy.url().should("eq", "http://localhost:7001/providers/admin/provider_captcha_default");
    });
})