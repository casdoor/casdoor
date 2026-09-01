describe('Test sysinfo', () => {
    beforeEach(()=>{
        cy.login();
    })
    it("test sysinfo", () => {
        cy.visit("http://localhost:7001");
        cy.visit("http://localhost:7001/sysinfo");
        cy.url().should("eq", "http://localhost:7001/sysinfo");
    });
})
