// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })
Cypress.Commands.add('login', ()=>{
    cy.request({
        method: "POST",
        url: "http://localhost:7001/api/login",
        body: {
          "application": "app-built-in",
          "organization": "built-in",
          "username": "admin",
          "password": "123",
          "autoSignin": true,
          "type": "login",
          "phonePrefix": "86",
        },
      }).then((Response) => {
        expect(Response).property("body").property("status").to.equal("ok");
      });
})
