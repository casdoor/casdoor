const { defineConfig } = require("cypress");

module.exports = defineConfig({
  e2e: {
    "retries": {
      "runMode": 3,
      "openMode": 0
    }
  },
});
Cypress.config('defaultCommandTimeout', 10000);
Cypress.config('pageLoadTimeout', 30000); 
