const { defineConfig } = require("cypress");

module.exports = defineConfig({
  e2e: {
    baseUrl: "http://localhost:7001",
    "retries": {
      "runMode": 2,
      "openMode": 0
    }
  },
});
